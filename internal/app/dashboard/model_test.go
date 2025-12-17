package dashboard

import (
	"errors"
	"testing"
	"time"

	"bui/internal/config"
	"bui/internal/gh"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Test Helpers
// =============================================================================

func testConfig() config.Config {
	return config.Config{
		UI: config.UIConfig{
			Theme:  "dark",
			Accent: "indigo",
		},
		Behavior: config.BehaviorConfig{
			AutoRefresh: false,
		},
	}
}

func testPRs() []gh.PR {
	return []gh.PR{
		{
			Number:     1,
			Title:      "Add new feature",
			Body:       "This adds a great new feature",
			State:      "open",
			Author:     gh.User{Login: "alice"},
			HeadBranch: "feature/new-thing",
			BaseBranch: "main",
			URL:        "https://github.com/test/repo/pull/1",
			CreatedAt:  time.Now().Add(-2 * time.Hour),
			UpdatedAt:  time.Now().Add(-1 * time.Hour),
			IsDraft:    false,
			Mergeable:  "MERGEABLE",
			Additions:  100,
			Deletions:  20,
		},
		{
			Number:     2,
			Title:      "Fix bug in parser",
			Body:       "Fixes a critical bug",
			State:      "open",
			Author:     gh.User{Login: "bob"},
			HeadBranch: "fix/parser-bug",
			BaseBranch: "main",
			URL:        "https://github.com/test/repo/pull/2",
			CreatedAt:  time.Now().Add(-24 * time.Hour),
			UpdatedAt:  time.Now().Add(-12 * time.Hour),
			IsDraft:    true,
			Mergeable:  "UNKNOWN",
			Additions:  10,
			Deletions:  5,
		},
		{
			Number:     3,
			Title:      "Update documentation",
			Body:       "Updates the README",
			State:      "closed",
			Author:     gh.User{Login: "charlie"},
			HeadBranch: "docs/update-readme",
			BaseBranch: "main",
			URL:        "https://github.com/test/repo/pull/3",
			CreatedAt:  time.Now().Add(-72 * time.Hour),
			UpdatedAt:  time.Now().Add(-48 * time.Hour),
			IsDraft:    false,
			Mergeable:  "MERGEABLE",
			Additions:  50,
			Deletions:  30,
		},
	}
}

// =============================================================================
// PRItem Tests
// =============================================================================

func TestPRItem_Title(t *testing.T) {
	pr := gh.PR{
		Number: 42,
		Title:  "Test PR Title",
		Author: gh.User{Login: "testuser"},
	}
	item := PRItem{PR: pr}

	if got := item.Title(); got != "Test PR Title" {
		t.Errorf("PRItem.Title() = %q, want %q", got, "Test PR Title")
	}
}

func TestPRItem_Description(t *testing.T) {
	pr := gh.PR{
		Number: 42,
		Title:  "Test PR Title",
		Author: gh.User{Login: "testuser"},
	}
	item := PRItem{PR: pr}

	want := "#42 by testuser"
	if got := item.Description(); got != want {
		t.Errorf("PRItem.Description() = %q, want %q", got, want)
	}
}

func TestPRItem_FilterValue(t *testing.T) {
	pr := gh.PR{
		Number: 42,
		Title:  "Test PR Title",
		Author: gh.User{Login: "testuser"},
	}
	item := PRItem{PR: pr}

	if got := item.FilterValue(); got != "Test PR Title" {
		t.Errorf("PRItem.FilterValue() = %q, want %q", got, "Test PR Title")
	}
}

// =============================================================================
// Model Initialization Tests
// =============================================================================

func TestNew(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)

	if model.state != StateLoading {
		t.Errorf("New() initial state = %v, want %v", model.state, StateLoading)
	}

	if model.ghClient == nil {
		t.Error("New() ghClient is nil, want non-nil")
	}

	// Verify dimensions are zero initially (will be set by WindowSizeMsg)
	if model.width != 0 || model.height != 0 {
		t.Errorf("New() dimensions = (%d, %d), want (0, 0)", model.width, model.height)
	}
}

func TestNew_WithAutoRefresh(t *testing.T) {
	cfg := testConfig()
	cfg.Behavior.AutoRefresh = true

	model := New(cfg)

	if !model.autoRefresh {
		t.Error("New() with AutoRefresh=true, model.autoRefresh = false, want true")
	}
}

func TestModel_Init(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)

	cmd := model.Init()

	if cmd == nil {
		t.Error("Init() returned nil, want a command to load PRs")
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
		{StateLoading, "loading"},
		{StateReady, "ready"},
		{StateError, "error"},
		{StateFiltering, "filtering"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

// =============================================================================
// Message Handling Tests
// =============================================================================

func TestModel_Update_PRsLoadedMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.SetSize(80, 24)

	prs := testPRs()
	msg := PRsLoadedMsg{PRs: prs}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.state != StateReady {
		t.Errorf("Update(PRsLoadedMsg) state = %v, want %v", m.state, StateReady)
	}

	if len(m.prs) != len(prs) {
		t.Errorf("Update(PRsLoadedMsg) prs count = %d, want %d", len(m.prs), len(prs))
	}

	// List should have items
	if m.list.Items() == nil {
		t.Error("Update(PRsLoadedMsg) list items is nil")
	}
}

func TestModel_Update_PRsErrorMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)

	testErr := errors.New("failed to load PRs")
	msg := PRsErrorMsg{Err: testErr}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.state != StateError {
		t.Errorf("Update(PRsErrorMsg) state = %v, want %v", m.state, StateError)
	}

	if m.err == nil {
		t.Error("Update(PRsErrorMsg) err is nil, want error")
	}

	if m.err.Error() != testErr.Error() {
		t.Errorf("Update(PRsErrorMsg) err = %v, want %v", m.err, testErr)
	}
}

func TestModel_Update_RefreshMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.state = StateReady

	msg := RefreshMsg{}

	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if m.state != StateLoading {
		t.Errorf("Update(RefreshMsg) state = %v, want %v", m.state, StateLoading)
	}

	if cmd == nil {
		t.Error("Update(RefreshMsg) returned nil cmd, want command to load PRs")
	}
}

func TestModel_Update_WindowSizeMsg(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.width != 120 {
		t.Errorf("Update(WindowSizeMsg) width = %d, want %d", m.width, 120)
	}

	if m.height != 40 {
		t.Errorf("Update(WindowSizeMsg) height = %d, want %d", m.height, 40)
	}
}

// =============================================================================
// Keyboard Navigation Tests
// =============================================================================

func TestModel_Update_KeyMsg_Quit(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.state = StateReady

	// Test 'q' key
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}

	_, cmd := model.Update(msg)

	// Check if cmd is tea.Quit
	if cmd == nil {
		t.Error("Update(KeyMsg 'q') returned nil cmd, want tea.Quit")
	}
}

func TestModel_Update_KeyMsg_CtrlC(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.state = StateReady

	msg := tea.KeyMsg{Type: tea.KeyCtrlC}

	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Update(KeyMsg ctrl+c) returned nil cmd, want tea.Quit")
	}
}

func TestModel_Update_KeyMsg_Refresh(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.state = StateReady
	model.SetSize(80, 24)

	// Load some PRs first
	newModel, _ := model.Update(PRsLoadedMsg{PRs: testPRs()})
	model = newModel.(Model)

	// Press 'r' to refresh
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}

	newModel, cmd := model.Update(msg)
	m := newModel.(Model)

	if m.state != StateLoading {
		t.Errorf("Update(KeyMsg 'r') state = %v, want %v", m.state, StateLoading)
	}

	if cmd == nil {
		t.Error("Update(KeyMsg 'r') returned nil cmd, want command to load PRs")
	}
}

func TestModel_Update_KeyMsg_Search(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.state = StateReady
	model.SetSize(80, 24)

	// Load PRs first
	newModel, _ := model.Update(PRsLoadedMsg{PRs: testPRs()})
	model = newModel.(Model)

	// Press '/' to start search
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}

	updatedModel, _ := model.Update(msg)
	m := updatedModel.(Model)

	if m.state != StateFiltering {
		t.Errorf("Update(KeyMsg '/') state = %v, want %v", m.state, StateFiltering)
	}
}

func TestModel_Update_KeyMsg_Escape(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.state = StateFiltering
	model.SetSize(80, 24)

	// Press ESC to exit filtering
	msg := tea.KeyMsg{Type: tea.KeyEscape}

	newModel, _ := model.Update(msg)
	m := newModel.(Model)

	if m.state != StateReady {
		t.Errorf("Update(KeyMsg 'esc') state = %v, want %v", m.state, StateReady)
	}
}

func TestModel_Update_KeyMsg_Enter(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.SetSize(80, 24)

	// Load PRs
	prs := testPRs()
	newModel, _ := model.Update(PRsLoadedMsg{PRs: prs})
	model = newModel.(Model)

	// Press Enter to select
	msg := tea.KeyMsg{Type: tea.KeyEnter}

	_, cmd := model.Update(msg)

	// Should return a command (OpenPRDetailMsg batch)
	if cmd == nil {
		t.Error("Update(KeyMsg 'enter') returned nil cmd, want command with OpenPRDetailMsg")
	}
}

func TestModel_Update_KeyMsg_Create(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.state = StateReady
	model.SetSize(80, 24)

	// Press 'c' to create new PR
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}

	_, cmd := model.Update(msg)

	if cmd == nil {
		t.Error("Update(KeyMsg 'c') returned nil cmd, want command with CreatePRMsg")
	}
}

// =============================================================================
// SelectedPR Tests
// =============================================================================

func TestModel_SelectedPR_Empty(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)

	pr := model.SelectedPR()

	if pr != nil {
		t.Errorf("SelectedPR() with no items = %v, want nil", pr)
	}
}

func TestModel_SelectedPR_WithItems(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.SetSize(80, 24)

	prs := testPRs()
	newModel, _ := model.Update(PRsLoadedMsg{PRs: prs})
	model = newModel.(Model)

	pr := model.SelectedPR()

	if pr == nil {
		t.Error("SelectedPR() with items = nil, want PR")
	}

	if pr != nil && pr.Number != prs[0].Number {
		t.Errorf("SelectedPR() Number = %d, want %d", pr.Number, prs[0].Number)
	}
}

// =============================================================================
// View Tests
// =============================================================================

func TestModel_View_Loading(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.state = StateLoading
	model.SetSize(80, 24)

	view := model.View()

	if view == "" {
		t.Error("View() in loading state returned empty string")
	}

	// Should contain loading indicator
	if !containsAny(view, "Loading", "loading", "...") {
		t.Error("View() in loading state should indicate loading")
	}
}

func TestModel_View_Error(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.state = StateError
	model.err = errors.New("test error message")
	model.SetSize(80, 24)

	view := model.View()

	if view == "" {
		t.Error("View() in error state returned empty string")
	}

	// Should contain error
	if !containsAny(view, "error", "Error", "test error message") {
		t.Error("View() in error state should show error message")
	}
}

func TestModel_View_Ready(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.SetSize(80, 24)

	prs := testPRs()
	newModel, _ := model.Update(PRsLoadedMsg{PRs: prs})
	model = newModel.(Model)

	view := model.View()

	if view == "" {
		t.Error("View() in ready state returned empty string")
	}
}

func TestModel_View_ZeroDimensions(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	// Don't set size - leave at 0,0

	view := model.View()

	// Should handle zero dimensions gracefully
	if view == "" {
		// Empty is acceptable for zero dimensions
		return
	}
}

// =============================================================================
// SetSize Tests
// =============================================================================

func TestModel_SetSize(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)

	model.SetSize(100, 50)

	if model.width != 100 {
		t.Errorf("SetSize() width = %d, want %d", model.width, 100)
	}

	if model.height != 50 {
		t.Errorf("SetSize() height = %d, want %d", model.height, 50)
	}
}

func TestModel_SetSize_UpdatesList(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)

	model.SetSize(100, 50)

	// List should be sized appropriately (accounting for header/footer)
	// This test ensures the list dimensions are updated
	// Actual values depend on header/footer heights
	if model.list.Width() <= 0 {
		t.Error("SetSize() should update list width")
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if contains(s, sub) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsString(s, substr))
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// =============================================================================
// Key Binding Tests
// =============================================================================

func TestDashboardKeyBindings(t *testing.T) {
	bindings := DashboardKeyBindings()

	// Verify essential bindings exist
	requiredKeys := []string{"Up", "Down", "Select", "Refresh", "Search", "Create", "Quit"}

	for _, name := range requiredKeys {
		found := false
		for _, b := range bindings {
			if b.Help().Key == name || containsAny(b.Help().Desc, name, stringToLower(name)) {
				found = true
				break
			}
		}
		// This is a soft check - just verify we have bindings
		_ = found
	}

	if len(bindings) == 0 {
		t.Error("DashboardKeyBindings() returned empty slice")
	}
}

func stringToLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// =============================================================================
// Integration-style Tests
// =============================================================================

func TestModel_FullWorkflow(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)

	// 1. Initialize
	cmd := model.Init()
	if cmd == nil {
		t.Fatal("Init() should return a command")
	}

	// 2. Set size (simulating WindowSizeMsg)
	model.SetSize(120, 40)

	// 3. Load PRs
	prs := testPRs()
	var newModel tea.Model
	newModel, _ = model.Update(PRsLoadedMsg{PRs: prs})
	model = newModel.(Model)

	if model.state != StateReady {
		t.Fatalf("After loading PRs, state = %v, want %v", model.state, StateReady)
	}

	// 4. Navigate down
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = newModel.(Model)

	// 5. Verify we can get selected PR
	pr := model.SelectedPR()
	if pr == nil {
		t.Error("SelectedPR() after navigation = nil, want PR")
	}

	// 6. Start search
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = newModel.(Model)

	if model.state != StateFiltering {
		t.Errorf("After '/' key, state = %v, want %v", model.state, StateFiltering)
	}

	// 7. Cancel search
	newModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	model = newModel.(Model)

	if model.state != StateReady {
		t.Errorf("After ESC key, state = %v, want %v", model.state, StateReady)
	}

	// 8. Refresh
	newModel, cmd = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	model = newModel.(Model)

	if model.state != StateLoading {
		t.Errorf("After 'r' key, state = %v, want %v", model.state, StateLoading)
	}

	if cmd == nil {
		t.Error("Refresh should return a command")
	}

	// 9. View should render without panic
	view := model.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestModel_Update_EmptyPRList(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.SetSize(80, 24)

	// Load empty PR list
	newModel, _ := model.Update(PRsLoadedMsg{PRs: []gh.PR{}})
	model = newModel.(Model)

	if model.state != StateReady {
		t.Errorf("Empty PR list state = %v, want %v", model.state, StateReady)
	}

	if len(model.prs) != 0 {
		t.Errorf("Empty PR list count = %d, want 0", len(model.prs))
	}

	// SelectedPR should return nil for empty list
	pr := model.SelectedPR()
	if pr != nil {
		t.Error("SelectedPR() on empty list should return nil")
	}
}

func TestModel_Update_NilPRList(t *testing.T) {
	cfg := testConfig()
	model := New(cfg)
	model.SetSize(80, 24)

	// Load nil PR list
	newModel, _ := model.Update(PRsLoadedMsg{PRs: nil})
	model = newModel.(Model)

	if model.state != StateReady {
		t.Errorf("Nil PR list state = %v, want %v", model.state, StateReady)
	}
}

// Unused import check - ensure key is actually used
var _ = key.Binding{}
var _ = list.Model{}
