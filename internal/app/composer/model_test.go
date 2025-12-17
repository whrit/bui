package composer

import (
	"context"
	"errors"
	"testing"

	"bui/internal/config"
	"bui/internal/llm"

	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Mock LLM Provider
// =============================================================================

// mockProvider is a mock LLM provider for testing.
type mockProvider struct {
	name     string
	tokens   []string
	err      error
	streamCh chan llm.Token
	errCh    chan error
}

func newMockProvider(tokens []string, err error) *mockProvider {
	return &mockProvider{
		name:   "mock",
		tokens: tokens,
		err:    err,
	}
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Stream(ctx context.Context, p llm.Prompt) (<-chan llm.Token, <-chan error) {
	tokenCh := make(chan llm.Token, len(m.tokens)+1)
	errCh := make(chan error, 1)

	// Send tokens
	go func() {
		defer close(tokenCh)
		for _, t := range m.tokens {
			select {
			case <-ctx.Done():
				return
			case tokenCh <- llm.Token{Text: t}:
			}
		}
		if m.err != nil {
			errCh <- m.err
		}
	}()

	m.streamCh = tokenCh
	m.errCh = errCh
	return tokenCh, errCh
}

// =============================================================================
// Test Helpers
// =============================================================================

func testConfig() config.Config {
	return config.Config{
		UI: config.UIConfig{
			Theme:  "dark",
			Accent: "indigo",
		},
		LLM: config.LLMConfig{
			Provider:    "mock",
			MaxTokens:   500,
			Temperature: 0.3,
		},
	}
}

func testPrompt() llm.Prompt {
	return llm.Prompt{
		System:      "You are a helpful assistant.",
		User:        "Write a PR description.",
		MaxTokens:   500,
		Temperature: 0.3,
	}
}

// =============================================================================
// Constructor Tests
// =============================================================================

func TestNew_CreatesValidModel(t *testing.T) {
	cfg := testConfig()
	provider := newMockProvider([]string{"Hello"}, nil)
	prompt := testPrompt()
	onSubmit := func(s string) tea.Cmd { return nil }
	onCancel := func() tea.Cmd { return nil }

	m := New(cfg, provider, prompt, onSubmit, onCancel)

	// Check initial state
	if m.state != StateEditing {
		t.Errorf("expected initial state to be StateEditing, got %v", m.state)
	}

	// Check that dimensions are zero (will be set by SetSize)
	if m.width != 0 || m.height != 0 {
		t.Errorf("expected initial dimensions to be 0, got %dx%d", m.width, m.height)
	}

	// Check that streamed text is empty
	if m.streamedText != "" {
		t.Errorf("expected streamedText to be empty, got %q", m.streamedText)
	}

	// Check that prompt is stored
	if m.prompt.System != prompt.System {
		t.Errorf("expected prompt.System to be %q, got %q", prompt.System, m.prompt.System)
	}
}

func TestNew_WithNilProvider(t *testing.T) {
	cfg := testConfig()
	prompt := testPrompt()
	onSubmit := func(s string) tea.Cmd { return nil }
	onCancel := func() tea.Cmd { return nil }

	m := New(cfg, nil, prompt, onSubmit, onCancel)

	// Should still create a valid model
	if m.state != StateEditing {
		t.Errorf("expected initial state to be StateEditing, got %v", m.state)
	}
}

// =============================================================================
// State Tests
// =============================================================================

func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateEditing, "editing"},
		{StateGenerating, "generating"},
		{StateReady, "ready"},
		{StateError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("State.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// SetSize Tests
// =============================================================================

func TestSetSize_UpdatesDimensions(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)

	m.SetSize(120, 40)

	if m.width != 120 {
		t.Errorf("expected width to be 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height to be 40, got %d", m.height)
	}
}

func TestSetSize_UpdatesTextareaSize(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)

	m.SetSize(100, 50)

	// Textarea should be sized appropriately (width - padding, and height calculation)
	// We can't easily verify internal textarea dimensions, but we can ensure no panic
	if m.width != 100 || m.height != 50 {
		t.Errorf("SetSize did not update dimensions correctly")
	}
}

// =============================================================================
// Init Tests
// =============================================================================

func TestInit_ReturnsTextareaBlink(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)

	cmd := m.Init()

	// Init should return a command (textarea blink)
	if cmd == nil {
		t.Error("expected Init() to return a command, got nil")
	}
}

// =============================================================================
// View Tests
// =============================================================================

func TestView_ReturnsNonEmptyString(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)
	m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("expected View() to return non-empty string")
	}
}

func TestView_EditingState(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)
	m.SetSize(80, 24)
	m.state = StateEditing

	view := m.View()

	if view == "" {
		t.Error("expected View() to return non-empty string for StateEditing")
	}
}

func TestView_GeneratingState(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)
	m.SetSize(80, 24)
	m.state = StateGenerating

	view := m.View()

	if view == "" {
		t.Error("expected View() to return non-empty string for StateGenerating")
	}
}

func TestView_ReadyState(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)
	m.SetSize(80, 24)
	m.state = StateReady
	m.streamedText = "Generated content"

	view := m.View()

	if view == "" {
		t.Error("expected View() to return non-empty string for StateReady")
	}
}

func TestView_ErrorState(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)
	m.SetSize(80, 24)
	m.state = StateError
	m.err = errors.New("test error")

	view := m.View()

	if view == "" {
		t.Error("expected View() to return non-empty string for StateError")
	}
}

// =============================================================================
// Update Tests - Key Handling
// =============================================================================

func TestUpdate_EscCancels(t *testing.T) {
	cfg := testConfig()
	cancelCalled := false
	onCancel := func() tea.Cmd {
		cancelCalled = true
		return func() tea.Msg { return ContentCancelledMsg{} }
	}

	m := New(cfg, nil, testPrompt(), nil, onCancel)
	m.SetSize(80, 24)
	m.state = StateEditing

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	_, _ = m.Update(msg)

	if !cancelCalled {
		t.Error("expected Esc to call onCancel callback")
	}
}

func TestUpdate_CtrlGStartsGeneration(t *testing.T) {
	cfg := testConfig()
	provider := newMockProvider([]string{"Hello", " World"}, nil)

	m := New(cfg, provider, testPrompt(), nil, nil)
	m.SetSize(80, 24)
	m.state = StateEditing

	// Simulate Ctrl+G key press
	msg := tea.KeyMsg{Type: tea.KeyCtrlG}
	newModel, cmd := m.Update(msg)
	updatedModel := newModel.(Model)

	// State should transition to generating
	if updatedModel.state != StateGenerating {
		t.Errorf("expected state to be StateGenerating after Ctrl+G, got %v", updatedModel.state)
	}

	// Command should not be nil (starts generation)
	if cmd == nil {
		t.Error("expected Ctrl+G to return a command")
	}
}

func TestUpdate_CtrlGWithoutProvider(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil) // No provider
	m.SetSize(80, 24)
	m.state = StateEditing

	msg := tea.KeyMsg{Type: tea.KeyCtrlG}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	// Should stay in editing state when no provider
	if updatedModel.state != StateEditing {
		t.Errorf("expected state to remain StateEditing without provider, got %v", updatedModel.state)
	}
}

func TestUpdate_CtrlXCancelsGeneration(t *testing.T) {
	cfg := testConfig()
	provider := newMockProvider([]string{"Hello"}, nil)

	m := New(cfg, provider, testPrompt(), nil, nil)
	m.SetSize(80, 24)
	m.state = StateGenerating

	// Set up a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	m.llmCtx = ctx
	m.llmCancel = cancel

	msg := tea.KeyMsg{Type: tea.KeyCtrlX}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	// Should transition back to editing
	if updatedModel.state != StateEditing {
		t.Errorf("expected state to be StateEditing after Ctrl+X, got %v", updatedModel.state)
	}
}

func TestUpdate_CtrlSSubmitsContent(t *testing.T) {
	cfg := testConfig()
	submittedContent := ""
	onSubmit := func(s string) tea.Cmd {
		submittedContent = s
		return func() tea.Msg { return ContentAcceptedMsg{Content: s} }
	}

	m := New(cfg, nil, testPrompt(), onSubmit, nil)
	m.SetSize(80, 24)
	m.state = StateReady
	m.streamedText = "Generated content"

	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	_, _ = m.Update(msg)

	if submittedContent != "Generated content" {
		t.Errorf("expected submitted content to be %q, got %q", "Generated content", submittedContent)
	}
}

// =============================================================================
// Update Tests - Message Handling
// =============================================================================

func TestUpdate_WindowSizeMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.width != 100 {
		t.Errorf("expected width to be 100, got %d", updatedModel.width)
	}
	if updatedModel.height != 50 {
		t.Errorf("expected height to be 50, got %d", updatedModel.height)
	}
}

func TestUpdate_GenerateTokenMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)
	m.SetSize(80, 24)
	m.state = StateGenerating

	// Set up channels for token reading
	tokenCh := make(chan llm.Token, 10)
	errCh := make(chan error, 1)
	m.tokenCh = tokenCh
	m.errCh = errCh

	// Send a token message
	msg := GenerateTokenMsg{Token: "Hello"}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.streamedText != "Hello" {
		t.Errorf("expected streamedText to be %q, got %q", "Hello", updatedModel.streamedText)
	}
}

func TestUpdate_GenerateDoneMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)
	m.SetSize(80, 24)
	m.state = StateGenerating
	m.streamedText = "Generated content"

	msg := GenerateDoneMsg{}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateReady {
		t.Errorf("expected state to be StateReady after GenerateDoneMsg, got %v", updatedModel.state)
	}
}

func TestUpdate_GenerateErrorMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)
	m.SetSize(80, 24)
	m.state = StateGenerating

	testErr := errors.New("LLM error")
	msg := GenerateErrorMsg{Err: testErr}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateError {
		t.Errorf("expected state to be StateError after GenerateErrorMsg, got %v", updatedModel.state)
	}
	if updatedModel.err != testErr {
		t.Errorf("expected err to be %v, got %v", testErr, updatedModel.err)
	}
}

// =============================================================================
// State Transition Tests
// =============================================================================

func TestStateTransition_EditingToGenerating(t *testing.T) {
	cfg := testConfig()
	provider := newMockProvider([]string{"Hello"}, nil)
	m := New(cfg, provider, testPrompt(), nil, nil)
	m.SetSize(80, 24)

	// Start generation
	msg := tea.KeyMsg{Type: tea.KeyCtrlG}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateGenerating {
		t.Errorf("expected transition from StateEditing to StateGenerating")
	}
}

func TestStateTransition_GeneratingToReady(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)
	m.SetSize(80, 24)
	m.state = StateGenerating

	msg := GenerateDoneMsg{}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateReady {
		t.Errorf("expected transition from StateGenerating to StateReady")
	}
}

func TestStateTransition_GeneratingToError(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)
	m.SetSize(80, 24)
	m.state = StateGenerating

	msg := GenerateErrorMsg{Err: errors.New("error")}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateError {
		t.Errorf("expected transition from StateGenerating to StateError")
	}
}

// =============================================================================
// Content Tests
// =============================================================================

func TestGetContent_ReturnsStreamedText(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)
	m.streamedText = "Test content"

	content := m.GetContent()
	if content != "Test content" {
		t.Errorf("expected GetContent() to return %q, got %q", "Test content", content)
	}
}

func TestSetContent_UpdatesTextarea(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)
	m.SetSize(80, 24)

	m.SetContent("New content")

	// The textarea should now have the content
	// We verify by checking that GetContent returns the correct value
	if m.GetContent() != "New content" {
		t.Errorf("expected SetContent() to update content, got %q", m.GetContent())
	}
}

// =============================================================================
// Accessor Tests
// =============================================================================

func TestState_ReturnsCurrentState(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)

	if m.State() != StateEditing {
		t.Errorf("expected State() to return StateEditing, got %v", m.State())
	}

	m.state = StateGenerating
	if m.State() != StateGenerating {
		t.Errorf("expected State() to return StateGenerating, got %v", m.State())
	}
}

func TestError_ReturnsCurrentError(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)

	if m.Error() != nil {
		t.Errorf("expected Error() to return nil, got %v", m.Error())
	}

	testErr := errors.New("test error")
	m.err = testErr
	if m.Error() != testErr {
		t.Errorf("expected Error() to return %v, got %v", testErr, m.Error())
	}
}

func TestStreamedText_ReturnsCurrentStreamedText(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, nil, testPrompt(), nil, nil)

	if m.StreamedText() != "" {
		t.Errorf("expected StreamedText() to return empty string, got %q", m.StreamedText())
	}

	m.streamedText = "Hello World"
	if m.StreamedText() != "Hello World" {
		t.Errorf("expected StreamedText() to return %q, got %q", "Hello World", m.StreamedText())
	}
}
