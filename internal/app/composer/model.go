// Package composer implements the composer screen for the bui application.
// It provides a multiline text editor with LLM streaming support for generating
// PR descriptions, commit messages, and other content.
package composer

import (
	"context"

	"bui/internal/config"
	"bui/internal/llm"
	"bui/internal/ui"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// headerHeight is the height of the header section.
	headerHeight = 3

	// footerHeight is the height of the footer/help section.
	footerHeight = 2

	// minTextareaHeight is the minimum height for the textarea.
	minTextareaHeight = 5

	// defaultTextareaWidth is the default width for the textarea.
	defaultTextareaWidth = 80

	// defaultTextareaHeight is the default height for the textarea.
	defaultTextareaHeight = 10
)

// =============================================================================
// State
// =============================================================================

// State represents the current state of the composer screen.
type State int

const (
	// StateEditing indicates the user is editing text in the textarea.
	StateEditing State = iota
	// StateGenerating indicates LLM is generating content.
	StateGenerating
	// StateReady indicates LLM generation is complete, content can be accepted or edited.
	StateReady
	// StateError indicates an error occurred.
	StateError
)

// String returns a string representation of the state.
func (s State) String() string {
	switch s {
	case StateEditing:
		return "editing"
	case StateGenerating:
		return "generating"
	case StateReady:
		return "ready"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// =============================================================================
// Model
// =============================================================================

// Model is the Bubble Tea model for the composer screen.
type Model struct {
	// Dependencies
	cfg         config.Config
	llmProvider llm.Provider

	// State
	state State
	err   error

	// UI components
	textarea textarea.Model
	spinner  ui.Spinner

	// LLM streaming
	tokenCh      <-chan llm.Token
	errCh        <-chan error
	llmCtx       context.Context
	llmCancel    context.CancelFunc
	streamedText string // Accumulated LLM output

	// Styling
	palette ui.Palette
	styles  ui.Styles
	keymap  ui.KeyMap

	// Dimensions
	width, height int

	// Callbacks/context
	prompt   llm.Prompt           // The prompt to use for generation
	onSubmit func(string) tea.Cmd // Callback when user accepts
	onCancel func() tea.Cmd       // Callback when user cancels
}

// New creates a new composer model.
func New(cfg config.Config, llmProvider llm.Provider, prompt llm.Prompt, onSubmit func(string) tea.Cmd, onCancel func() tea.Cmd) Model {
	palette := ui.NewPalette(cfg)
	styles := ui.NewStyles(palette)
	keymap := ui.NewKeyMap(cfg)

	// Initialize textarea
	ta := textarea.New()
	ta.Placeholder = "Enter your text here, or press Ctrl+G to generate with LLM..."
	ta.Focus()
	ta.SetWidth(defaultTextareaWidth)
	ta.SetHeight(defaultTextareaHeight)
	ta.CharLimit = 10000
	ta.ShowLineNumbers = false

	// Initialize spinner
	spinner := ui.NewSpinner(palette, styles).WithMessage("Generating...")

	return Model{
		cfg:         cfg,
		llmProvider: llmProvider,
		state:       StateEditing,
		textarea:    ta,
		spinner:     spinner,
		palette:     palette,
		styles:      styles,
		keymap:      keymap,
		prompt:      prompt,
		onSubmit:    onSubmit,
		onCancel:    onCancel,
	}
}

// Init implements tea.Model. It returns the initial command to start the textarea cursor blink.
func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

// SetSize updates the model dimensions and recalculates textarea size.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate textarea dimensions accounting for header and footer
	textareaWidth := width - 4 // Account for border padding
	if textareaWidth < 20 {
		textareaWidth = 20
	}

	textareaHeight := height - headerHeight - footerHeight - 2
	if textareaHeight < minTextareaHeight {
		textareaHeight = minTextareaHeight
	}

	m.textarea.SetWidth(textareaWidth)
	m.textarea.SetHeight(textareaHeight)
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

// StreamedText returns the accumulated LLM output.
func (m Model) StreamedText() string {
	return m.streamedText
}

// GetContent returns the current content from the textarea or streamed text.
func (m Model) GetContent() string {
	// If we have streamed text from LLM, return that
	if m.streamedText != "" {
		return m.streamedText
	}
	// Otherwise return textarea content
	return m.textarea.Value()
}

// SetContent sets the content of the textarea.
func (m *Model) SetContent(content string) {
	m.textarea.SetValue(content)
	m.streamedText = content
}

// HasProvider returns true if an LLM provider is configured.
func (m Model) HasProvider() bool {
	return m.llmProvider != nil
}

// IsGenerating returns true if LLM generation is in progress.
func (m Model) IsGenerating() bool {
	return m.state == StateGenerating
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
				return GenerateDoneMsg{}
			}
			return GenerateTokenMsg{Token: token.Text}
		case err := <-m.errCh:
			if err != nil {
				return GenerateErrorMsg{Err: err}
			}
			return GenerateDoneMsg{}
		}
	}
}

// startGeneration begins LLM content generation.
func (m *Model) startGeneration() tea.Cmd {
	if m.llmProvider == nil {
		return nil
	}

	m.state = StateGenerating
	m.streamedText = "" // Clear previous output

	// Create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	m.llmCtx = ctx
	m.llmCancel = cancel

	// Start streaming
	tokenCh, errCh := m.llmProvider.Stream(ctx, m.prompt)
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
	m.state = StateEditing
}
