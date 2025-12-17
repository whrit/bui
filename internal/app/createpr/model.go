package createpr

import (
	"context"

	"bui/internal/config"
	"bui/internal/gh"
	"bui/internal/git"
	"bui/internal/llm"
	"bui/internal/ui"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// headerHeight is the height of the header section.
	headerHeight = 4

	// footerHeight is the height of the footer/help section.
	footerHeight = 3

	// minListHeight is the minimum height for the branch list.
	minListHeight = 5

	// defaultTextareaWidth is the default width for the textarea.
	defaultTextareaWidth = 70

	// defaultTextareaHeight is the default height for the textarea.
	defaultTextareaHeight = 10

	// maxTitleLength is the maximum length for PR title.
	maxTitleLength = 200

	// maxCommitsToShow is the maximum number of commits to display.
	maxCommitsToShow = 10
)

// =============================================================================
// Step Enum
// =============================================================================

// Step represents the current step in the wizard.
type Step int

const (
	// StepBranchSelect is the branch selection step.
	StepBranchSelect Step = iota
	// StepDraftGeneration is the LLM draft generation step.
	StepDraftGeneration
	// StepEditTitle is the title editing step.
	StepEditTitle
	// StepEditBody is the body editing step.
	StepEditBody
	// StepReview is the review step before submission.
	StepReview
	// StepSubmitting is when the PR is being submitted.
	StepSubmitting
	// StepComplete is when the PR has been created successfully.
	StepComplete
	// StepError is when an error has occurred.
	StepError
)

// String returns a string representation of the step.
func (s Step) String() string {
	switch s {
	case StepBranchSelect:
		return "branch_select"
	case StepDraftGeneration:
		return "draft_generation"
	case StepEditTitle:
		return "edit_title"
	case StepEditBody:
		return "edit_body"
	case StepReview:
		return "review"
	case StepSubmitting:
		return "submitting"
	case StepComplete:
		return "complete"
	case StepError:
		return "error"
	default:
		return "unknown"
	}
}

// DisplayName returns a human-readable name for the step.
func (s Step) DisplayName() string {
	switch s {
	case StepBranchSelect:
		return "Select Branches"
	case StepDraftGeneration:
		return "Generate Draft"
	case StepEditTitle:
		return "Edit Title"
	case StepEditBody:
		return "Edit Body"
	case StepReview:
		return "Review"
	case StepSubmitting:
		return "Submitting"
	case StepComplete:
		return "Complete"
	case StepError:
		return "Error"
	default:
		return "Unknown"
	}
}

// =============================================================================
// Branch Selection Focus
// =============================================================================

// BranchFocus indicates which branch selector is focused.
type BranchFocus int

const (
	// FocusBaseBranch focuses the base branch selector.
	FocusBaseBranch BranchFocus = iota
	// FocusHeadBranch focuses the head branch selector.
	FocusHeadBranch
)

// =============================================================================
// BranchItem for list.Model
// =============================================================================

// BranchItem wraps a branch name to implement list.Item interface.
type BranchItem struct {
	Name      string
	IsCurrent bool
}

// FilterValue implements list.Item.
func (b BranchItem) FilterValue() string {
	return b.Name
}

// Title implements list.DefaultItem.
func (b BranchItem) Title() string {
	if b.IsCurrent {
		return b.Name + " (current)"
	}
	return b.Name
}

// Description implements list.DefaultItem.
func (b BranchItem) Description() string {
	return ""
}

// =============================================================================
// Model
// =============================================================================

// Model is the Bubble Tea model for the create PR wizard.
type Model struct {
	// Dependencies
	cfg         config.Config
	ghClient    *gh.Client
	gitRepo     *git.Repo
	llmProvider llm.Provider

	// Wizard state
	step Step

	// Branch selection (Step 1)
	baseBranch    string
	headBranch    string
	branches      []string
	branchList    list.Model
	branchFocus   BranchFocus
	branchesReady bool

	// Commits between branches
	commits []git.Commit

	// PR content
	title   string
	body    string
	isDraft bool

	// Edit state (Step 3-4)
	titleInput textinput.Model
	bodyInput  textarea.Model

	// LLM streaming (Step 2)
	tokenCh        <-chan llm.Token
	errCh          <-chan error
	llmCtx         context.Context
	llmCancel      context.CancelFunc
	generatedTitle string
	generatedBody  string
	streamedText   string

	// Result
	createdPR *gh.PR
	err       error

	// UI
	spinner ui.Spinner
	palette ui.Palette
	styles  ui.Styles
	keymap  ui.KeyMap

	// Dimensions
	width, height int
}

// New creates a new create PR wizard model.
func New(cfg config.Config, ghClient *gh.Client, gitRepo *git.Repo, llmProvider llm.Provider) Model {
	palette := ui.NewPalette(cfg)
	styles := ui.NewStyles(palette)
	keymap := ui.NewKeyMap(cfg)

	// Initialize branch list
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Select Base Branch"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.Styles.Title = styles.Header

	// Initialize title input
	ti := textinput.New()
	ti.Placeholder = "Enter PR title..."
	ti.CharLimit = maxTitleLength
	ti.Width = defaultTextareaWidth

	// Initialize body textarea
	ta := textarea.New()
	ta.Placeholder = "Enter PR description..."
	ta.SetWidth(defaultTextareaWidth)
	ta.SetHeight(defaultTextareaHeight)
	ta.CharLimit = 10000
	ta.ShowLineNumbers = false

	// Initialize spinner
	spinner := ui.NewSpinner(palette, styles).WithMessage("Loading...")

	return Model{
		cfg:         cfg,
		ghClient:    ghClient,
		gitRepo:     gitRepo,
		llmProvider: llmProvider,
		step:        StepBranchSelect,
		branchList:  l,
		branchFocus: FocusBaseBranch,
		titleInput:  ti,
		bodyInput:   ta,
		spinner:     spinner,
		palette:     palette,
		styles:      styles,
		keymap:      keymap,
		isDraft:     false,
	}
}

// Init implements tea.Model. It returns the initial command to load branches.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		loadBranches(m.gitRepo),
		m.spinner.Init(),
	}
	return tea.Batch(cmds...)
}

// SetSize updates the model dimensions and recalculates component sizes.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate list dimensions
	listHeight := height - headerHeight - footerHeight - 4
	if listHeight < minListHeight {
		listHeight = minListHeight
	}
	m.branchList.SetSize(width-4, listHeight)

	// Calculate textarea dimensions
	textareaWidth := width - 8
	if textareaWidth < 30 {
		textareaWidth = 30
	}
	textareaHeight := height - headerHeight - footerHeight - 8
	if textareaHeight < 5 {
		textareaHeight = 5
	}

	m.titleInput.Width = textareaWidth
	m.bodyInput.SetWidth(textareaWidth)
	m.bodyInput.SetHeight(textareaHeight)
}

// =============================================================================
// Step Navigation
// =============================================================================

// nextStep returns the next step in the wizard, considering LLM availability.
func (m Model) nextStep() Step {
	switch m.step {
	case StepBranchSelect:
		// Skip draft generation if no LLM provider
		if m.llmProvider == nil {
			return StepEditTitle
		}
		return StepDraftGeneration
	case StepDraftGeneration:
		return StepEditTitle
	case StepEditTitle:
		return StepEditBody
	case StepEditBody:
		return StepReview
	case StepReview:
		return StepSubmitting
	case StepSubmitting:
		return StepComplete
	default:
		return m.step
	}
}

// prevStep returns the previous step in the wizard.
func (m Model) prevStep() Step {
	switch m.step {
	case StepDraftGeneration:
		return StepBranchSelect
	case StepEditTitle:
		// Skip draft generation when going back if no LLM provider
		if m.llmProvider == nil {
			return StepBranchSelect
		}
		return StepDraftGeneration
	case StepEditBody:
		return StepEditTitle
	case StepReview:
		return StepEditBody
	default:
		return m.step
	}
}

// totalSteps returns the total number of visible steps.
func (m Model) totalSteps() int {
	if m.llmProvider == nil {
		return 4 // Branch, Title, Body, Review
	}
	return 5 // Branch, Draft, Title, Body, Review
}

// currentStepNumber returns the 1-indexed step number for display.
func (m Model) currentStepNumber() int {
	if m.llmProvider == nil {
		switch m.step {
		case StepBranchSelect:
			return 1
		case StepEditTitle:
			return 2
		case StepEditBody:
			return 3
		case StepReview:
			return 4
		default:
			return 0
		}
	}

	switch m.step {
	case StepBranchSelect:
		return 1
	case StepDraftGeneration:
		return 2
	case StepEditTitle:
		return 3
	case StepEditBody:
		return 4
	case StepReview:
		return 5
	default:
		return 0
	}
}

// =============================================================================
// Accessors
// =============================================================================

// Step returns the current wizard step.
func (m Model) Step() Step {
	return m.step
}

// Title returns the PR title.
func (m Model) Title() string {
	return m.title
}

// Body returns the PR body.
func (m Model) Body() string {
	return m.body
}

// BaseBranch returns the selected base branch.
func (m Model) BaseBranch() string {
	return m.baseBranch
}

// HeadBranch returns the selected head branch.
func (m Model) HeadBranch() string {
	return m.headBranch
}

// IsDraft returns whether the PR will be created as a draft.
func (m Model) IsDraft() bool {
	return m.isDraft
}

// Error returns any error that occurred.
func (m Model) Error() error {
	return m.err
}

// CreatedPR returns the created PR if successful.
func (m Model) CreatedPR() *gh.PR {
	return m.createdPR
}

// HasLLMProvider returns true if an LLM provider is configured.
func (m Model) HasLLMProvider() bool {
	return m.llmProvider != nil
}

// Commits returns the commits between base and head branches.
func (m Model) Commits() []git.Commit {
	return m.commits
}

// =============================================================================
// State Modifiers
// =============================================================================

// toggleDraft toggles the draft mode.
func (m *Model) toggleDraft() {
	m.isDraft = !m.isDraft
}

// focusNextBranch moves focus to the next branch selector.
func (m *Model) focusNextBranch() {
	if m.branchFocus == FocusBaseBranch {
		m.branchFocus = FocusHeadBranch
		m.branchList.Title = "Select Head Branch"
	} else {
		m.branchFocus = FocusBaseBranch
		m.branchList.Title = "Select Base Branch"
	}
	m.updateBranchListItems()
}

// updateBranchListItems updates the branch list based on current focus.
func (m *Model) updateBranchListItems() {
	items := make([]list.Item, 0, len(m.branches))
	currentBranch := m.headBranch

	for _, name := range m.branches {
		// When selecting base, exclude the head branch
		if m.branchFocus == FocusBaseBranch && name == m.headBranch {
			continue
		}
		// When selecting head, exclude the base branch
		if m.branchFocus == FocusHeadBranch && name == m.baseBranch {
			continue
		}

		items = append(items, BranchItem{
			Name:      name,
			IsCurrent: name == currentBranch,
		})
	}

	m.branchList.SetItems(items)
}

// getSelectedBranch returns the currently selected branch from the list.
func (m Model) getSelectedBranch() string {
	if item := m.branchList.SelectedItem(); item != nil {
		if bi, ok := item.(BranchItem); ok {
			return bi.Name
		}
	}
	return ""
}

// =============================================================================
// LLM Streaming
// =============================================================================

// readNextToken creates a command to read the next LLM token.
// Uses a priority select pattern to ensure errors are handled promptly.
func (m Model) readNextToken() tea.Cmd {
	return func() tea.Msg {
		// First, check for errors without blocking (priority check)
		select {
		case err := <-m.errCh:
			if err != nil {
				return DraftErrorMsg{Err: err}
			}
			return DraftDoneMsg{}
		default:
			// No error pending, continue to read token
		}

		// Now read token or wait for error
		select {
		case token, ok := <-m.tokenCh:
			if !ok {
				return DraftDoneMsg{}
			}
			return DraftTokenMsg{Token: token.Text}
		case err := <-m.errCh:
			if err != nil {
				return DraftErrorMsg{Err: err}
			}
			return DraftDoneMsg{}
		}
	}
}

// cancelGeneration cancels any ongoing LLM generation.
func (m *Model) cancelGeneration() {
	if m.llmCancel != nil {
		m.llmCancel()
		m.llmCancel = nil
	}
}

// buildCommitLog builds a commit log string for LLM context.
func (m Model) buildCommitLog() string {
	if len(m.commits) == 0 {
		return "No commits found"
	}

	var result string
	count := len(m.commits)
	if count > maxCommitsToShow {
		count = maxCommitsToShow
	}

	for i := 0; i < count; i++ {
		c := m.commits[i]
		result += "- " + c.Subject + "\n"
	}

	if len(m.commits) > maxCommitsToShow {
		result += "... and more commits\n"
	}

	return result
}
