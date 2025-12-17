// Package review implements the review screen for submitting PR reviews.
// It provides a multi-step flow: select review type, compose comment (with optional LLM assistance),
// preview, and submit via `gh pr review`.
package review

import (
	"context"
	"strings"

	"bui/internal/config"
	"bui/internal/gh"
	"bui/internal/llm"
	"bui/internal/llm/prompts"
	"bui/internal/ui"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
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

	// minTextareaHeight is the minimum height for the textarea.
	minTextareaHeight = 5

	// defaultTextareaWidth is the default width for the textarea.
	defaultTextareaWidth = 70

	// defaultTextareaHeight is the default height for the textarea.
	defaultTextareaHeight = 10

	// maxDiffLinesForLLM is the maximum number of diff lines to send to LLM.
	maxDiffLinesForLLM = 500
)

// =============================================================================
// State Enum
// =============================================================================

// State represents the current state of the review screen.
type State int

const (
	// StateTypeSelect is the initial state for selecting review type.
	StateTypeSelect State = iota
	// StateCompose is the state for writing/editing the review comment.
	StateCompose
	// StatePreview is the state for previewing before submission.
	StatePreview
	// StateSubmitting is the state while submitting the review.
	StateSubmitting
	// StateComplete is the state after successful submission.
	StateComplete
	// StateError is the state when an error occurred.
	StateError
)

// String returns a string representation of the state.
func (s State) String() string {
	switch s {
	case StateTypeSelect:
		return "type_select"
	case StateCompose:
		return "compose"
	case StatePreview:
		return "preview"
	case StateSubmitting:
		return "submitting"
	case StateComplete:
		return "complete"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// =============================================================================
// ReviewType Enum
// =============================================================================

// ReviewType represents the type of review action.
type ReviewType int

const (
	// ReviewTypeApprove approves the PR.
	ReviewTypeApprove ReviewType = iota
	// ReviewTypeRequestChanges requests changes on the PR.
	ReviewTypeRequestChanges
	// ReviewTypeComment leaves a comment without approval/rejection.
	ReviewTypeComment
)

// String returns a string representation of the review type.
func (rt ReviewType) String() string {
	switch rt {
	case ReviewTypeApprove:
		return "approve"
	case ReviewTypeRequestChanges:
		return "request_changes"
	case ReviewTypeComment:
		return "comment"
	default:
		return "unknown"
	}
}

// DisplayName returns a human-readable name for the review type.
func (rt ReviewType) DisplayName() string {
	switch rt {
	case ReviewTypeApprove:
		return "Approve"
	case ReviewTypeRequestChanges:
		return "Request Changes"
	case ReviewTypeComment:
		return "Comment"
	default:
		return "Unknown"
	}
}

// ToGHReviewEvent converts the ReviewType to gh.ReviewEvent.
func (rt ReviewType) ToGHReviewEvent() gh.ReviewEvent {
	switch rt {
	case ReviewTypeApprove:
		return gh.ReviewApprove
	case ReviewTypeRequestChanges:
		return gh.ReviewRequestChanges
	case ReviewTypeComment:
		return gh.ReviewComment
	default:
		return gh.ReviewComment
	}
}

// Icon returns an icon/symbol for the review type.
func (rt ReviewType) Icon() string {
	switch rt {
	case ReviewTypeApprove:
		return "+"
	case ReviewTypeRequestChanges:
		return "x"
	case ReviewTypeComment:
		return "#"
	default:
		return "?"
	}
}

// =============================================================================
// Model
// =============================================================================

// Model is the Bubble Tea model for the review screen.
type Model struct {
	// Dependencies
	cfg         config.Config
	ghClient    *gh.Client
	llmProvider llm.Provider

	// PR context
	prNumber int
	prTitle  string
	diff     string // PR diff for LLM context

	// State
	state      State
	reviewType ReviewType

	// Comment editing
	textarea textarea.Model
	comment  string

	// LLM streaming
	tokenCh          <-chan llm.Token
	errCh            <-chan error
	llmCtx           context.Context
	llmCancel        context.CancelFunc
	generatedComment string
	isGenerating     bool

	// Result
	err error

	// UI
	spinner  ui.Spinner
	palette  ui.Palette
	styles   ui.Styles
	keymap   ui.KeyMap
	viewport viewport.Model // For preview

	// Dimensions
	width, height int
}

// New creates a new review screen model.
func New(cfg config.Config, ghClient *gh.Client, llmProvider llm.Provider, prNumber int, prTitle string, diff string) Model {
	palette := ui.NewPalette(cfg)
	styles := ui.NewStyles(palette)
	keymap := ui.NewKeyMap(cfg)

	// Initialize textarea
	ta := textarea.New()
	ta.Placeholder = "Enter your review comment..."
	ta.Focus()
	ta.SetWidth(defaultTextareaWidth)
	ta.SetHeight(defaultTextareaHeight)
	ta.CharLimit = 10000
	ta.ShowLineNumbers = false

	// Initialize viewport for preview
	vp := viewport.New(defaultTextareaWidth, defaultTextareaHeight)

	// Initialize spinner
	spinner := ui.NewSpinner(palette, styles).WithMessage("Submitting review...")

	return Model{
		cfg:         cfg,
		ghClient:    ghClient,
		llmProvider: llmProvider,
		prNumber:    prNumber,
		prTitle:     prTitle,
		diff:        diff,
		state:       StateTypeSelect,
		reviewType:  ReviewTypeApprove, // Default to approve
		textarea:    ta,
		viewport:    vp,
		spinner:     spinner,
		palette:     palette,
		styles:      styles,
		keymap:      keymap,
	}
}

// Init implements tea.Model. It returns the initial command to start the textarea cursor blink.
func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

// SetSize updates the model dimensions and recalculates component sizes.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate textarea dimensions
	textareaWidth := width - 8
	if textareaWidth < 30 {
		textareaWidth = 30
	}
	textareaHeight := height - headerHeight - footerHeight - 8
	if textareaHeight < minTextareaHeight {
		textareaHeight = minTextareaHeight
	}

	m.textarea.SetWidth(textareaWidth)
	m.textarea.SetHeight(textareaHeight)

	// Update viewport for preview
	m.viewport.Width = textareaWidth
	m.viewport.Height = textareaHeight
}

// =============================================================================
// Accessors
// =============================================================================

// State returns the current state.
func (m Model) State() State {
	return m.state
}

// Error returns the current error, if any.
func (m Model) Error() error {
	return m.err
}

// ReviewType returns the selected review type.
func (m Model) ReviewType() ReviewType {
	return m.reviewType
}

// PRNumber returns the PR number being reviewed.
func (m Model) PRNumber() int {
	return m.prNumber
}

// PRTitle returns the PR title.
func (m Model) PRTitle() string {
	return m.prTitle
}

// Comment returns the current comment text.
func (m Model) Comment() string {
	return m.comment
}

// HasProvider returns true if an LLM provider is configured.
func (m Model) HasProvider() bool {
	return m.llmProvider != nil
}

// IsGenerating returns true if LLM generation is in progress.
func (m Model) IsGenerating() bool {
	return m.isGenerating
}

// =============================================================================
// Validation
// =============================================================================

// canProceedToPreview checks if the current state allows proceeding to preview.
func (m Model) canProceedToPreview() bool {
	switch m.reviewType {
	case ReviewTypeApprove:
		// Approve can have empty comment
		return true
	case ReviewTypeRequestChanges:
		// Request changes must have a comment
		return strings.TrimSpace(m.comment) != ""
	case ReviewTypeComment:
		// Comment must have text
		return strings.TrimSpace(m.comment) != ""
	default:
		return false
	}
}

// =============================================================================
// LLM Streaming
// =============================================================================

// readNextToken creates a command to read the next LLM token.
func (m Model) readNextToken() tea.Cmd {
	return func() tea.Msg {
		select {
		case token, ok := <-m.tokenCh:
			if !ok {
				return GenerateCommentDoneMsg{}
			}
			return GenerateCommentTokenMsg{Token: token.Text}
		case err := <-m.errCh:
			if err != nil {
				return GenerateCommentErrorMsg{Err: err}
			}
			return GenerateCommentDoneMsg{}
		}
	}
}

// startGeneration begins LLM comment generation.
func (m *Model) startGeneration() tea.Cmd {
	if m.llmProvider == nil {
		return nil
	}

	m.isGenerating = true
	m.generatedComment = "" // Clear previous output

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	m.llmCtx = ctx
	m.llmCancel = cancel

	// Build prompt
	prompt := m.buildLLMPrompt()

	// Start streaming
	tokenCh, errCh := m.llmProvider.Stream(ctx, prompt)
	m.tokenCh = tokenCh
	m.errCh = errCh

	// Return commands to start spinner and read first token
	return tea.Batch(
		m.spinner.Init(),
		m.readNextToken(),
	)
}

// cancelGeneration cancels any ongoing LLM generation.
func (m *Model) cancelGeneration() {
	if m.llmCancel != nil {
		m.llmCancel()
		m.llmCancel = nil
	}
	m.isGenerating = false
}

// buildLLMPrompt creates the prompt for LLM comment generation.
func (m Model) buildLLMPrompt() llm.Prompt {
	// Truncate diff if too long
	diff := m.diff
	lines := strings.Split(diff, "\n")
	if len(lines) > maxDiffLinesForLLM {
		lines = lines[:maxDiffLinesForLLM]
		diff = strings.Join(lines, "\n") + "\n... (truncated)"
	}

	vars := prompts.PRReviewCommentVars{
		ReviewType:      m.reviewType.String(),
		PRTitle:         m.prTitle,
		PRNumber:        m.prNumber,
		Diff:            diff,
		ExistingComment: m.textarea.Value(),
	}

	system, user, _ := prompts.RenderPRReviewComment(vars)

	return llm.Prompt{
		System:      system,
		User:        user,
		MaxTokens:   m.cfg.LLM.MaxTokens,
		Temperature: m.cfg.LLM.Temperature,
	}
}

// =============================================================================
// Review Submission
// =============================================================================

// submitReview creates a command to submit the review via gh.
func (m Model) submitReview() tea.Cmd {
	return func() tea.Msg {
		if m.ghClient == nil {
			return ReviewSubmitErrorMsg{Err: errNoGHClient}
		}

		event := m.reviewType.ToGHReviewEvent()
		err := m.ghClient.SubmitReview(m.prNumber, m.comment, event)
		if err != nil {
			return ReviewSubmitErrorMsg{Err: err}
		}

		return ReviewSubmittedMsg{}
	}
}

// errNoGHClient is returned when no gh client is available.
var errNoGHClient = &noGHClientError{}

type noGHClientError struct{}

func (e *noGHClientError) Error() string {
	return "no GitHub client configured"
}
