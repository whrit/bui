package createpr

import (
	"testing"

	"bui/internal/config"
	"bui/internal/gh"
	"bui/internal/git"

	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Mock Implementations
// =============================================================================

// mockGitExecutor implements git.CommandExecutor for testing.
type mockGitExecutor struct {
	responses map[string][]byte
	errors    map[string]error
}

func newMockGitExecutor() *mockGitExecutor {
	return &mockGitExecutor{
		responses: make(map[string][]byte),
		errors:    make(map[string]error),
	}
}

func (m *mockGitExecutor) Run(cmd string, args ...string) ([]byte, error) {
	// Build a key from cmd + first arg for simple lookup
	key := cmd
	if len(args) > 0 {
		key = cmd + " " + args[0]
	}

	if err, ok := m.errors[key]; ok {
		return nil, err
	}

	if resp, ok := m.responses[key]; ok {
		return resp, nil
	}

	// Default responses
	switch {
	case len(args) > 0 && args[0] == "rev-parse":
		if len(args) > 1 && args[1] == "--abbrev-ref" {
			return []byte("feature/test-branch\n"), nil
		}
	case len(args) > 0 && args[0] == "branch":
		return []byte("*\x1fmain\x1forigin/main\x1fabc123\n \x1ffeature/test-branch\x1f\x1fdef456\n"), nil
	case len(args) > 0 && args[0] == "log":
		return []byte("abc123\x1fabc1234\x1fAuthor\x1fauthor@test.com\x1f1700000000\x1ffeat: add feature\x1fsome body\x1e"), nil
	}

	return []byte{}, nil
}

func (m *mockGitExecutor) RunInDir(dir, cmd string, args ...string) ([]byte, error) {
	return m.Run(cmd, args...)
}

// mockGHExecutor implements gh.CommandExecutor for testing.
type mockGHExecutor struct {
	responses map[string][]byte
	errors    map[string]error
}

func newMockGHExecutor() *mockGHExecutor {
	return &mockGHExecutor{
		responses: make(map[string][]byte),
		errors:    make(map[string]error),
	}
}

func (m *mockGHExecutor) Run(cmd string, args ...string) ([]byte, error) {
	key := cmd
	if len(args) > 0 {
		key = cmd + " " + args[0]
	}

	if err, ok := m.errors[key]; ok {
		return nil, err
	}

	if resp, ok := m.responses[key]; ok {
		return resp, nil
	}

	return []byte{}, nil
}

// testConfig returns a minimal config for testing.
func testConfig() config.Config {
	return config.Config{
		UI: config.UIConfig{
			Theme:  "dark",
			Accent: "indigo",
		},
		Keys: config.KeysConfig{
			Quit:    "q",
			Help:    "?",
			Confirm: "enter",
		},
		Git: config.GitConfig{
			DefaultBase: "main",
		},
	}
}

// =============================================================================
// Constructor Tests
// =============================================================================

func TestNew(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	// Create model without LLM provider (nil)
	model := New(cfg, ghClient, gitRepo, nil)

	// Verify initial state
	if model.step != StepBranchSelect {
		t.Errorf("expected step %v, got %v", StepBranchSelect, model.step)
	}

	if model.ghClient == nil {
		t.Error("expected ghClient to be set")
	}

	if model.gitRepo == nil {
		t.Error("expected gitRepo to be set")
	}

	// llmProvider should be nil
	if model.llmProvider != nil {
		t.Error("expected llmProvider to be nil")
	}

	// isDraft should default to false
	if model.isDraft != false {
		t.Error("expected isDraft to be false")
	}
}

func TestNew_WithNilGitRepo(t *testing.T) {
	cfg := testConfig()
	ghExec := newMockGHExecutor()
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, nil, nil)

	if model.gitRepo != nil {
		t.Error("expected gitRepo to be nil when passed nil")
	}
}

// =============================================================================
// Step Enum Tests
// =============================================================================

func TestStepString(t *testing.T) {
	tests := []struct {
		step     Step
		expected string
	}{
		{StepBranchSelect, "branch_select"},
		{StepDraftGeneration, "draft_generation"},
		{StepEditTitle, "edit_title"},
		{StepEditBody, "edit_body"},
		{StepReview, "review"},
		{StepSubmitting, "submitting"},
		{StepComplete, "complete"},
		{StepError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.step.String(); got != tt.expected {
				t.Errorf("Step(%d).String() = %q, want %q", tt.step, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Init Tests
// =============================================================================

func TestInit(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)
	cmd := model.Init()

	// Init should return a command
	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

// =============================================================================
// Step Transition Tests
// =============================================================================

func TestStepTransitions(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	t.Run("branch select to draft generation", func(t *testing.T) {
		model := New(cfg, ghClient, gitRepo, nil)
		model.step = StepBranchSelect
		model.baseBranch = "main"
		model.headBranch = "feature/test"
		model.branches = []string{"main", "feature/test"}

		// Without LLM provider, should skip draft generation
		model.step = model.nextStep()

		// When llmProvider is nil, should skip to StepEditTitle
		if model.step != StepEditTitle {
			t.Errorf("expected step to be StepEditTitle when llmProvider is nil, got %v", model.step)
		}
	})

	t.Run("edit title to edit body", func(t *testing.T) {
		model := New(cfg, ghClient, gitRepo, nil)
		model.step = StepEditTitle
		model.title = "Test PR Title"

		model.step = model.nextStep()

		if model.step != StepEditBody {
			t.Errorf("expected step StepEditBody, got %v", model.step)
		}
	})

	t.Run("edit body to review", func(t *testing.T) {
		model := New(cfg, ghClient, gitRepo, nil)
		model.step = StepEditBody
		model.body = "Test PR Body"

		model.step = model.nextStep()

		if model.step != StepReview {
			t.Errorf("expected step StepReview, got %v", model.step)
		}
	})

	t.Run("review to submitting", func(t *testing.T) {
		model := New(cfg, ghClient, gitRepo, nil)
		model.step = StepReview

		model.step = model.nextStep()

		if model.step != StepSubmitting {
			t.Errorf("expected step StepSubmitting, got %v", model.step)
		}
	})
}

func TestPrevStep(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	t.Run("edit body to edit title", func(t *testing.T) {
		model := New(cfg, ghClient, gitRepo, nil)
		model.step = StepEditBody

		model.step = model.prevStep()

		if model.step != StepEditTitle {
			t.Errorf("expected step StepEditTitle, got %v", model.step)
		}
	})

	t.Run("review to edit body", func(t *testing.T) {
		model := New(cfg, ghClient, gitRepo, nil)
		model.step = StepReview

		model.step = model.prevStep()

		if model.step != StepEditBody {
			t.Errorf("expected step StepEditBody, got %v", model.step)
		}
	})

	t.Run("branch select cannot go back", func(t *testing.T) {
		model := New(cfg, ghClient, gitRepo, nil)
		model.step = StepBranchSelect

		model.step = model.prevStep()

		// Should stay at branch select
		if model.step != StepBranchSelect {
			t.Errorf("expected step StepBranchSelect, got %v", model.step)
		}
	})
}

// =============================================================================
// Message Handling Tests
// =============================================================================

func TestUpdate_BranchesLoaded(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)

	msg := BranchesLoadedMsg{
		Branches:      []string{"main", "develop", "feature/test"},
		CurrentBranch: "feature/test",
		DefaultBranch: "main",
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if len(m.branches) != 3 {
		t.Errorf("expected 3 branches, got %d", len(m.branches))
	}

	if m.headBranch != "feature/test" {
		t.Errorf("expected head branch to be feature/test, got %s", m.headBranch)
	}

	if m.baseBranch != "main" {
		t.Errorf("expected base branch to be main, got %s", m.baseBranch)
	}
}

func TestUpdate_BranchesLoadError(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)

	msg := BranchesLoadErrorMsg{
		Err: &testError{"failed to load branches"},
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.step != StepError {
		t.Errorf("expected step StepError, got %v", m.step)
	}

	if m.err == nil {
		t.Error("expected err to be set")
	}
}

func TestUpdate_PRCreated(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)
	model.step = StepSubmitting

	pr := &gh.PR{
		Number: 123,
		Title:  "Test PR",
		URL:    "https://github.com/test/repo/pull/123",
	}

	msg := PRCreatedMsg{PR: pr}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.step != StepComplete {
		t.Errorf("expected step StepComplete, got %v", m.step)
	}

	if m.createdPR == nil {
		t.Error("expected createdPR to be set")
	}

	if m.createdPR.Number != 123 {
		t.Errorf("expected PR number 123, got %d", m.createdPR.Number)
	}
}

func TestUpdate_PRCreateError(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)
	model.step = StepSubmitting

	msg := PRCreateErrorMsg{
		Err: &testError{"failed to create PR"},
	}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.step != StepError {
		t.Errorf("expected step StepError, got %v", m.step)
	}
}

func TestUpdate_CommitsLoaded(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)

	commits := []git.Commit{
		{ShortHash: "abc1234", Subject: "feat: add feature"},
		{ShortHash: "def5678", Subject: "fix: bug fix"},
	}

	msg := CommitsLoadedMsg{Commits: commits}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if len(m.commits) != 2 {
		t.Errorf("expected 2 commits, got %d", len(m.commits))
	}
}

// =============================================================================
// Key Handling Tests
// =============================================================================

func TestUpdate_KeyEscape_BranchSelect(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)
	model.step = StepBranchSelect

	msg := tea.KeyMsg{Type: tea.KeyEscape}

	_, cmd := model.Update(msg)

	// Escape on branch select should return BackToDashboardMsg
	if cmd == nil {
		t.Error("expected command to be returned for escape key")
	}
}

func TestUpdate_KeyEscape_EditTitle(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)
	model.step = StepEditTitle

	msg := tea.KeyMsg{Type: tea.KeyEscape}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Escape during edit should go back
	if m.step != StepBranchSelect {
		t.Errorf("expected step StepBranchSelect after escape, got %v", m.step)
	}
}

func TestUpdate_KeyTab_EditTitle(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)
	model.step = StepEditTitle
	model.title = "Test Title"

	msg := tea.KeyMsg{Type: tea.KeyTab}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// Tab in edit title should go to edit body
	if m.step != StepEditBody {
		t.Errorf("expected step StepEditBody after tab, got %v", m.step)
	}
}

func TestUpdate_KeyD_Review(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)
	model.step = StepReview
	model.isDraft = false

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	// 'd' in review should toggle draft mode
	if m.isDraft != true {
		t.Error("expected isDraft to be toggled to true")
	}
}

// =============================================================================
// View Tests
// =============================================================================

func TestView_BranchSelect(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)
	model.step = StepBranchSelect
	model.branches = []string{"main", "develop", "feature/test"}
	model.baseBranch = "main"
	model.headBranch = "feature/test"
	model.width = 80
	model.height = 24

	view := model.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// View should contain step indicator
	if !containsString(view, "Step 1") {
		t.Error("expected view to contain 'Step 1'")
	}
}

func TestView_EditTitle(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)
	model.step = StepEditTitle
	model.title = "Test Title"
	model.width = 80
	model.height = 24

	view := model.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// View should contain title editing indicator
	if !containsString(view, "Title") {
		t.Error("expected view to contain 'Title'")
	}
}

func TestView_Review(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)
	model.step = StepReview
	model.title = "Test PR Title"
	model.body = "Test PR Body"
	model.baseBranch = "main"
	model.headBranch = "feature/test"
	model.width = 80
	model.height = 24

	view := model.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// View should contain review content
	if !containsString(view, "Review") {
		t.Error("expected view to contain 'Review'")
	}
}

func TestView_Complete(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)
	model.step = StepComplete
	model.createdPR = &gh.PR{
		Number: 123,
		Title:  "Test PR",
		URL:    "https://github.com/test/repo/pull/123",
	}
	model.width = 80
	model.height = 24

	view := model.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// View should contain success message or PR URL
	if !containsString(view, "123") && !containsString(view, "Created") {
		t.Error("expected view to contain PR number or success indicator")
	}
}

func TestView_Error(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)
	model.step = StepError
	model.err = &testError{"something went wrong"}
	model.width = 80
	model.height = 24

	view := model.View()

	if view == "" {
		t.Error("expected non-empty view")
	}

	// View should contain error message
	if !containsString(view, "Error") && !containsString(view, "wrong") {
		t.Error("expected view to contain error indicator")
	}
}

// =============================================================================
// Accessor Tests
// =============================================================================

func TestAccessors(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)
	model.step = StepEditTitle
	model.title = "Test Title"
	model.body = "Test Body"
	model.baseBranch = "main"
	model.headBranch = "feature/test"
	model.isDraft = true
	model.err = &testError{"test error"}

	if model.Step() != StepEditTitle {
		t.Errorf("expected Step() to return StepEditTitle, got %v", model.Step())
	}

	if model.Title() != "Test Title" {
		t.Errorf("expected Title() to return 'Test Title', got %s", model.Title())
	}

	if model.Body() != "Test Body" {
		t.Errorf("expected Body() to return 'Test Body', got %s", model.Body())
	}

	if model.BaseBranch() != "main" {
		t.Errorf("expected BaseBranch() to return 'main', got %s", model.BaseBranch())
	}

	if model.HeadBranch() != "feature/test" {
		t.Errorf("expected HeadBranch() to return 'feature/test', got %s", model.HeadBranch())
	}

	if model.IsDraft() != true {
		t.Error("expected IsDraft() to return true")
	}

	if model.Error() == nil {
		t.Error("expected Error() to return non-nil error")
	}

	if model.HasLLMProvider() {
		t.Error("expected HasLLMProvider() to return false when provider is nil")
	}
}

// =============================================================================
// Draft Mode Toggle Tests
// =============================================================================

func TestToggleDraft(t *testing.T) {
	cfg := testConfig()
	gitExec := newMockGitExecutor()
	ghExec := newMockGHExecutor()

	gitRepo := git.NewWithExecutor(gitExec)
	ghClient := gh.NewWithExecutor(ghExec)

	model := New(cfg, ghClient, gitRepo, nil)

	if model.isDraft != false {
		t.Error("expected initial isDraft to be false")
	}

	model.toggleDraft()
	if model.isDraft != true {
		t.Error("expected isDraft to be true after first toggle")
	}

	model.toggleDraft()
	if model.isDraft != false {
		t.Error("expected isDraft to be false after second toggle")
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// testError implements error interface for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// containsString checks if a string contains a substring.
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
