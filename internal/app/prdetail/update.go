package prdetail

import (
	"context"

	"bui/internal/llm"
	"bui/internal/llm/prompts"
	"bui/internal/ui"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model. It handles all incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// Window size change
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		m.updateViewportContent()
		return m, nil

	// PR loaded
	case PRLoadedMsg:
		return m.handlePRLoaded(msg)

	// Checks loaded
	case ChecksLoadedMsg:
		return m.handleChecksLoaded(msg)

	// Reviews loaded
	case ReviewsLoadedMsg:
		return m.handleReviewsLoaded(msg)

	// Diff loaded
	case DiffLoadedMsg:
		return m.handleDiffLoaded(msg)

	// Load error
	case LoadErrorMsg:
		return m.handleLoadError(msg)

	// LLM summary messages
	case SummaryStartMsg:
		return m.handleSummaryStart()

	case SummaryTokenMsg:
		return m.handleSummaryToken(msg)

	case SummaryDoneMsg:
		return m.handleSummaryDone()

	case SummaryErrorMsg:
		return m.handleSummaryError(msg)

	// Merge/Close messages
	case MergeDoneMsg:
		return m.handleMergeDone()

	case MergeErrorMsg:
		return m.handleMergeError(msg)

	case CloseDoneMsg:
		return m.handleCloseDone()

	case CloseErrorMsg:
		return m.handleCloseError(msg)

	// Spinner tick
	case spinner.TickMsg:
		if m.state == StateLoading || m.state == StateSummarizing ||
			m.state == StateMerging || m.state == StateClosing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	// Keyboard input
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	// Pass remaining messages to viewport when ready
	if m.state == StateReady {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// =============================================================================
// Data Loading Handlers
// =============================================================================

// handlePRLoaded processes successfully loaded PR data.
func (m Model) handlePRLoaded(msg PRLoadedMsg) (tea.Model, tea.Cmd) {
	m.pr = msg.PR
	m.prLoaded = true

	if m.allDataLoaded() {
		m.state = StateReady
		m.updateViewportContent()
	}

	return m, nil
}

// handleChecksLoaded processes loaded CI checks.
func (m Model) handleChecksLoaded(msg ChecksLoadedMsg) (tea.Model, tea.Cmd) {
	m.checks = msg.Checks
	m.checksLoaded = true

	if m.allDataLoaded() {
		m.state = StateReady
		m.updateViewportContent()
	}

	return m, nil
}

// handleReviewsLoaded processes loaded reviews.
func (m Model) handleReviewsLoaded(msg ReviewsLoadedMsg) (tea.Model, tea.Cmd) {
	m.reviews = msg.Reviews
	m.reviewsLoaded = true

	if m.allDataLoaded() {
		m.state = StateReady
		m.updateViewportContent()
	}

	return m, nil
}

// handleDiffLoaded processes loaded diff.
func (m Model) handleDiffLoaded(msg DiffLoadedMsg) (tea.Model, tea.Cmd) {
	m.diff = msg.Diff
	m.diffLoaded = true

	if m.allDataLoaded() {
		m.state = StateReady
		m.updateViewportContent()
	}

	return m, nil
}

// handleLoadError processes loading errors.
func (m Model) handleLoadError(msg LoadErrorMsg) (tea.Model, tea.Cmd) {
	m.state = StateError
	m.err = msg.Err
	return m, nil
}

// =============================================================================
// LLM Handlers
// =============================================================================

// handleSummaryStart begins the LLM summary generation.
func (m Model) handleSummaryStart() (tea.Model, tea.Cmd) {
	if m.llmProvider == nil {
		m.state = StateReady
		return m, nil
	}

	m.state = StateSummarizing
	m.spinner = m.spinner.WithMessage("Generating summary...")
	m.summary = "" // Clear previous summary

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	m.llmCtx = ctx
	m.llmCancel = cancel

	// Build prompt
	system, user, err := prompts.RenderPRSummary(prompts.PRSummaryVars{
		Title:       m.pr.Title,
		Description: m.pr.Body,
		CommitLog:   "", // Not available in this context
		DiffStat:    m.diff,
	})
	if err != nil {
		return m, func() tea.Msg {
			return SummaryErrorMsg{Err: err}
		}
	}

	prompt := llm.Prompt{
		System:      system,
		User:        user,
		MaxTokens:   500,
		Temperature: 0.3,
	}

	// Start streaming
	tokenCh, errCh := m.llmProvider.Stream(ctx, prompt)
	m.tokenCh = tokenCh
	m.errCh = errCh

	// Return command to read first token
	return m, m.readNextToken()
}

// readNextToken creates a command to read the next LLM token.
func (m Model) readNextToken() tea.Cmd {
	return func() tea.Msg {
		select {
		case token, ok := <-m.tokenCh:
			if !ok {
				return SummaryDoneMsg{}
			}
			return SummaryTokenMsg{Token: token.Text}
		case err := <-m.errCh:
			if err != nil {
				return SummaryErrorMsg{Err: err}
			}
			return SummaryDoneMsg{}
		}
	}
}

// handleSummaryToken processes an incoming LLM token.
func (m Model) handleSummaryToken(msg SummaryTokenMsg) (tea.Model, tea.Cmd) {
	m.summary += msg.Token
	m.updateViewportContent()
	return m, m.readNextToken()
}

// handleSummaryDone completes the LLM summary generation.
func (m Model) handleSummaryDone() (tea.Model, tea.Cmd) {
	m.state = StateReady
	if m.llmCancel != nil {
		m.llmCancel()
	}
	m.updateViewportContent()
	return m, nil
}

// handleSummaryError processes LLM errors.
func (m Model) handleSummaryError(msg SummaryErrorMsg) (tea.Model, tea.Cmd) {
	m.state = StateError
	m.err = msg.Err
	if m.llmCancel != nil {
		m.llmCancel()
	}
	return m, nil
}

// =============================================================================
// Action Handlers
// =============================================================================

// handleMergeDone processes successful merge.
func (m Model) handleMergeDone() (tea.Model, tea.Cmd) {
	m.state = StateReady
	// Reload PR to get updated state
	return m, loadPR(m.ghClient, m.prNumber)
}

// handleMergeError processes merge errors.
func (m Model) handleMergeError(msg MergeErrorMsg) (tea.Model, tea.Cmd) {
	m.state = StateError
	m.err = msg.Err
	return m, nil
}

// handleCloseDone processes successful close.
func (m Model) handleCloseDone() (tea.Model, tea.Cmd) {
	m.state = StateReady
	// Reload PR to get updated state
	return m, loadPR(m.ghClient, m.prNumber)
}

// handleCloseError processes close errors.
func (m Model) handleCloseError(msg CloseErrorMsg) (tea.Model, tea.Cmd) {
	m.state = StateError
	m.err = msg.Err
	return m, nil
}

// =============================================================================
// Key Handlers
// =============================================================================

// handleKeyMsg processes keyboard input.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle confirmation dialog if visible
	if m.state == StateConfirming {
		return m.handleConfirmKeys(msg)
	}

	// Handle state-specific key bindings
	switch m.state {
	case StateLoading:
		return m.handleLoadingKeys(msg)
	case StateError:
		return m.handleErrorKeys(msg)
	case StateSummarizing:
		return m.handleSummarizingKeys(msg)
	case StateMerging, StateClosing:
		return m.handleActionInProgressKeys(msg)
	case StateReady:
		return m.handleReadyKeys(msg)
	}

	return m, nil
}

// handleLoadingKeys handles keys during loading state.
func (m Model) handleLoadingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Back), key.Matches(msg, m.keymap.Cancel):
		return m, backToDashboard()
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit
	}
	return m, nil
}

// handleErrorKeys handles keys during error state.
func (m Model) handleErrorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Back), key.Matches(msg, m.keymap.Cancel):
		return m, backToDashboard()
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keymap.Refresh):
		// Retry loading
		m.state = StateLoading
		m.err = nil
		m.prLoaded = false
		m.checksLoaded = false
		m.reviewsLoaded = false
		m.diffLoaded = false
		return m, tea.Batch(
			loadPR(m.ghClient, m.prNumber),
			loadChecks(m.ghClient, m.prNumber),
			loadReviews(m.ghClient, m.prNumber),
			loadDiff(m.ghClient, m.prNumber),
			m.spinner.Init(),
		)
	}
	return m, nil
}

// handleSummarizingKeys handles keys during LLM summarization.
func (m Model) handleSummarizingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Cancel):
		// Cancel LLM streaming
		if m.llmCancel != nil {
			m.llmCancel()
		}
		m.state = StateReady
		return m, nil
	case key.Matches(msg, m.keymap.Quit):
		if m.llmCancel != nil {
			m.llmCancel()
		}
		return m, tea.Quit
	}
	return m, nil
}

// handleActionInProgressKeys handles keys during merge/close operations.
func (m Model) handleActionInProgressKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Only allow quit during operations
	if key.Matches(msg, m.keymap.Quit) {
		return m, tea.Quit
	}
	return m, nil
}

// handleReadyKeys handles keys when ready for interaction.
func (m Model) handleReadyKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch {
	// Back to dashboard
	case key.Matches(msg, m.keymap.Back), key.Matches(msg, m.keymap.Cancel):
		return m, backToDashboard()

	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit

	// Tab navigation
	case msg.Type == tea.KeyTab:
		m.nextSection()
		return m, nil

	case msg.Type == tea.KeyShiftTab:
		m.prevSection()
		return m, nil

	// Section selection by number
	case msg.String() == "1":
		m.activeSection = SectionInfo
		m.updateViewportContent()
		return m, nil

	case msg.String() == "2":
		m.activeSection = SectionDescription
		m.updateViewportContent()
		return m, nil

	case msg.String() == "3":
		m.activeSection = SectionChecks
		m.updateViewportContent()
		return m, nil

	case msg.String() == "4":
		m.activeSection = SectionReviews
		m.updateViewportContent()
		return m, nil

	// View diff
	case key.Matches(msg, m.keymap.ViewDiff):
		return m, viewDiff(m.prNumber)

	// Open in browser
	case key.Matches(msg, m.keymap.OpenInBrowser):
		if m.pr != nil {
			return m, openBrowser(m.pr.URL)
		}
		return m, nil

	// Generate LLM summary
	case key.Matches(msg, m.keymap.GenerateWithLLM):
		if m.llmProvider != nil && m.pr != nil {
			return m, func() tea.Msg { return SummaryStartMsg{} }
		}
		return m, nil

	// Merge PR
	case key.Matches(msg, m.keymap.MergePR):
		if m.canMerge() {
			return m.showMergeConfirmation()
		}
		return m, nil

	// Close PR
	case key.Matches(msg, m.keymap.ClosePR):
		if m.canClose() {
			return m.showCloseConfirmation()
		}
		return m, nil

	// Refresh
	case key.Matches(msg, m.keymap.Refresh):
		m.state = StateLoading
		m.prLoaded = false
		m.checksLoaded = false
		m.reviewsLoaded = false
		m.diffLoaded = false
		return m, tea.Batch(
			loadPR(m.ghClient, m.prNumber),
			loadChecks(m.ghClient, m.prNumber),
			loadReviews(m.ghClient, m.prNumber),
			loadDiff(m.ghClient, m.prNumber),
			m.spinner.Init(),
		)

	// Viewport navigation
	case key.Matches(msg, m.keymap.NextItem), msg.String() == "j":
		m.viewport.LineDown(1)
		return m, nil

	case key.Matches(msg, m.keymap.PrevItem), msg.String() == "k":
		m.viewport.LineUp(1)
		return m, nil

	case key.Matches(msg, m.keymap.PageDown):
		m.viewport.ViewDown()
		return m, nil

	case key.Matches(msg, m.keymap.PageUp):
		m.viewport.ViewUp()
		return m, nil

	case key.Matches(msg, m.keymap.Home):
		m.viewport.GotoTop()
		return m, nil

	case key.Matches(msg, m.keymap.End):
		m.viewport.GotoBottom()
		return m, nil
	}

	// Pass to viewport for scroll handling
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// handleConfirmKeys handles keys during confirmation dialog.
func (m Model) handleConfirmKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Update confirm dialog
	var cmd tea.Cmd
	m.confirmDialog, cmd = m.confirmDialog.Update(msg)

	// Check if dialog was dismissed
	if !m.confirmDialog.Visible() {
		if m.confirmDialog.Confirmed() {
			// Execute confirmed action
			switch m.confirmAction {
			case ConfirmMerge:
				m.state = StateMerging
				m.spinner = m.spinner.WithMessage("Merging PR...")
				return m, tea.Batch(
					mergePR(m.ghClient, m.prNumber, m.mergeMethod),
					m.spinner.Init(),
				)
			case ConfirmClose:
				m.state = StateClosing
				m.spinner = m.spinner.WithMessage("Closing PR...")
				return m, tea.Batch(
					closePR(m.ghClient, m.prNumber),
					m.spinner.Init(),
				)
			}
		}
		// Dialog was cancelled
		m.state = StateReady
		m.confirmAction = ConfirmNone
	}

	return m, cmd
}

// =============================================================================
// Confirmation Helpers
// =============================================================================

// showMergeConfirmation shows the merge confirmation dialog.
func (m Model) showMergeConfirmation() (tea.Model, tea.Cmd) {
	m.state = StateConfirming
	m.confirmAction = ConfirmMerge
	m.confirmDialog = m.confirmDialog.Show()

	// Update dialog text
	methodName := string(m.mergeMethod)
	m.confirmDialog = ui.NewConfirmDialog(
		"Merge Pull Request",
		"Are you sure you want to merge this PR using "+methodName+"?",
		m.palette,
		m.styles,
	).Show()

	return m, nil
}

// showCloseConfirmation shows the close confirmation dialog.
func (m Model) showCloseConfirmation() (tea.Model, tea.Cmd) {
	m.state = StateConfirming
	m.confirmAction = ConfirmClose

	m.confirmDialog = ui.NewConfirmDialog(
		"Close Pull Request",
		"Are you sure you want to close this PR without merging?",
		m.palette,
		m.styles,
	).Show()

	return m, nil
}

