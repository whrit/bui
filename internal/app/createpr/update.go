package createpr

import (
	"context"
	"strings"
	"time"

	"bui/internal/llm"
	"bui/internal/llm/prompts"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// defaultLLMTimeout is the default timeout for LLM generation in seconds.
const defaultLLMTimeout = 60

// Update implements tea.Model. It handles all incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// Window size change
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	// Branch loading messages
	case BranchesLoadedMsg:
		return m.handleBranchesLoaded(msg)

	case BranchesLoadErrorMsg:
		return m.handleBranchesLoadError(msg)

	// Commit loading messages
	case CommitsLoadedMsg:
		return m.handleCommitsLoaded(msg)

	case CommitsLoadErrorMsg:
		return m.handleCommitsLoadError(msg)

	// LLM draft generation messages
	case DraftStartMsg:
		return m.handleDraftStart()

	case DraftTokenMsg:
		return m.handleDraftToken(msg)

	case DraftDoneMsg:
		return m.handleDraftDone()

	case DraftErrorMsg:
		return m.handleDraftError(msg)

	case SkipDraftMsg:
		return m.handleSkipDraft()

	// PR creation messages
	case SubmitPRMsg:
		return m.handleSubmitPR()

	case PRCreatedMsg:
		return m.handlePRCreated(msg)

	case PRCreateErrorMsg:
		return m.handlePRCreateError(msg)

	// Spinner tick
	case spinner.TickMsg:
		if m.step == StepBranchSelect && !m.branchesReady ||
			m.step == StepDraftGeneration ||
			m.step == StepSubmitting {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	// Keyboard input
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	// Pass remaining messages to active components based on step
	switch m.step {
	case StepBranchSelect:
		if m.branchesReady {
			var cmd tea.Cmd
			m.branchList, cmd = m.branchList.Update(msg)
			cmds = append(cmds, cmd)
		}

	case StepEditTitle:
		var cmd tea.Cmd
		m.titleInput, cmd = m.titleInput.Update(msg)
		cmds = append(cmds, cmd)

	case StepEditBody:
		var cmd tea.Cmd
		m.bodyInput, cmd = m.bodyInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// =============================================================================
// Branch Loading Handlers
// =============================================================================

// handleBranchesLoaded processes successfully loaded branches.
func (m Model) handleBranchesLoaded(msg BranchesLoadedMsg) (tea.Model, tea.Cmd) {
	m.branches = msg.Branches
	m.headBranch = msg.CurrentBranch
	m.baseBranch = msg.DefaultBranch
	m.branchesReady = true

	// Update the branch list with items
	m.updateBranchListItems()

	// Load commits between base and head
	return m, loadCommits(m.gitRepo, m.baseBranch, m.headBranch)
}

// handleBranchesLoadError processes branch loading errors.
func (m Model) handleBranchesLoadError(msg BranchesLoadErrorMsg) (tea.Model, tea.Cmd) {
	m.step = StepError
	m.err = msg.Err
	return m, nil
}

// =============================================================================
// Commit Loading Handlers
// =============================================================================

// handleCommitsLoaded processes successfully loaded commits.
func (m Model) handleCommitsLoaded(msg CommitsLoadedMsg) (tea.Model, tea.Cmd) {
	m.commits = msg.Commits
	return m, nil
}

// handleCommitsLoadError processes commit loading errors.
func (m Model) handleCommitsLoadError(msg CommitsLoadErrorMsg) (tea.Model, tea.Cmd) {
	// Don't fail the whole wizard, just proceed without commits
	m.commits = nil
	return m, nil
}

// =============================================================================
// LLM Draft Handlers
// =============================================================================

// handleDraftStart begins the LLM draft generation.
func (m Model) handleDraftStart() (tea.Model, tea.Cmd) {
	if m.llmProvider == nil {
		// No LLM provider, skip to editing
		m.step = StepEditTitle
		m.titleInput.Focus()
		return m, nil
	}

	m.step = StepDraftGeneration
	m.spinner = m.spinner.WithMessage("Generating PR draft...")
	m.streamedText = ""
	m.generatedTitle = ""
	m.generatedBody = ""

	// Determine timeout duration
	timeout := defaultLLMTimeout * time.Second
	if m.cfg.LLM.Timeout > 0 {
		timeout = time.Duration(m.cfg.LLM.Timeout) * time.Second
	}

	// Create context with timeout for cancellation
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	m.llmCtx = ctx
	m.llmCancel = cancel

	// Build prompt for PR creation
	commitLog := m.buildCommitLog()

	system, user, err := prompts.RenderPRCreate(prompts.PRCreateVars{
		Repo:      "", // Could be populated from git repo info
		Base:      m.baseBranch,
		Head:      m.headBranch,
		CommitLog: commitLog,
		DiffStat:  "", // Could be populated from git diff stat
		UserNotes: "",
	})
	if err != nil {
		return m, func() tea.Msg {
			return DraftErrorMsg{Err: err}
		}
	}

	prompt := llm.Prompt{
		System:      system,
		User:        user,
		MaxTokens:   1000,
		Temperature: 0.3,
	}

	// Start streaming
	tokenCh, errCh := m.llmProvider.Stream(ctx, prompt)
	m.tokenCh = tokenCh
	m.errCh = errCh

	// Return commands to start spinner and read first token
	return m, tea.Batch(
		m.spinner.Init(),
		m.readNextToken(),
	)
}

// handleDraftToken processes an incoming LLM token.
func (m Model) handleDraftToken(msg DraftTokenMsg) (tea.Model, tea.Cmd) {
	m.streamedText += msg.Token
	return m, m.readNextToken()
}

// handleDraftDone completes the LLM draft generation.
func (m Model) handleDraftDone() (tea.Model, tea.Cmd) {
	if m.llmCancel != nil {
		m.llmCancel()
		m.llmCancel = nil
	}

	// Parse the streamed text into title and body
	m.parseDraftOutput()

	// Move to title editing
	m.step = StepEditTitle
	m.titleInput.SetValue(m.generatedTitle)
	m.bodyInput.SetValue(m.generatedBody)
	m.title = m.generatedTitle
	m.body = m.generatedBody
	m.titleInput.Focus()

	return m, nil
}

// parseDraftOutput parses the LLM output into title and body.
func (m *Model) parseDraftOutput() {
	text := strings.TrimSpace(m.streamedText)
	if text == "" {
		return
	}

	lines := strings.Split(text, "\n")
	titleFound := false
	bodyStart := 0

	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Check for Title: prefix (case-insensitive with whitespace handling)
		if !titleFound {
			if titleContent, found := extractLabeledContent(trimmedLine, "title"); found {
				m.generatedTitle = strings.Trim(titleContent, "\"'")
				titleFound = true
				continue
			}

			// Check for markdown header format: "# Title" or "## Title"
			if mdContent, found := extractMarkdownHeader(trimmedLine, "title"); found {
				m.generatedTitle = strings.Trim(mdContent, "\"'")
				titleFound = true
				continue
			}

			// If first non-empty line doesn't have Title:, use it as title
			if trimmedLine != "" && i == 0 {
				m.generatedTitle = strings.Trim(trimmedLine, "\"'")
				titleFound = true
				bodyStart = i + 1
				continue
			}
		}

		// Look for Body: or start of body content
		if titleFound {
			if _, found := extractLabeledContent(trimmedLine, "body"); found {
				bodyStart = i + 1
				break
			}
			// Check for "body (markdown):" variant
			if strings.HasPrefix(normalizeLabel(trimmedLine), "body(markdown):") ||
				strings.HasPrefix(normalizeLabel(trimmedLine), "body (markdown):") {
				bodyStart = i + 1
				break
			}
			// Check for markdown header format: "# Body" or "## Body"
			if _, found := extractMarkdownHeader(trimmedLine, "body"); found {
				bodyStart = i + 1
				break
			}
			// Found a non-empty line after title, start body here
			if trimmedLine != "" {
				bodyStart = i
				break
			}
		}
	}

	// Extract body
	if bodyStart > 0 && bodyStart < len(lines) {
		bodyLines := lines[bodyStart:]
		m.generatedBody = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	}

	// If no title was found, use first line
	if m.generatedTitle == "" && len(lines) > 0 {
		m.generatedTitle = strings.TrimSpace(lines[0])
		if len(lines) > 1 {
			m.generatedBody = strings.TrimSpace(strings.Join(lines[1:], "\n"))
		}
	}

	// Validate and sanitize the generated content
	m.validateGeneratedContent()
}

// extractLabeledContent extracts content after a label like "Title:" or "Body:".
// It handles case variations (Title:, title:, TITLE:) and whitespace (Title :, Title  :).
// Returns the content and true if the label was found, empty string and false otherwise.
func extractLabeledContent(line, label string) (string, bool) {
	normalized := normalizeLabel(line)
	labelLower := strings.ToLower(label) + ":"

	if strings.HasPrefix(normalized, labelLower) {
		// Find the position of the colon in the original line to extract content
		colonIdx := strings.Index(strings.ToLower(line), ":")
		if colonIdx >= 0 && colonIdx < len(line)-1 {
			return strings.TrimSpace(line[colonIdx+1:]), true
		}
		// Label found but no content after colon
		return "", true
	}
	return "", false
}

// extractMarkdownHeader extracts content from markdown header format.
// Handles "# Title", "## Title", "# Title:", "## Body", etc.
// Returns the content and true if the header was found.
func extractMarkdownHeader(line, label string) (string, bool) {
	trimmed := strings.TrimSpace(line)

	// Check for markdown header pattern: # or ## followed by the label
	if !strings.HasPrefix(trimmed, "#") {
		return "", false
	}

	// Remove leading # characters
	content := strings.TrimLeft(trimmed, "#")
	content = strings.TrimSpace(content)

	// Check if it starts with the label (case-insensitive)
	labelLower := strings.ToLower(label)
	contentLower := strings.ToLower(content)

	if strings.HasPrefix(contentLower, labelLower) {
		// Remove the label and any trailing colon
		remaining := strings.TrimSpace(content[len(label):])
		remaining = strings.TrimPrefix(remaining, ":")
		remaining = strings.TrimSpace(remaining)

		// If there's content on the same line, return it
		if remaining != "" {
			return remaining, true
		}
		// Label header found but content is on next line
		return "", true
	}

	return "", false
}

// normalizeLabel normalizes a line for label comparison by:
// - Converting to lowercase
// - Removing extra whitespace around the colon
func normalizeLabel(line string) string {
	lower := strings.ToLower(line)
	// Remove spaces before colon: "title :" -> "title:"
	// Handle multiple spaces: "title  :" -> "title:"
	result := strings.Builder{}
	prevSpace := false
	for _, r := range lower {
		if r == ' ' || r == '\t' {
			prevSpace = true
			continue
		}
		if r == ':' {
			// Skip any accumulated spaces before colon
			prevSpace = false
		} else if prevSpace {
			result.WriteRune(' ')
			prevSpace = false
		}
		result.WriteRune(r)
	}
	return result.String()
}

// validateGeneratedContent validates and sanitizes the generated title and body.
func (m *Model) validateGeneratedContent() {
	// Trim whitespace
	m.generatedTitle = strings.TrimSpace(m.generatedTitle)
	m.generatedBody = strings.TrimSpace(m.generatedBody)

	// Truncate title if too long
	if len(m.generatedTitle) > maxTitleLength {
		m.generatedTitle = m.generatedTitle[:maxTitleLength]
	}

	// Remove any remaining quotes at start/end of title
	m.generatedTitle = strings.Trim(m.generatedTitle, "\"'`")
}

// handleDraftError processes LLM errors.
func (m Model) handleDraftError(msg DraftErrorMsg) (tea.Model, tea.Cmd) {
	if m.llmCancel != nil {
		m.llmCancel()
		m.llmCancel = nil
	}

	// Don't fail completely, just move to editing without draft
	m.step = StepEditTitle
	m.titleInput.Focus()
	// Store error but don't block the wizard
	// m.err = msg.Err

	return m, nil
}

// handleSkipDraft skips the draft generation step.
func (m Model) handleSkipDraft() (tea.Model, tea.Cmd) {
	if m.llmCancel != nil {
		m.llmCancel()
		m.llmCancel = nil
	}

	m.step = StepEditTitle
	m.titleInput.Focus()
	return m, nil
}

// =============================================================================
// PR Creation Handlers
// =============================================================================

// handleSubmitPR starts the PR creation process.
func (m Model) handleSubmitPR() (tea.Model, tea.Cmd) {
	m.step = StepSubmitting
	m.spinner = m.spinner.WithMessage("Creating pull request...")

	return m, tea.Batch(
		m.spinner.Init(),
		createPR(m.ghClient, m.baseBranch, m.headBranch, m.title, m.body, m.isDraft),
	)
}

// handlePRCreated processes successful PR creation.
func (m Model) handlePRCreated(msg PRCreatedMsg) (tea.Model, tea.Cmd) {
	m.step = StepComplete
	m.createdPR = msg.PR
	return m, nil
}

// handlePRCreateError processes PR creation errors.
func (m Model) handlePRCreateError(msg PRCreateErrorMsg) (tea.Model, tea.Cmd) {
	m.step = StepError
	m.err = msg.Err
	return m, nil
}

// =============================================================================
// Key Handlers
// =============================================================================

// handleKeyMsg processes keyboard input based on current step.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.step {
	case StepBranchSelect:
		return m.handleBranchSelectKeys(msg)
	case StepDraftGeneration:
		return m.handleDraftGenerationKeys(msg)
	case StepEditTitle:
		return m.handleEditTitleKeys(msg)
	case StepEditBody:
		return m.handleEditBodyKeys(msg)
	case StepReview:
		return m.handleReviewKeys(msg)
	case StepComplete:
		return m.handleCompleteKeys(msg)
	case StepError:
		return m.handleErrorKeys(msg)
	}

	return m, nil
}

// handleBranchSelectKeys handles keys during branch selection.
func (m Model) handleBranchSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch {
	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit

	// Cancel/Back - return to dashboard
	case key.Matches(msg, m.keymap.Cancel), msg.Type == tea.KeyEscape:
		return m, backToDashboard()

	// Tab - switch between base and head branch selection
	case msg.Type == tea.KeyTab:
		m.focusNextBranch()
		return m, nil

	// Enter - select branch and proceed
	case msg.Type == tea.KeyEnter:
		selected := m.getSelectedBranch()
		if selected != "" {
			if m.branchFocus == FocusBaseBranch {
				m.baseBranch = selected
				m.focusNextBranch() // Switch to head selection
			} else {
				m.headBranch = selected
				// Both branches selected, proceed to next step
				m.step = m.nextStep()
				switch m.step {
				case StepDraftGeneration:
					return m, startDraft()
				case StepEditTitle:
					m.titleInput.Focus()
				}
			}
		}
		return m, nil

	default:
		// Pass to branch list
		if m.branchesReady {
			var cmd tea.Cmd
			m.branchList, cmd = m.branchList.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// handleDraftGenerationKeys handles keys during LLM draft generation.
func (m Model) handleDraftGenerationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// Quit
	case key.Matches(msg, m.keymap.Quit):
		m.cancelGeneration()
		return m, tea.Quit

	// Cancel generation with Ctrl+X or Escape
	case msg.Type == tea.KeyCtrlX, msg.Type == tea.KeyEscape:
		m.cancelGeneration()
		return m, skipDraft()

	// Skip with 's'
	case msg.String() == "s":
		m.cancelGeneration()
		return m, skipDraft()
	}

	return m, nil
}

// handleEditTitleKeys handles keys during title editing.
func (m Model) handleEditTitleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch {
	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit

	// Escape - go back to previous step
	case msg.Type == tea.KeyEscape:
		m.step = m.prevStep()
		if m.step == StepBranchSelect {
			m.updateBranchListItems()
		}
		return m, nil

	// Tab or Enter - proceed to body editing
	case msg.Type == tea.KeyTab, msg.Type == tea.KeyEnter:
		m.title = m.titleInput.Value()
		m.step = StepEditBody
		m.bodyInput.Focus()
		return m, nil

	default:
		// Pass to title input
		var cmd tea.Cmd
		m.titleInput, cmd = m.titleInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleEditBodyKeys handles keys during body editing.
func (m Model) handleEditBodyKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch {
	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit

	// Escape - go back to title editing
	case msg.Type == tea.KeyEscape:
		m.body = m.bodyInput.Value()
		m.step = StepEditTitle
		m.titleInput.Focus()
		return m, nil

	// Ctrl+S or Ctrl+Enter - proceed to review
	case msg.Type == tea.KeyCtrlS:
		m.body = m.bodyInput.Value()
		m.title = m.titleInput.Value()
		m.step = StepReview
		return m, nil

	// Tab - proceed to review (alternative)
	case msg.Type == tea.KeyTab:
		m.body = m.bodyInput.Value()
		m.title = m.titleInput.Value()
		m.step = StepReview
		return m, nil

	default:
		// Pass to body input
		var cmd tea.Cmd
		m.bodyInput, cmd = m.bodyInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleReviewKeys handles keys during review step.
func (m Model) handleReviewKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit

	// Escape - go back to body editing
	case msg.Type == tea.KeyEscape:
		m.step = StepEditBody
		m.bodyInput.Focus()
		return m, nil

	// 'd' - toggle draft mode
	case msg.String() == "d":
		m.toggleDraft()
		return m, nil

	// 'e' - edit (go back to title)
	case msg.String() == "e":
		m.step = StepEditTitle
		m.titleInput.Focus()
		return m, nil

	// Enter - submit PR
	case msg.Type == tea.KeyEnter:
		return m, submitPR()
	}

	return m, nil
}

// handleCompleteKeys handles keys after PR creation.
func (m Model) handleCompleteKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// Quit or Enter or Escape - return to dashboard
	case key.Matches(msg, m.keymap.Quit),
		msg.Type == tea.KeyEnter,
		msg.Type == tea.KeyEscape:
		return m, backToDashboard()
	}

	return m, nil
}

// handleErrorKeys handles keys during error state.
func (m Model) handleErrorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit

	// Escape - return to dashboard
	case msg.Type == tea.KeyEscape:
		return m, backToDashboard()

	// 'r' - retry (go back to branch selection)
	case msg.String() == "r":
		m.step = StepBranchSelect
		m.err = nil
		m.branchesReady = false
		return m, tea.Batch(
			loadBranches(m.gitRepo),
			m.spinner.Init(),
		)
	}

	return m, nil
}
