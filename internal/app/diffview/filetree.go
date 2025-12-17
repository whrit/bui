package diffview

import (
	"fmt"
	"io"
	"path/filepath"

	"bui/internal/ui"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// =============================================================================
// FileItem
// =============================================================================

// FileItem represents a file in the file tree list.
// It implements the list.Item and list.DefaultItem interfaces.
type FileItem struct {
	Path      string // File path
	Status    string // "added", "deleted", "modified", "renamed"
	Additions int    // Number of lines added
	Deletions int    // Number of lines deleted
	Index     int    // Index into the FileDiffs slice
}

// Title returns the display title for the file item.
// Shows the filename with status indicator.
func (f FileItem) Title() string {
	name := filepath.Base(f.Path)
	symbol := FileStatusSymbol(f.Status)
	return fmt.Sprintf("%s %s", symbol, name)
}

// Description returns the description for the file item.
// Shows the directory path and +/- counts.
func (f FileItem) Description() string {
	dir := filepath.Dir(f.Path)
	if dir == "." {
		dir = ""
	} else {
		dir = dir + "/"
	}

	stats := ""
	if f.Additions > 0 || f.Deletions > 0 {
		stats = fmt.Sprintf(" +%d -%d", f.Additions, f.Deletions)
	}

	if dir != "" {
		return dir + stats
	}
	return stats
}

// FilterValue returns the value used for filtering.
func (f FileItem) FilterValue() string {
	return f.Path
}

// =============================================================================
// File List Creation
// =============================================================================

// NewFileList creates a new list model for the file tree.
func NewFileList(files []FileDiff, palette ui.Palette, styles ui.Styles) list.Model {
	items := make([]list.Item, len(files))
	for i, f := range files {
		items[i] = FileItem{
			Path:      f.DisplayPath(),
			Status:    f.Status,
			Additions: f.Additions,
			Deletions: f.Deletions,
			Index:     i,
		}
	}

	// Create custom delegate with styled items
	delegate := newFileDelegate(palette)

	// Create the list
	l := list.New(items, delegate, 0, 0)
	l.Title = "Files"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	// Style the list
	l.Styles.Title = lipgloss.NewStyle().
		Foreground(palette.Text).
		Bold(true).
		Padding(0, 0, 1, 0)

	l.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(palette.Accent)

	l.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(palette.Accent)

	return l
}

// =============================================================================
// Custom Delegate
// =============================================================================

// fileDelegate is a custom delegate for rendering file items.
type fileDelegate struct {
	palette   ui.Palette
	itemStyle lipgloss.Style
}

// newFileDelegate creates a new file delegate with the given palette.
func newFileDelegate(palette ui.Palette) fileDelegate {
	return fileDelegate{
		palette: palette,
		itemStyle: lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1),
	}
}

// Height returns the height of a single item.
func (d fileDelegate) Height() int {
	return 2
}

// Spacing returns the spacing between items.
func (d fileDelegate) Spacing() int {
	return 0
}

// Update handles messages for the delegate.
func (d fileDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

// Render renders a single item to the provided writer.
func (d fileDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	fileItem, ok := item.(FileItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	// Status symbol styling based on file status
	var symbolStyle lipgloss.Style
	switch fileItem.Status {
	case "added":
		symbolStyle = lipgloss.NewStyle().Foreground(d.palette.Good)
	case "deleted":
		symbolStyle = lipgloss.NewStyle().Foreground(d.palette.Bad)
	case "renamed":
		symbolStyle = lipgloss.NewStyle().Foreground(d.palette.Warn)
	default:
		symbolStyle = lipgloss.NewStyle().Foreground(d.palette.Accent)
	}

	// Name styling
	nameStyle := lipgloss.NewStyle().Foreground(d.palette.Text)
	descStyle := lipgloss.NewStyle().Foreground(d.palette.Muted)

	// Add/delete count styling
	addStyle := lipgloss.NewStyle().Foreground(d.palette.Good)
	delStyle := lipgloss.NewStyle().Foreground(d.palette.Bad)

	// Selected item styling
	if isSelected {
		nameStyle = nameStyle.Bold(true)
	}

	symbol := FileStatusSymbol(fileItem.Status)
	name := filepath.Base(fileItem.Path)

	// Build the first line: symbol + name
	line1 := symbolStyle.Render(symbol) + " " + nameStyle.Render(name)

	// Build the second line: directory + stats
	dir := filepath.Dir(fileItem.Path)
	var line2 string
	if dir != "." && dir != "" {
		line2 = descStyle.Render(dir + "/")
	}

	// Add stats
	if fileItem.Additions > 0 {
		line2 += " " + addStyle.Render(fmt.Sprintf("+%d", fileItem.Additions))
	}
	if fileItem.Deletions > 0 {
		line2 += " " + delStyle.Render(fmt.Sprintf("-%d", fileItem.Deletions))
	}

	// Combine lines
	if line2 != "" {
		line2 = "  " + line2 // Indent to align with name
	}

	// Apply selection highlighting
	baseStyle := d.itemStyle
	if isSelected {
		baseStyle = baseStyle.
			Background(d.palette.Panel).
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(d.palette.Accent)
	}

	_, _ = fmt.Fprint(w, baseStyle.Render(line1+"\n"+line2))
}

// =============================================================================
// File Tree Utilities
// =============================================================================

// FileTreeStats holds aggregate statistics for the file tree.
type FileTreeStats struct {
	TotalFiles    int
	AddedFiles    int
	DeletedFiles  int
	ModifiedFiles int
	RenamedFiles  int
	TotalAdded    int
	TotalDeleted  int
}

// CalculateStats calculates aggregate statistics from file diffs.
func CalculateStats(files []FileDiff) FileTreeStats {
	stats := FileTreeStats{
		TotalFiles: len(files),
	}

	for _, f := range files {
		stats.TotalAdded += f.Additions
		stats.TotalDeleted += f.Deletions

		switch f.Status {
		case "added":
			stats.AddedFiles++
		case "deleted":
			stats.DeletedFiles++
		case "modified":
			stats.ModifiedFiles++
		case "renamed":
			stats.RenamedFiles++
		}
	}

	return stats
}

// FormatStats formats the file tree statistics as a string.
func FormatStats(stats FileTreeStats, palette ui.Palette) string {
	addStyle := lipgloss.NewStyle().Foreground(palette.Good)
	delStyle := lipgloss.NewStyle().Foreground(palette.Bad)
	textStyle := lipgloss.NewStyle().Foreground(palette.Muted)

	return textStyle.Render(fmt.Sprintf("%d files", stats.TotalFiles)) +
		" " + addStyle.Render(fmt.Sprintf("+%d", stats.TotalAdded)) +
		" " + delStyle.Render(fmt.Sprintf("-%d", stats.TotalDeleted))
}
