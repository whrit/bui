package prdetail

import (
	"context"

	"bui/internal/config"
	"bui/internal/gh"
	"bui/internal/llm"
	"bui/internal/ui"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// headerHeight is the height of the header section.
	headerHeight = 6

	// tabHeight is the height of the tab bar.
	tabHeight = 1

	// footerHeight is the height of the footer/help section.
	footerHeight = 2

	// minViewportHeight is the minimum viewport height.
	minViewportHeight = 5
)

// =============================================================================
// State
// =============================================================================

// State represents the current state of the PR detail screen.
type State int

const (
	// StateLoading indicates data is being loaded.
	StateLoading State = iota
	// StateReady indicates data is loaded and ready for interaction.
	StateReady
	// StateError indicates an error occurred.
	StateError
	// StateSummarizing indicates LLM is generating a summary.
	StateSummarizing
	// StateMerging indicates a merge operation is in progress.
	StateMerging
	// StateClosing indicates a close operation is in progress.
	StateClosing
	// StateConfirming indicates a confirmation dialog is shown.
	StateConfirming
)

// String returns a string representation of the state.
func (s State) String() string {
	switch s {
	case StateLoading:
		return "loading"
	case StateReady:
		return "ready"
	case StateError:
		return "error"
	case StateSummarizing:
		return "summarizing"
	case StateMerging:
		return "merging"
	case StateClosing:
		return "closing"
	case StateConfirming:
		return "confirming"
	default:
		return "unknown"
	}
}

// =============================================================================
// Section
// =============================================================================

// Section represents the active section/tab in the PR detail view.
type Section int

const (
	// SectionInfo shows basic PR information.
	SectionInfo Section = iota
	// SectionDescription shows the PR description/body.
	SectionDescription
	// SectionChecks shows CI check status.
	SectionChecks
	// SectionReviews shows reviewer states.
	SectionReviews
)

// String returns the display name of the section.
func (s Section) String() string {
	switch s {
	case SectionInfo:
		return "Info"
	case SectionDescription:
		return "Description"
	case SectionChecks:
		return "Checks"
	case SectionReviews:
		return "Reviews"
	default:
		return "Unknown"
	}
}

// sectionCount is the total number of sections.
const sectionCount = 4

// =============================================================================
// ConfirmAction
// =============================================================================

// ConfirmAction represents the action being confirmed.
type ConfirmAction int

const (
	// ConfirmNone means no confirmation is pending.
	ConfirmNone ConfirmAction = iota
	// ConfirmMerge means merge is being confirmed.
	ConfirmMerge
	// ConfirmClose means close is being confirmed.
	ConfirmClose
)

// =============================================================================
// Model
// =============================================================================

// Model is the Bubble Tea model for the PR detail screen.
type Model struct {
	// Dependencies
	cfg         config.Config
	ghClient    *gh.Client
	llmProvider llm.Provider

	// Data
	pr       *gh.PR
	prNumber int
	checks   []gh.Check
	reviews  []gh.Review
	diff     string
	summary  string // LLM-generated summary

	// Loading state tracking
	prLoaded      bool
	checksLoaded  bool
	reviewsLoaded bool
	diffLoaded    bool

	// UI state
	state         State
	activeSection Section
	viewport      viewport.Model
	spinner       ui.Spinner
	confirmDialog ui.ConfirmDialog
	confirmAction ConfirmAction
	mergeMethod   gh.MergeMethod
	err           error

	// LLM streaming
	llmCtx    context.Context
	llmCancel context.CancelFunc
	tokenCh   <-chan llm.Token
	errCh     <-chan error

	// Styling
	palette ui.Palette
	styles  ui.Styles
	keymap  ui.KeyMap

	// Dimensions
	width, height int
}

// New creates a new PR detail model.
func New(cfg config.Config, prNumber int) Model {
	palette := ui.NewPalette(cfg)
	styles := ui.NewStyles(palette)
	keymap := ui.NewKeyMap(cfg)

	// Initialize viewport (will be resized on first WindowSizeMsg)
	vp := viewport.New(80, 20)
	vp.Style = styles.App

	// Initialize spinner
	spinner := ui.NewSpinner(palette, styles).WithMessage("Loading PR details...")

	// Initialize confirm dialog
	confirmDialog := ui.NewConfirmDialog("", "", palette, styles)

	// Default merge method
	mergeMethod := gh.MergeMethodSquash

	return Model{
		cfg:           cfg,
		ghClient:      gh.New(),
		prNumber:      prNumber,
		state:         StateLoading,
		activeSection: SectionInfo,
		viewport:      vp,
		spinner:       spinner,
		confirmDialog: confirmDialog,
		confirmAction: ConfirmNone,
		mergeMethod:   mergeMethod,
		palette:       palette,
		styles:        styles,
		keymap:        keymap,
	}
}

// NewWithClient creates a new PR detail model with a custom GitHub client.
// This is useful for testing.
func NewWithClient(cfg config.Config, prNumber int, client *gh.Client) Model {
	m := New(cfg, prNumber)
	m.ghClient = client
	return m
}

// NewWithProvider creates a new PR detail model with an LLM provider.
func NewWithProvider(cfg config.Config, prNumber int, provider llm.Provider) Model {
	m := New(cfg, prNumber)
	m.llmProvider = provider
	return m
}

// Init implements tea.Model. It returns the initial commands to load data.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadPR(m.ghClient, m.prNumber),
		loadChecks(m.ghClient, m.prNumber),
		loadReviews(m.ghClient, m.prNumber),
		loadDiff(m.ghClient, m.prNumber),
		m.spinner.Init(),
	)
}

// SetSize updates the model dimensions and recalculates viewport size.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate viewport dimensions accounting for header, tabs, and footer
	viewportHeight := height - headerHeight - tabHeight - footerHeight
	if viewportHeight < minViewportHeight {
		viewportHeight = minViewportHeight
	}

	m.viewport.Width = width - 4 // Account for border padding
	m.viewport.Height = viewportHeight
}

// PRNumber returns the PR number being displayed.
func (m Model) PRNumber() int {
	return m.prNumber
}

// PR returns the loaded PR data, or nil if not yet loaded.
func (m Model) PR() *gh.PR {
	return m.pr
}

// State returns the current state.
func (m Model) State() State {
	return m.state
}

// ActiveSection returns the currently active section.
func (m Model) ActiveSection() Section {
	return m.activeSection
}

// Summary returns the LLM-generated summary.
func (m Model) Summary() string {
	return m.summary
}

// Checks returns the loaded CI checks.
func (m Model) Checks() []gh.Check {
	return m.checks
}

// Reviews returns the loaded reviews.
func (m Model) Reviews() []gh.Review {
	return m.reviews
}

// IsLoading returns true if any data is still being loaded.
func (m Model) IsLoading() bool {
	return !m.prLoaded || !m.checksLoaded || !m.reviewsLoaded || !m.diffLoaded
}

// allDataLoaded checks if all data has been loaded.
func (m Model) allDataLoaded() bool {
	return m.prLoaded && m.checksLoaded && m.reviewsLoaded && m.diffLoaded
}

// canMerge returns true if the PR can be merged.
func (m Model) canMerge() bool {
	if m.pr == nil {
		return false
	}
	return m.pr.State == "OPEN" && m.pr.Mergeable == "MERGEABLE"
}

// canClose returns true if the PR can be closed.
func (m Model) canClose() bool {
	if m.pr == nil {
		return false
	}
	return m.pr.State == "OPEN"
}

// nextSection moves to the next section (wrapping around).
func (m *Model) nextSection() {
	m.activeSection = Section((int(m.activeSection) + 1) % sectionCount)
	m.updateViewportContent()
}

// prevSection moves to the previous section (wrapping around).
func (m *Model) prevSection() {
	m.activeSection = Section((int(m.activeSection) - 1 + sectionCount) % sectionCount)
	m.updateViewportContent()
}

// updateViewportContent updates the viewport content based on the active section.
func (m *Model) updateViewportContent() {
	var content string
	switch m.activeSection {
	case SectionInfo:
		content = m.renderInfoSection()
	case SectionDescription:
		content = m.renderDescriptionSection()
	case SectionChecks:
		content = m.renderChecksSection()
	case SectionReviews:
		content = m.renderReviewsSection()
	}
	m.viewport.SetContent(content)
	m.viewport.GotoTop()
}

// SetLLMProvider sets the LLM provider for summary generation.
func (m *Model) SetLLMProvider(provider llm.Provider) {
	m.llmProvider = provider
}
