package review

import (
	"context"
	"errors"
	"testing"

	"bui/internal/config"
	"bui/internal/gh"
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
// Mock GH Client
// =============================================================================

type mockExecutor struct {
	output []byte
	err    error
}

func (m *mockExecutor) Run(_ string, _ ...string) ([]byte, error) {
	return m.output, m.err
}

func newMockGHClient() *gh.Client {
	return gh.NewWithExecutor(&mockExecutor{output: []byte("{}"), err: nil})
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

// =============================================================================
// Constructor Tests
// =============================================================================

func TestNew_CreatesValidModel(t *testing.T) {
	cfg := testConfig()
	ghClient := newMockGHClient()
	provider := newMockProvider([]string{"Hello"}, nil)

	m := New(cfg, ghClient, provider, 42, "Test PR Title", "test diff content")

	// Check initial state
	if m.state != StateTypeSelect {
		t.Errorf("expected initial state to be StateTypeSelect, got %v", m.state)
	}

	// Check PR context
	if m.prNumber != 42 {
		t.Errorf("expected prNumber to be 42, got %d", m.prNumber)
	}
	if m.prTitle != "Test PR Title" {
		t.Errorf("expected prTitle to be %q, got %q", "Test PR Title", m.prTitle)
	}
	if m.diff != "test diff content" {
		t.Errorf("expected diff to be %q, got %q", "test diff content", m.diff)
	}

	// Check that comment is empty
	if m.comment != "" {
		t.Errorf("expected comment to be empty, got %q", m.comment)
	}

	// Check default review type
	if m.reviewType != ReviewTypeApprove {
		t.Errorf("expected default reviewType to be ReviewTypeApprove, got %v", m.reviewType)
	}
}

func TestNew_WithNilProvider(t *testing.T) {
	cfg := testConfig()
	ghClient := newMockGHClient()

	m := New(cfg, ghClient, nil, 42, "Test PR", "diff")

	// Should still create a valid model
	if m.state != StateTypeSelect {
		t.Errorf("expected initial state to be StateTypeSelect, got %v", m.state)
	}

	// HasProvider should return false
	if m.HasProvider() {
		t.Error("expected HasProvider() to return false when provider is nil")
	}
}

func TestNew_WithNilGHClient(t *testing.T) {
	cfg := testConfig()
	provider := newMockProvider([]string{"Hello"}, nil)

	m := New(cfg, nil, provider, 42, "Test PR", "diff")

	// Should still create a valid model
	if m.state != StateTypeSelect {
		t.Errorf("expected initial state to be StateTypeSelect, got %v", m.state)
	}
}

// =============================================================================
// State Enum Tests
// =============================================================================

func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateTypeSelect, "type_select"},
		{StateCompose, "compose"},
		{StatePreview, "preview"},
		{StateSubmitting, "submitting"},
		{StateComplete, "complete"},
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
// ReviewType Enum Tests
// =============================================================================

func TestReviewType_String(t *testing.T) {
	tests := []struct {
		rt   ReviewType
		want string
	}{
		{ReviewTypeApprove, "approve"},
		{ReviewTypeRequestChanges, "request_changes"},
		{ReviewTypeComment, "comment"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.rt.String()
			if got != tt.want {
				t.Errorf("ReviewType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReviewType_DisplayName(t *testing.T) {
	tests := []struct {
		rt   ReviewType
		want string
	}{
		{ReviewTypeApprove, "Approve"},
		{ReviewTypeRequestChanges, "Request Changes"},
		{ReviewTypeComment, "Comment"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.rt.DisplayName()
			if got != tt.want {
				t.Errorf("ReviewType.DisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestReviewType_ToGHReviewEvent(t *testing.T) {
	tests := []struct {
		rt   ReviewType
		want gh.ReviewEvent
	}{
		{ReviewTypeApprove, gh.ReviewApprove},
		{ReviewTypeRequestChanges, gh.ReviewRequestChanges},
		{ReviewTypeComment, gh.ReviewComment},
	}

	for _, tt := range tests {
		t.Run(tt.rt.String(), func(t *testing.T) {
			got := tt.rt.ToGHReviewEvent()
			if got != tt.want {
				t.Errorf("ReviewType.ToGHReviewEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// SetSize Tests
// =============================================================================

func TestSetSize_UpdatesDimensions(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")

	m.SetSize(120, 40)

	if m.width != 120 {
		t.Errorf("expected width to be 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height to be 40, got %d", m.height)
	}
}

// =============================================================================
// Init Tests
// =============================================================================

func TestInit_ReturnsCommand(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")

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
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)

	view := m.View()

	if view == "" {
		t.Error("expected View() to return non-empty string")
	}
}

func TestView_TypeSelectState(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateTypeSelect

	view := m.View()

	if view == "" {
		t.Error("expected View() to return non-empty string for StateTypeSelect")
	}
}

func TestView_ComposeState(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateCompose

	view := m.View()

	if view == "" {
		t.Error("expected View() to return non-empty string for StateCompose")
	}
}

func TestView_PreviewState(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StatePreview

	view := m.View()

	if view == "" {
		t.Error("expected View() to return non-empty string for StatePreview")
	}
}

func TestView_SubmittingState(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateSubmitting

	view := m.View()

	if view == "" {
		t.Error("expected View() to return non-empty string for StateSubmitting")
	}
}

func TestView_CompleteState(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateComplete

	view := m.View()

	if view == "" {
		t.Error("expected View() to return non-empty string for StateComplete")
	}
}

func TestView_ErrorState(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateError
	m.err = errors.New("test error")

	view := m.View()

	if view == "" {
		t.Error("expected View() to return non-empty string for StateError")
	}
}

// =============================================================================
// Update Tests - Key Handling in TypeSelect State
// =============================================================================

func TestUpdate_TypeSelect_Key1SelectsApprove(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateTypeSelect

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.reviewType != ReviewTypeApprove {
		t.Errorf("expected reviewType to be ReviewTypeApprove, got %v", updatedModel.reviewType)
	}
	if updatedModel.state != StateCompose {
		t.Errorf("expected state to transition to StateCompose, got %v", updatedModel.state)
	}
}

func TestUpdate_TypeSelect_KeyASelectsApprove(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateTypeSelect

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.reviewType != ReviewTypeApprove {
		t.Errorf("expected reviewType to be ReviewTypeApprove, got %v", updatedModel.reviewType)
	}
	if updatedModel.state != StateCompose {
		t.Errorf("expected state to transition to StateCompose, got %v", updatedModel.state)
	}
}

func TestUpdate_TypeSelect_Key2SelectsRequestChanges(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateTypeSelect

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.reviewType != ReviewTypeRequestChanges {
		t.Errorf("expected reviewType to be ReviewTypeRequestChanges, got %v", updatedModel.reviewType)
	}
	if updatedModel.state != StateCompose {
		t.Errorf("expected state to transition to StateCompose, got %v", updatedModel.state)
	}
}

func TestUpdate_TypeSelect_KeyCSelectsRequestChanges(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateTypeSelect

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.reviewType != ReviewTypeRequestChanges {
		t.Errorf("expected reviewType to be ReviewTypeRequestChanges, got %v", updatedModel.reviewType)
	}
}

func TestUpdate_TypeSelect_Key3SelectsComment(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateTypeSelect

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.reviewType != ReviewTypeComment {
		t.Errorf("expected reviewType to be ReviewTypeComment, got %v", updatedModel.reviewType)
	}
	if updatedModel.state != StateCompose {
		t.Errorf("expected state to transition to StateCompose, got %v", updatedModel.state)
	}
}

func TestUpdate_TypeSelect_KeyMSelectsComment(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateTypeSelect

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.reviewType != ReviewTypeComment {
		t.Errorf("expected reviewType to be ReviewTypeComment, got %v", updatedModel.reviewType)
	}
}

func TestUpdate_TypeSelect_EscCancels(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateTypeSelect

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	_, cmd := m.Update(msg)

	// Should return a back command
	if cmd == nil {
		t.Error("expected Esc to return a command")
	}
}

// =============================================================================
// Update Tests - Key Handling in Compose State
// =============================================================================

func TestUpdate_Compose_CtrlGStartsGeneration(t *testing.T) {
	cfg := testConfig()
	provider := newMockProvider([]string{"Hello", " World"}, nil)
	m := New(cfg, newMockGHClient(), provider, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateCompose

	msg := tea.KeyMsg{Type: tea.KeyCtrlG}
	newModel, cmd := m.Update(msg)
	updatedModel := newModel.(Model)

	if !updatedModel.isGenerating {
		t.Error("expected isGenerating to be true after Ctrl+G")
	}
	if cmd == nil {
		t.Error("expected Ctrl+G to return a command")
	}
}

func TestUpdate_Compose_CtrlGWithoutProvider(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateCompose

	msg := tea.KeyMsg{Type: tea.KeyCtrlG}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	// Should not start generating without provider
	if updatedModel.isGenerating {
		t.Error("expected isGenerating to be false without provider")
	}
}

func TestUpdate_Compose_CtrlSGoesToPreview(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateCompose
	m.reviewType = ReviewTypeApprove

	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StatePreview {
		t.Errorf("expected state to transition to StatePreview, got %v", updatedModel.state)
	}
}

func TestUpdate_Compose_RequestChangesRequiresComment(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateCompose
	m.reviewType = ReviewTypeRequestChanges
	m.comment = "" // Empty comment

	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	// Should stay in compose state if comment is empty for request changes
	if updatedModel.state != StateCompose {
		t.Errorf("expected state to remain StateCompose for empty request changes comment, got %v", updatedModel.state)
	}
}

func TestUpdate_Compose_RequestChangesWithComment(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateCompose
	m.reviewType = ReviewTypeRequestChanges
	// Set both the textarea value and the comment field
	m.textarea.SetValue("Please fix this issue")
	m.comment = "Please fix this issue"

	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	// Should proceed to preview with non-empty comment
	if updatedModel.state != StatePreview {
		t.Errorf("expected state to transition to StatePreview, got %v", updatedModel.state)
	}
}

func TestUpdate_Compose_EscGoesBack(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StateCompose

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateTypeSelect {
		t.Errorf("expected state to go back to StateTypeSelect, got %v", updatedModel.state)
	}
}

// =============================================================================
// Update Tests - Key Handling in Preview State
// =============================================================================

func TestUpdate_Preview_EnterSubmits(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StatePreview

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateSubmitting {
		t.Errorf("expected state to transition to StateSubmitting, got %v", updatedModel.state)
	}
	if cmd == nil {
		t.Error("expected Enter to return a submit command")
	}
}

func TestUpdate_Preview_KeyEGoesToEdit(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StatePreview

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateCompose {
		t.Errorf("expected state to go back to StateCompose, got %v", updatedModel.state)
	}
}

func TestUpdate_Preview_EscGoesBack(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test PR", "diff")
	m.SetSize(80, 24)
	m.state = StatePreview

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateCompose {
		t.Errorf("expected state to go back to StateCompose, got %v", updatedModel.state)
	}
}

// =============================================================================
// Update Tests - Message Handling
// =============================================================================

func TestUpdate_WindowSizeMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")

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

func TestUpdate_GenerateCommentTokenMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.SetSize(80, 24)
	m.state = StateCompose
	m.isGenerating = true

	// Set up channels
	tokenCh := make(chan llm.Token, 10)
	errCh := make(chan error, 1)
	m.tokenCh = tokenCh
	m.errCh = errCh

	msg := GenerateCommentTokenMsg{Token: "Hello"}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.generatedComment != "Hello" {
		t.Errorf("expected generatedComment to be %q, got %q", "Hello", updatedModel.generatedComment)
	}
}

func TestUpdate_GenerateCommentDoneMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.SetSize(80, 24)
	m.state = StateCompose
	m.isGenerating = true
	m.generatedComment = "Generated comment"

	msg := GenerateCommentDoneMsg{}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.isGenerating {
		t.Error("expected isGenerating to be false after GenerateCommentDoneMsg")
	}
}

func TestUpdate_GenerateCommentErrorMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.SetSize(80, 24)
	m.state = StateCompose
	m.isGenerating = true

	testErr := errors.New("LLM error")
	msg := GenerateCommentErrorMsg{Err: testErr}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.isGenerating {
		t.Error("expected isGenerating to be false after error")
	}
	if updatedModel.err != testErr {
		t.Errorf("expected err to be %v, got %v", testErr, updatedModel.err)
	}
}

func TestUpdate_ReviewSubmittedMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.SetSize(80, 24)
	m.state = StateSubmitting

	msg := ReviewSubmittedMsg{}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateComplete {
		t.Errorf("expected state to be StateComplete, got %v", updatedModel.state)
	}
}

func TestUpdate_ReviewSubmitErrorMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.SetSize(80, 24)
	m.state = StateSubmitting

	testErr := errors.New("submit error")
	msg := ReviewSubmitErrorMsg{Err: testErr}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateError {
		t.Errorf("expected state to be StateError, got %v", updatedModel.state)
	}
	if updatedModel.err != testErr {
		t.Errorf("expected err to be %v, got %v", testErr, updatedModel.err)
	}
}

// =============================================================================
// State Transition Tests
// =============================================================================

func TestStateTransition_TypeSelectToCompose(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.SetSize(80, 24)
	m.state = StateTypeSelect

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateCompose {
		t.Error("expected transition from StateTypeSelect to StateCompose")
	}
}

func TestStateTransition_ComposeToPreview(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.SetSize(80, 24)
	m.state = StateCompose
	m.reviewType = ReviewTypeApprove

	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StatePreview {
		t.Error("expected transition from StateCompose to StatePreview")
	}
}

func TestStateTransition_PreviewToSubmitting(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.SetSize(80, 24)
	m.state = StatePreview

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateSubmitting {
		t.Error("expected transition from StatePreview to StateSubmitting")
	}
}

func TestStateTransition_SubmittingToComplete(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.SetSize(80, 24)
	m.state = StateSubmitting

	msg := ReviewSubmittedMsg{}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateComplete {
		t.Error("expected transition from StateSubmitting to StateComplete")
	}
}

func TestStateTransition_SubmittingToError(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.SetSize(80, 24)
	m.state = StateSubmitting

	msg := ReviewSubmitErrorMsg{Err: errors.New("error")}
	newModel, _ := m.Update(msg)
	updatedModel := newModel.(Model)

	if updatedModel.state != StateError {
		t.Error("expected transition from StateSubmitting to StateError")
	}
}

// =============================================================================
// Accessor Tests
// =============================================================================

func TestState_ReturnsCurrentState(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")

	if m.State() != StateTypeSelect {
		t.Errorf("expected State() to return StateTypeSelect, got %v", m.State())
	}

	m.state = StateCompose
	if m.State() != StateCompose {
		t.Errorf("expected State() to return StateCompose, got %v", m.State())
	}
}

func TestError_ReturnsCurrentError(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")

	if m.Error() != nil {
		t.Errorf("expected Error() to return nil, got %v", m.Error())
	}

	testErr := errors.New("test error")
	m.err = testErr
	if m.Error() != testErr {
		t.Errorf("expected Error() to return %v, got %v", testErr, m.Error())
	}
}

func TestHasProvider(t *testing.T) {
	cfg := testConfig()

	// With provider
	m := New(cfg, newMockGHClient(), newMockProvider(nil, nil), 42, "Test", "diff")
	if !m.HasProvider() {
		t.Error("expected HasProvider() to return true")
	}

	// Without provider
	m = New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	if m.HasProvider() {
		t.Error("expected HasProvider() to return false")
	}
}

func TestReviewType_ReturnsCurrentType(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")

	if m.ReviewType() != ReviewTypeApprove {
		t.Errorf("expected ReviewType() to return ReviewTypeApprove, got %v", m.ReviewType())
	}

	m.reviewType = ReviewTypeRequestChanges
	if m.ReviewType() != ReviewTypeRequestChanges {
		t.Errorf("expected ReviewType() to return ReviewTypeRequestChanges, got %v", m.ReviewType())
	}
}

func TestPRNumber(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")

	if m.PRNumber() != 42 {
		t.Errorf("expected PRNumber() to return 42, got %d", m.PRNumber())
	}
}

func TestPRTitle(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test Title", "diff")

	if m.PRTitle() != "Test Title" {
		t.Errorf("expected PRTitle() to return %q, got %q", "Test Title", m.PRTitle())
	}
}

// =============================================================================
// Validation Tests
// =============================================================================

func TestCanProceedToPreview_ApproveWithoutComment(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.reviewType = ReviewTypeApprove
	m.comment = ""

	// Approve without comment should be allowed
	if !m.canProceedToPreview() {
		t.Error("expected canProceedToPreview() to return true for approve without comment")
	}
}

func TestCanProceedToPreview_RequestChangesWithoutComment(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.reviewType = ReviewTypeRequestChanges
	m.comment = ""

	// Request changes requires a comment
	if m.canProceedToPreview() {
		t.Error("expected canProceedToPreview() to return false for request changes without comment")
	}
}

func TestCanProceedToPreview_RequestChangesWithComment(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.reviewType = ReviewTypeRequestChanges
	m.comment = "Please fix this"

	if !m.canProceedToPreview() {
		t.Error("expected canProceedToPreview() to return true for request changes with comment")
	}
}

func TestCanProceedToPreview_CommentWithoutText(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.reviewType = ReviewTypeComment
	m.comment = ""

	// Comment type requires some text
	if m.canProceedToPreview() {
		t.Error("expected canProceedToPreview() to return false for empty comment")
	}
}

func TestCanProceedToPreview_CommentWithText(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, newMockGHClient(), nil, 42, "Test", "diff")
	m.reviewType = ReviewTypeComment
	m.comment = "Nice work!"

	if !m.canProceedToPreview() {
		t.Error("expected canProceedToPreview() to return true for comment with text")
	}
}
