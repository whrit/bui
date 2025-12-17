package diffview

import (
	"bui/internal/gh"

	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Data Loading Messages
// =============================================================================

// DiffLoadedMsg is sent when the diff has been successfully loaded.
type DiffLoadedMsg struct {
	Diff string
}

// DiffErrorMsg is sent when loading the diff fails.
type DiffErrorMsg struct {
	Err error
}

// =============================================================================
// Navigation Messages
// =============================================================================

// BackToPRDetailMsg signals navigation back to PR detail view.
type BackToPRDetailMsg struct {
	PRNumber int
}

// =============================================================================
// Hunk Selection Messages (for LLM context)
// =============================================================================

// SelectHunkMsg signals hunk selection for LLM context.
type SelectHunkMsg struct {
	File     string // File path containing the hunk
	HunkIdx  int    // Index of the hunk within the file
	Content  string // The hunk content
	Selected bool   // Whether the hunk is now selected or deselected
}

// HunkSelectionChangedMsg is sent when the selection state changes.
type HunkSelectionChangedMsg struct {
	TotalSelected int
	TotalHunks    int
}

// =============================================================================
// File Selection Messages
// =============================================================================

// FileSelectedMsg is sent when a file is selected in the file tree.
type FileSelectedMsg struct {
	Index int
	Path  string
}

// =============================================================================
// Commands
// =============================================================================

// loadDiff creates a command that loads the diff for a PR.
func loadDiff(client *gh.Client, prNumber int) tea.Cmd {
	return func() tea.Msg {
		diff, err := client.GetPRDiff(prNumber)
		if err != nil {
			return DiffErrorMsg{Err: err}
		}
		return DiffLoadedMsg{Diff: diff}
	}
}

// backToPRDetail creates a command that signals navigation back to PR detail.
func backToPRDetail(prNumber int) tea.Cmd {
	return func() tea.Msg {
		return BackToPRDetailMsg{PRNumber: prNumber}
	}
}

// selectHunk creates a command that signals hunk selection.
func selectHunk(file string, hunkIdx int, content string, selected bool) tea.Cmd {
	return func() tea.Msg {
		return SelectHunkMsg{
			File:     file,
			HunkIdx:  hunkIdx,
			Content:  content,
			Selected: selected,
		}
	}
}

// selectFile creates a command that signals file selection.
func selectFile(index int, path string) tea.Cmd {
	return func() tea.Msg {
		return FileSelectedMsg{
			Index: index,
			Path:  path,
		}
	}
}
