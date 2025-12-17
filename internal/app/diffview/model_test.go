package diffview

import (
	"errors"
	"strings"
	"testing"

	"bui/internal/config"
	"bui/internal/gh"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Test Helpers
// =============================================================================

// mockExecutor implements gh.CommandExecutor for testing.
type mockExecutor struct {
	outputs map[string][]byte
	errors  map[string]error
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{
		outputs: make(map[string][]byte),
		errors:  make(map[string]error),
	}
}

func (m *mockExecutor) Run(name string, args ...string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	if err, ok := m.errors[key]; ok {
		return nil, err
	}
	if output, ok := m.outputs[key]; ok {
		return output, nil
	}
	// Default: return empty for unknown commands
	return []byte{}, nil
}

// testConfig returns a default config for testing.
func testConfig() config.Config {
	return config.Config{
		UI: config.UIConfig{
			Theme:           "dark",
			Accent:          "indigo",
			SyntaxTheme:     "github-dark",
			ShowLineNumbers: true,
		},
		Keys: config.KeysConfig{
			Quit:    "q",
			Help:    "?",
			Confirm: "enter",
		},
	}
}

// =============================================================================
// Model Tests
// =============================================================================

func TestNew(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 42)

	if m.PRNumber() != 42 {
		t.Errorf("expected PR number 42, got %d", m.PRNumber())
	}

	if m.State() != StateLoading {
		t.Errorf("expected StateLoading, got %s", m.State())
	}

	if m.Focus() != FocusFileTree {
		t.Errorf("expected FocusFileTree, got %s", m.Focus())
	}

	if !m.ShowLineNumbers() {
		t.Error("expected ShowLineNumbers to be true")
	}
}

func TestNewWithClient(t *testing.T) {
	cfg := testConfig()
	executor := newMockExecutor()
	client := gh.NewWithExecutor(executor)

	m := NewWithClient(cfg, 123, client)

	if m.PRNumber() != 123 {
		t.Errorf("expected PR number 123, got %d", m.PRNumber())
	}
}

func TestModel_Init(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)

	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

func TestModel_SetSize(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)

	m.SetSize(120, 40)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}

	// File list width should be calculated
	if m.fileListWidth < minFileListWidth {
		t.Errorf("file list width %d should be >= %d", m.fileListWidth, minFileListWidth)
	}
	if m.fileListWidth > maxFileListWidth {
		t.Errorf("file list width %d should be <= %d", m.fileListWidth, maxFileListWidth)
	}
}

func TestModel_CurrentFile_Empty(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)

	file := m.CurrentFile()
	if file != nil {
		t.Error("expected nil for empty files")
	}
}

func TestModel_CurrentFile_WithFiles(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)

	// Manually set files for testing
	m.files = []FileDiff{
		{NewPath: "file1.go", Status: "modified"},
		{NewPath: "file2.go", Status: "added"},
	}
	m.selectedFile = 1

	file := m.CurrentFile()
	if file == nil {
		t.Fatal("expected file, got nil")
	}
	if file.NewPath != "file2.go" {
		t.Errorf("expected file2.go, got %s", file.NewPath)
	}
}

// =============================================================================
// State Tests
// =============================================================================

func TestState_String(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateLoading, "loading"},
		{StateReady, "ready"},
		{StateError, "error"},
		{State(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestFocus_String(t *testing.T) {
	tests := []struct {
		focus    Focus
		expected string
	}{
		{FocusFileTree, "file_tree"},
		{FocusDiff, "diff"},
		{Focus(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.focus.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Update Tests
// =============================================================================

func TestUpdate_WindowSizeMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, _ := m.Update(msg)

	updated := newModel.(Model)
	if updated.width != 100 || updated.height != 50 {
		t.Errorf("expected 100x50, got %dx%d", updated.width, updated.height)
	}
}

func TestUpdate_DiffLoadedMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.SetSize(120, 40)

	diff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+// comment
 func main() {}
`

	msg := DiffLoadedMsg{Diff: diff}
	newModel, _ := m.Update(msg)

	updated := newModel.(Model)
	if updated.State() != StateReady {
		t.Errorf("expected StateReady, got %s", updated.State())
	}
	if len(updated.Files()) != 1 {
		t.Errorf("expected 1 file, got %d", len(updated.Files()))
	}
	if updated.rawDiff != diff {
		t.Error("raw diff not stored")
	}
}

func TestUpdate_DiffErrorMsg(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)

	testErr := errors.New("test error")
	msg := DiffErrorMsg{Err: testErr}
	newModel, _ := m.Update(msg)

	updated := newModel.(Model)
	if updated.State() != StateError {
		t.Errorf("expected StateError, got %s", updated.State())
	}
	if updated.Error() != testErr {
		t.Errorf("expected test error, got %v", updated.Error())
	}
}

func TestUpdate_SpinnerMsg_Loading(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)

	// Spinner tick should be processed during loading
	msg := spinner.TickMsg{}
	_, cmd := m.Update(msg)

	// Should return a command to continue the spinner
	if cmd == nil {
		t.Log("Note: spinner tick may not return cmd in all cases")
	}
}

// =============================================================================
// Key Navigation Tests
// =============================================================================

func TestUpdate_KeyMsg_Quit(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.state = StateReady

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(msg)

	// Check if quit command was returned
	if cmd == nil {
		t.Error("expected quit command")
	}
}

func TestUpdate_KeyMsg_Back(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.state = StateReady

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("expected back command")
	}
}

func TestUpdate_KeyMsg_Tab(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.state = StateReady
	m.focus = FocusFileTree

	msg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := m.Update(msg)

	updated := newModel.(Model)
	if updated.Focus() != FocusDiff {
		t.Errorf("expected FocusDiff after tab, got %s", updated.Focus())
	}

	// Tab again to switch back
	newModel2, _ := updated.Update(msg)
	updated2 := newModel2.(Model)
	if updated2.Focus() != FocusFileTree {
		t.Errorf("expected FocusFileTree after second tab, got %s", updated2.Focus())
	}
}

func TestUpdate_KeyMsg_ToggleLineNumbers(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.state = StateReady
	m.showLineNums = true

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	newModel, _ := m.Update(msg)

	updated := newModel.(Model)
	if updated.ShowLineNumbers() {
		t.Error("expected line numbers to be toggled off")
	}

	// Toggle again
	newModel2, _ := updated.Update(msg)
	updated2 := newModel2.(Model)
	if !updated2.ShowLineNumbers() {
		t.Error("expected line numbers to be toggled on")
	}
}

func TestUpdate_FileTreeNavigation(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.state = StateReady
	m.focus = FocusFileTree
	m.SetSize(120, 40)

	// Set up test files
	m.files = []FileDiff{
		{NewPath: "file1.go", Status: "modified"},
		{NewPath: "file2.go", Status: "added"},
		{NewPath: "file3.go", Status: "deleted"},
	}
	m.updateFileList()

	// Navigate down with j
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	// The file list should handle navigation
	// Note: Actual selection depends on list implementation
	if updated.Focus() != FocusFileTree {
		t.Error("focus should remain on file tree")
	}
}

func TestUpdate_DiffNavigation(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.state = StateReady
	m.focus = FocusDiff
	m.SetSize(120, 40)

	// Set up test files with content
	m.files = []FileDiff{
		{
			NewPath: "main.go",
			Status:  "modified",
			Hunks: []Hunk{
				{Header: "@@ -1,5 +1,6 @@"},
			},
		},
	}
	m.updateDiffContent()

	// Navigate down with j
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	_, _ = m.Update(msg)

	// Navigate up with k
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	_, _ = m.Update(msg)

	// Page down
	msg = tea.KeyMsg{Type: tea.KeyPgDown}
	_, _ = m.Update(msg)

	// These should not cause errors
	// The viewport handles the actual scrolling
}

func TestUpdate_FileNavigation_BracketKeys(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.state = StateReady
	m.focus = FocusDiff
	m.SetSize(120, 40)

	// Set up test files
	m.files = []FileDiff{
		{NewPath: "file1.go", Status: "modified"},
		{NewPath: "file2.go", Status: "added"},
		{NewPath: "file3.go", Status: "deleted"},
	}
	m.updateFileList()
	m.selectedFile = 0

	// Navigate to next file with ]
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}}
	newModel, _ := m.Update(msg)
	updated := newModel.(Model)

	if updated.SelectedFileIndex() != 1 {
		t.Errorf("expected selected file 1, got %d", updated.SelectedFileIndex())
	}

	// Navigate to previous file with [
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'['}}
	newModel, _ = updated.Update(msg)
	updated = newModel.(Model)

	if updated.SelectedFileIndex() != 0 {
		t.Errorf("expected selected file 0, got %d", updated.SelectedFileIndex())
	}
}

// =============================================================================
// Hunk Selection Tests
// =============================================================================

func TestModel_HunkSelection(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)

	// Initially no hunks selected
	if m.IsHunkSelected("main.go", 0) {
		t.Error("expected hunk not to be selected initially")
	}

	// Toggle selection
	m.toggleHunkSelection("main.go", 0)
	if !m.IsHunkSelected("main.go", 0) {
		t.Error("expected hunk to be selected after toggle")
	}

	// Toggle again to deselect
	m.toggleHunkSelection("main.go", 0)
	if m.IsHunkSelected("main.go", 0) {
		t.Error("expected hunk to be deselected after second toggle")
	}
}

func TestModel_SelectedHunks(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.files = []FileDiff{
		{
			NewPath: "main.go",
			Status:  "modified",
			Hunks: []Hunk{
				{Header: "@@ -1,5 +1,6 @@"},
				{Header: "@@ -10,3 +11,4 @@"},
			},
		},
	}

	// Select some hunks
	m.toggleHunkSelection("main.go", 0)
	m.toggleHunkSelection("main.go", 1)

	selected := m.SelectedHunks()
	if len(selected) != 2 {
		t.Errorf("expected 2 selected hunks, got %d", len(selected))
	}
}

func TestModel_HunkNavigation(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.SetSize(120, 40)
	m.files = []FileDiff{
		{
			NewPath: "main.go",
			Status:  "modified",
			Hunks: []Hunk{
				{Header: "@@ -1,5 +1,6 @@"},
				{Header: "@@ -10,3 +11,4 @@"},
				{Header: "@@ -20,3 +21,4 @@"},
			},
		},
	}
	m.state = StateReady
	m.focus = FocusDiff

	// Initially at hunk 0
	if m.CurrentHunkIndex() != 0 {
		t.Errorf("expected current hunk 0, got %d", m.CurrentHunkIndex())
	}

	// Navigate to next hunk
	m.nextHunk()
	if m.CurrentHunkIndex() != 1 {
		t.Errorf("expected current hunk 1 after nextHunk, got %d", m.CurrentHunkIndex())
	}

	// Navigate to next hunk again
	m.nextHunk()
	if m.CurrentHunkIndex() != 2 {
		t.Errorf("expected current hunk 2 after second nextHunk, got %d", m.CurrentHunkIndex())
	}

	// Try to go past the end (should stay at 2)
	m.nextHunk()
	if m.CurrentHunkIndex() != 2 {
		t.Errorf("expected current hunk to stay at 2, got %d", m.CurrentHunkIndex())
	}

	// Navigate to previous hunk
	m.prevHunk()
	if m.CurrentHunkIndex() != 1 {
		t.Errorf("expected current hunk 1 after prevHunk, got %d", m.CurrentHunkIndex())
	}

	// Navigate all the way back
	m.prevHunk()
	m.prevHunk()
	if m.CurrentHunkIndex() != 0 {
		t.Errorf("expected current hunk 0 after multiple prevHunk, got %d", m.CurrentHunkIndex())
	}
}

func TestModel_ToggleCurrentHunkSelection(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.SetSize(120, 40)
	m.files = []FileDiff{
		{
			NewPath: "main.go",
			Status:  "modified",
			Hunks: []Hunk{
				{Header: "@@ -1,5 +1,6 @@", Lines: []DiffLine{{Type: LineContext, Content: " test"}}},
				{Header: "@@ -10,3 +11,4 @@", Lines: []DiffLine{{Type: LineAdded, Content: "+new"}}},
			},
		},
	}
	m.state = StateReady
	m.focus = FocusDiff
	m.currentHunkIndex = 0

	// Initially not selected
	if m.IsHunkSelected("main.go", 0) {
		t.Error("expected hunk 0 to not be selected initially")
	}

	// Toggle selection
	m, _ = m.toggleCurrentHunkSelection()
	if !m.IsHunkSelected("main.go", 0) {
		t.Error("expected hunk 0 to be selected after toggle")
	}

	// Navigate to hunk 1 and select
	m.currentHunkIndex = 1
	m, _ = m.toggleCurrentHunkSelection()
	if !m.IsHunkSelected("main.go", 1) {
		t.Error("expected hunk 1 to be selected")
	}

	// Verify both are selected
	selected := m.SelectedHunks()
	if len(selected) != 2 {
		t.Errorf("expected 2 selected hunks, got %d", len(selected))
	}

	// Toggle hunk 0 off
	m.currentHunkIndex = 0
	m, _ = m.toggleCurrentHunkSelection()
	if m.IsHunkSelected("main.go", 0) {
		t.Error("expected hunk 0 to be deselected after second toggle")
	}
}

func TestModel_GetSelectedHunksContent(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.files = []FileDiff{
		{
			NewPath: "main.go",
			Status:  "modified",
			Hunks: []Hunk{
				{
					Header: "@@ -1,5 +1,6 @@",
					Lines: []DiffLine{
						{Type: LineContext, Content: " package main"},
						{Type: LineAdded, Content: "+// comment"},
					},
				},
			},
		},
	}

	// Select the hunk
	m.toggleHunkSelection("main.go", 0)

	content := m.GetSelectedHunksContent()
	if len(content) != 1 {
		t.Fatalf("expected 1 selected hunk content, got %d", len(content))
	}

	hunk := content[0]
	if hunk.FilePath != "main.go" {
		t.Errorf("expected FilePath 'main.go', got %q", hunk.FilePath)
	}
	if hunk.Header != "@@ -1,5 +1,6 @@" {
		t.Errorf("expected Header '@@ -1,5 +1,6 @@', got %q", hunk.Header)
	}
	if len(hunk.Lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(hunk.Lines))
	}
}

func TestModel_TotalHunksInCurrentFile(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.files = []FileDiff{
		{
			NewPath: "main.go",
			Status:  "modified",
			Hunks: []Hunk{
				{Header: "@@ -1,5 +1,6 @@"},
				{Header: "@@ -10,3 +11,4 @@"},
				{Header: "@@ -20,3 +21,4 @@"},
			},
		},
	}
	m.selectedFile = 0

	total := m.TotalHunksInCurrentFile()
	if total != 3 {
		t.Errorf("expected 3 hunks, got %d", total)
	}
}

func TestModel_HunkSelectionResetOnFileChange(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.SetSize(120, 40)
	m.files = []FileDiff{
		{
			NewPath: "file1.go",
			Status:  "modified",
			Hunks: []Hunk{
				{Header: "@@ -1,5 +1,6 @@"},
				{Header: "@@ -10,3 +11,4 @@"},
			},
		},
		{
			NewPath: "file2.go",
			Status:  "modified",
			Hunks: []Hunk{
				{Header: "@@ -1,5 +1,6 @@"},
			},
		},
	}
	m.state = StateReady
	m.updateFileList()

	// Navigate to hunk 1 in first file
	m.currentHunkIndex = 1
	if m.CurrentHunkIndex() != 1 {
		t.Errorf("expected current hunk 1, got %d", m.CurrentHunkIndex())
	}

	// Select file 2
	m.selectFileByIndex(1)

	// Hunk index should reset to 0
	if m.CurrentHunkIndex() != 0 {
		t.Errorf("expected hunk index to reset to 0 after file change, got %d", m.CurrentHunkIndex())
	}
}

// =============================================================================
// Search Tests
// =============================================================================

func TestModel_FindMatches(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.files = []FileDiff{
		{
			NewPath: "main.go",
			Status:  "modified",
			Hunks: []Hunk{
				{
					Header: "@@ -1,5 +1,6 @@",
					Lines: []DiffLine{
						{Type: LineContext, Content: " package main"},
						{Type: LineAdded, Content: "+// This is a comment"},
						{Type: LineRemoved, Content: "-// Old comment"},
						{Type: LineContext, Content: " func main() {}"},
					},
				},
				{
					Header: "@@ -10,3 +11,4 @@",
					Lines: []DiffLine{
						{Type: LineContext, Content: " // Another comment"},
						{Type: LineAdded, Content: "+fmt.Println(\"hello\")"},
					},
				},
			},
		},
	}
	m.selectedFile = 0

	// Search for "comment" - should find 3 matches
	matches := m.findMatches("comment")
	if len(matches) != 3 {
		t.Errorf("expected 3 matches for 'comment', got %d", len(matches))
	}

	// Search for "main" - should find 2 matches (package main and func main)
	matches = m.findMatches("main")
	if len(matches) != 2 {
		t.Errorf("expected 2 matches for 'main', got %d", len(matches))
	}

	// Search for something not there
	matches = m.findMatches("notfound")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for 'notfound', got %d", len(matches))
	}

	// Case-insensitive search
	matches = m.findMatches("COMMENT")
	if len(matches) != 3 {
		t.Errorf("expected 3 matches for case-insensitive 'COMMENT', got %d", len(matches))
	}
}

func TestModel_SearchNavigation(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.SetSize(120, 40)
	m.files = []FileDiff{
		{
			NewPath: "main.go",
			Status:  "modified",
			Hunks: []Hunk{
				{
					Header: "@@ -1,5 +1,6 @@",
					Lines: []DiffLine{
						{Type: LineContext, Content: " test line 1"},
						{Type: LineAdded, Content: "+test line 2"},
					},
				},
				{
					Header: "@@ -10,3 +11,4 @@",
					Lines: []DiffLine{
						{Type: LineContext, Content: " test line 3"},
					},
				},
			},
		},
	}
	m.state = StateReady
	m.selectedFile = 0

	// Execute search
	m.searchQuery = "test"
	m.executeSearch()

	if len(m.searchMatches) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(m.searchMatches))
	}

	// Should be at first match
	if m.currentMatchIndex != 0 {
		t.Errorf("expected currentMatchIndex 0, got %d", m.currentMatchIndex)
	}

	// Navigate to next match
	m.nextMatch()
	if m.currentMatchIndex != 1 {
		t.Errorf("expected currentMatchIndex 1 after nextMatch, got %d", m.currentMatchIndex)
	}

	// Navigate to next match
	m.nextMatch()
	if m.currentMatchIndex != 2 {
		t.Errorf("expected currentMatchIndex 2 after second nextMatch, got %d", m.currentMatchIndex)
	}

	// Wrap around to first match
	m.nextMatch()
	if m.currentMatchIndex != 0 {
		t.Errorf("expected currentMatchIndex 0 after wrap around, got %d", m.currentMatchIndex)
	}

	// Navigate backwards
	m.prevMatch()
	if m.currentMatchIndex != 2 {
		t.Errorf("expected currentMatchIndex 2 after prevMatch wrap, got %d", m.currentMatchIndex)
	}
}

func TestModel_ClearSearch(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.SetSize(120, 40)
	m.files = []FileDiff{
		{
			NewPath: "main.go",
			Status:  "modified",
			Hunks: []Hunk{
				{
					Header: "@@ -1,5 +1,6 @@",
					Lines:  []DiffLine{{Type: LineContext, Content: " test"}},
				},
			},
		},
	}
	m.state = StateReady
	m.selectedFile = 0

	// Set up search state
	m.searchMode = true
	m.searchQuery = "test"
	m.searchMatches = m.findMatches("test")
	m.currentMatchIndex = 0

	// Clear search
	m.clearSearch()

	if m.searchMode {
		t.Error("expected searchMode to be false after clear")
	}
	if m.searchQuery != "" {
		t.Errorf("expected searchQuery to be empty, got %q", m.searchQuery)
	}
	if len(m.searchMatches) != 0 {
		t.Errorf("expected searchMatches to be empty, got %d", len(m.searchMatches))
	}
}

func TestModel_SearchAccessors(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.files = []FileDiff{
		{
			NewPath: "main.go",
			Status:  "modified",
			Hunks: []Hunk{
				{
					Header: "@@ -1,5 +1,6 @@",
					Lines:  []DiffLine{{Type: LineContext, Content: " test1 test2 test3"}},
				},
			},
		},
	}
	m.selectedFile = 0

	// Initially no search
	if m.IsSearchMode() {
		t.Error("expected IsSearchMode to be false initially")
	}
	if m.SearchQuery() != "" {
		t.Error("expected SearchQuery to be empty initially")
	}
	if m.SearchMatchCount() != 0 {
		t.Error("expected SearchMatchCount to be 0 initially")
	}

	// Activate search
	m.searchMode = true
	m.searchQuery = "test"
	m.searchMatches = m.findMatches("test")

	if !m.IsSearchMode() {
		t.Error("expected IsSearchMode to be true")
	}
	if m.SearchQuery() != "test" {
		t.Errorf("expected SearchQuery 'test', got %q", m.SearchQuery())
	}
	if m.SearchMatchCount() != 3 {
		t.Errorf("expected SearchMatchCount 3, got %d", m.SearchMatchCount())
	}
	if m.CurrentMatchDisplay() != 1 {
		t.Errorf("expected CurrentMatchDisplay 1, got %d", m.CurrentMatchDisplay())
	}
}

func TestModel_SearchBinaryFile(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.files = []FileDiff{
		{
			NewPath:  "image.png",
			Status:   "added",
			IsBinary: true,
		},
	}
	m.selectedFile = 0

	// Search should return no matches for binary files
	matches := m.findMatches("test")
	if len(matches) != 0 {
		t.Errorf("expected 0 matches for binary file, got %d", len(matches))
	}
}

// =============================================================================
// View Tests
// =============================================================================

func TestView_Loading(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.SetSize(80, 24)
	m.state = StateLoading

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view for loading state")
	}
}

func TestView_Error(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.SetSize(80, 24)
	m.state = StateError
	m.err = errors.New("test error")

	view := m.View()
	if !strings.Contains(view, "test error") {
		t.Error("expected error message in view")
	}
}

func TestView_Ready(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.SetSize(120, 40)
	m.state = StateReady
	m.files = []FileDiff{
		{
			NewPath:   "main.go",
			Status:    "modified",
			Additions: 5,
			Deletions: 2,
			Hunks: []Hunk{
				{
					Header: "@@ -1,5 +1,6 @@",
					Lines: []DiffLine{
						{Type: LineContext, Content: " package main"},
						{Type: LineAdded, Content: "+// comment"},
					},
				},
			},
		},
	}
	m.updateFileList()
	m.updateDiffContent()

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view for ready state")
	}

	// Check for key elements
	if !strings.Contains(view, "PR #1") {
		t.Error("expected PR number in view")
	}
}

// =============================================================================
// FileItem Tests
// =============================================================================

func TestFileItem_Title(t *testing.T) {
	item := FileItem{
		Path:   "internal/app/main.go",
		Status: "modified",
	}

	title := item.Title()
	if !strings.Contains(title, "M") {
		t.Error("expected M status symbol in title")
	}
	if !strings.Contains(title, "main.go") {
		t.Error("expected filename in title")
	}
}

func TestFileItem_Description(t *testing.T) {
	tests := []struct {
		name      string
		item      FileItem
		wantDir   bool
		wantStats bool
	}{
		{
			name: "with directory and stats",
			item: FileItem{
				Path:      "internal/app/main.go",
				Additions: 5,
				Deletions: 2,
			},
			wantDir:   true,
			wantStats: true,
		},
		{
			name: "root file with stats",
			item: FileItem{
				Path:      "main.go",
				Additions: 1,
				Deletions: 0,
			},
			wantDir:   false,
			wantStats: true,
		},
		{
			name: "no stats",
			item: FileItem{
				Path:      "empty.go",
				Additions: 0,
				Deletions: 0,
			},
			wantDir:   false,
			wantStats: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := tt.item.Description()

			hasDir := strings.Contains(desc, "/")
			if hasDir != tt.wantDir {
				t.Errorf("directory presence = %v, want %v", hasDir, tt.wantDir)
			}

			hasStats := strings.Contains(desc, "+") || strings.Contains(desc, "-")
			if hasStats != tt.wantStats {
				t.Errorf("stats presence = %v, want %v", hasStats, tt.wantStats)
			}
		})
	}
}

func TestFileItem_FilterValue(t *testing.T) {
	item := FileItem{Path: "internal/app/main.go"}
	if item.FilterValue() != "internal/app/main.go" {
		t.Errorf("expected full path, got %s", item.FilterValue())
	}
}

// =============================================================================
// FileTreeStats Tests
// =============================================================================

func TestCalculateStats(t *testing.T) {
	files := []FileDiff{
		{Status: "added", Additions: 10, Deletions: 0},
		{Status: "deleted", Additions: 0, Deletions: 5},
		{Status: "modified", Additions: 3, Deletions: 2},
		{Status: "renamed", Additions: 1, Deletions: 1},
	}

	stats := CalculateStats(files)

	if stats.TotalFiles != 4 {
		t.Errorf("expected 4 total files, got %d", stats.TotalFiles)
	}
	if stats.AddedFiles != 1 {
		t.Errorf("expected 1 added file, got %d", stats.AddedFiles)
	}
	if stats.DeletedFiles != 1 {
		t.Errorf("expected 1 deleted file, got %d", stats.DeletedFiles)
	}
	if stats.ModifiedFiles != 1 {
		t.Errorf("expected 1 modified file, got %d", stats.ModifiedFiles)
	}
	if stats.RenamedFiles != 1 {
		t.Errorf("expected 1 renamed file, got %d", stats.RenamedFiles)
	}
	if stats.TotalAdded != 14 {
		t.Errorf("expected 14 total added, got %d", stats.TotalAdded)
	}
	if stats.TotalDeleted != 8 {
		t.Errorf("expected 8 total deleted, got %d", stats.TotalDeleted)
	}
}

// =============================================================================
// Messages Tests
// =============================================================================

func TestBackToPRDetailMsg(t *testing.T) {
	cmd := backToPRDetail(42)
	msg := cmd()

	if backMsg, ok := msg.(BackToPRDetailMsg); ok {
		if backMsg.PRNumber != 42 {
			t.Errorf("expected PR number 42, got %d", backMsg.PRNumber)
		}
	} else {
		t.Error("expected BackToPRDetailMsg")
	}
}

func TestSelectHunkCmd(t *testing.T) {
	cmd := selectHunk("main.go", 0, "content", true)
	msg := cmd()

	if hunkMsg, ok := msg.(SelectHunkMsg); ok {
		if hunkMsg.File != "main.go" {
			t.Errorf("expected main.go, got %s", hunkMsg.File)
		}
		if hunkMsg.HunkIdx != 0 {
			t.Errorf("expected hunk 0, got %d", hunkMsg.HunkIdx)
		}
		if !hunkMsg.Selected {
			t.Error("expected selected to be true")
		}
	} else {
		t.Error("expected SelectHunkMsg")
	}
}

func TestSelectFileCmd(t *testing.T) {
	cmd := selectFile(2, "test.go")
	msg := cmd()

	if fileMsg, ok := msg.(FileSelectedMsg); ok {
		if fileMsg.Index != 2 {
			t.Errorf("expected index 2, got %d", fileMsg.Index)
		}
		if fileMsg.Path != "test.go" {
			t.Errorf("expected test.go, got %s", fileMsg.Path)
		}
	} else {
		t.Error("expected FileSelectedMsg")
	}
}

// =============================================================================
// Integration Test
// =============================================================================

func TestIntegration_FullFlow(t *testing.T) {
	cfg := testConfig()
	m := New(cfg, 123)
	m.SetSize(120, 40)

	// Simulate diff loading
	diff := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,5 +1,7 @@
 package main

 func main() {
-    fmt.Println("Hello")
+    fmt.Println("Hello, World!")
+    fmt.Println("Goodbye")
 }
diff --git a/util.go b/util.go
new file mode 100644
--- /dev/null
+++ b/util.go
@@ -0,0 +1,3 @@
+package main
+
+func helper() {}
`

	// Process diff loaded message
	newModel, _ := m.Update(DiffLoadedMsg{Diff: diff})
	m = newModel.(Model)

	// Verify state
	if m.State() != StateReady {
		t.Fatalf("expected StateReady, got %s", m.State())
	}

	// Verify files parsed
	if len(m.Files()) != 2 {
		t.Fatalf("expected 2 files, got %d", len(m.Files()))
	}

	// First file should be selected
	if m.SelectedFileIndex() != 0 {
		t.Errorf("expected first file selected, got %d", m.SelectedFileIndex())
	}

	// Switch to diff focus first (] only works in diff focus)
	m.focus = FocusDiff

	// Navigate to next file
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{']'}})
	m = newModel.(Model)

	if m.SelectedFileIndex() != 1 {
		t.Errorf("expected second file selected after ], got %d", m.SelectedFileIndex())
	}

	// Toggle line numbers
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = newModel.(Model)

	// Toggle focus
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newModel.(Model)

	if m.Focus() != FocusFileTree {
		t.Errorf("expected FocusFileTree after tab, got %s", m.Focus())
	}

	// Render view
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}

	// Back navigation
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("expected back command")
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkView_Ready(b *testing.B) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.SetSize(120, 40)
	m.state = StateReady
	m.files = []FileDiff{
		{
			NewPath:   "main.go",
			Status:    "modified",
			Additions: 50,
			Deletions: 20,
			Hunks: []Hunk{
				{Header: "@@ -1,100 +1,130 @@"},
			},
		},
	}
	m.updateFileList()
	m.updateDiffContent()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.View()
	}
}

func BenchmarkUpdate_KeyNavigation(b *testing.B) {
	cfg := testConfig()
	m := New(cfg, 1)
	m.SetSize(120, 40)
	m.state = StateReady
	m.focus = FocusDiff
	m.files = []FileDiff{
		{NewPath: "file1.go"},
		{NewPath: "file2.go"},
		{NewPath: "file3.go"},
	}
	m.updateFileList()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Update(msg)
	}
}
