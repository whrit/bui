package dashboard

import (
	"fmt"
	"io"

	"bui/internal/gh"
	"bui/internal/ui"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// =============================================================================
// Delegate
// =============================================================================

// NewDelegate creates a new list delegate for PR items.
func NewDelegate(palette ui.Palette, styles ui.Styles) list.DefaultDelegate {
	d := list.NewDefaultDelegate()

	// Customize styles for selected and normal items
	selectedBorder := lipgloss.Border{
		Left: ">",
	}

	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(selectedBorder, false, false, false, true).
		BorderForeground(palette.Accent).
		Foreground(palette.Text).
		Bold(true).
		Padding(0, 0, 0, 1)

	d.Styles.SelectedDesc = lipgloss.NewStyle().
		Border(selectedBorder, false, false, false, true).
		BorderForeground(palette.Accent).
		Foreground(palette.Muted).
		Padding(0, 0, 0, 1)

	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(palette.Text).
		Padding(0, 0, 0, 2)

	d.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(palette.Muted).
		Padding(0, 0, 0, 2)

	d.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(palette.Muted).
		Padding(0, 0, 0, 2)

	d.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(palette.Muted).
		Padding(0, 0, 0, 2)

	d.Styles.FilterMatch = lipgloss.NewStyle().
		Underline(true).
		Bold(true).
		Foreground(palette.Accent)

	// Set spacing
	d.SetSpacing(1)

	return d
}

// =============================================================================
// Custom Delegate (for more control)
// =============================================================================

// customDelegate provides full control over item rendering.
type customDelegate struct {
	palette    ui.Palette
	styles     ui.Styles
	height     int
	spacing    int
	showDetail bool
}

// NewCustomDelegate creates a custom delegate with full rendering control.
func NewCustomDelegate(palette ui.Palette, styles ui.Styles) customDelegate {
	return customDelegate{
		palette:    palette,
		styles:     styles,
		height:     3, // Title + Description + Status line
		spacing:    1,
		showDetail: true,
	}
}

// Height returns the height of each item.
func (d customDelegate) Height() int {
	return d.height
}

// Spacing returns the spacing between items.
func (d customDelegate) Spacing() int {
	return d.spacing
}

// Update handles delegate updates.
func (d customDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

// Render renders a single list item.
func (d customDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	prItem, ok := item.(PRItem)
	if !ok {
		return
	}

	pr := prItem.PR
	selected := index == m.Index()
	width := m.Width()

	// Build the item view
	view := d.renderItem(pr, selected, width)
	_, _ = fmt.Fprint(w, view)
}

// renderItem renders a single PR item with all its details.
func (d customDelegate) renderItem(pr gh.PR, selected bool, width int) string {
	// Determine the state for badge
	state := ui.PRStateOpen
	if pr.IsDraft {
		state = ui.PRStateDraft
	} else if pr.State == "closed" {
		state = ui.PRStateClosed
	} else if pr.State == "merged" {
		state = ui.PRStateMerged
	}

	// Create status badge
	badge := ui.NewStatusBadge(state, d.palette).View()

	// Format PR number
	numberStyle := lipgloss.NewStyle().
		Foreground(d.palette.Muted).
		Bold(true)
	number := numberStyle.Render(fmt.Sprintf("#%d", pr.Number))

	// Format title (truncate if too long)
	titleWidth := width - 20 // Account for number, badge, padding
	if titleWidth < 10 {
		titleWidth = 10
	}
	title := ui.Truncate(pr.Title, titleWidth)

	titleStyle := lipgloss.NewStyle().Foreground(d.palette.Text)
	if selected {
		titleStyle = titleStyle.Bold(true)
	}
	titleRendered := titleStyle.Render(title)

	// Format branch info
	branchStyle := lipgloss.NewStyle().Foreground(d.palette.Muted)
	branchInfo := branchStyle.Render(fmt.Sprintf("%s -> %s", pr.HeadBranch, pr.BaseBranch))

	// Format author and time
	authorStyle := lipgloss.NewStyle().Foreground(d.palette.Muted)
	timeAgo := ui.RelativeTime(pr.UpdatedAt)
	authorInfo := authorStyle.Render(fmt.Sprintf("by %s, updated %s", pr.Author.Login, timeAgo))

	// Format changes
	addStyle := lipgloss.NewStyle().Foreground(d.palette.Good)
	delStyle := lipgloss.NewStyle().Foreground(d.palette.Bad)
	changes := addStyle.Render(fmt.Sprintf("+%d", pr.Additions)) + " " +
		delStyle.Render(fmt.Sprintf("-%d", pr.Deletions))

	// Build lines
	line1 := lipgloss.JoinHorizontal(lipgloss.Left,
		number, " ", titleRendered, " ", badge,
	)

	line2 := lipgloss.JoinHorizontal(lipgloss.Left,
		branchInfo, " ", changes,
	)

	line3 := authorInfo

	// Selection indicator
	cursor := "  "
	if selected {
		cursorStyle := lipgloss.NewStyle().Foreground(d.palette.Accent).Bold(true)
		cursor = cursorStyle.Render("> ")
	}

	// Combine all lines
	content := lipgloss.JoinVertical(lipgloss.Left, line1, line2, line3)

	// Add cursor/indent
	lines := splitLines(content)
	result := ""
	for i, line := range lines {
		if i == 0 {
			result += cursor + line
		} else {
			result += "  " + line
		}
		if i < len(lines)-1 {
			result += "\n"
		}
	}

	return result
}

// splitLines splits a string into lines.
func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// =============================================================================
// Compact Delegate (single line per item)
// =============================================================================

// compactDelegate renders items in a single line for dense display.
type compactDelegate struct {
	palette ui.Palette
	styles  ui.Styles
}

// NewCompactDelegate creates a compact single-line delegate.
func NewCompactDelegate(palette ui.Palette, styles ui.Styles) compactDelegate {
	return compactDelegate{
		palette: palette,
		styles:  styles,
	}
}

// Height returns the height of each item (1 for compact).
func (d compactDelegate) Height() int {
	return 1
}

// Spacing returns the spacing between items.
func (d compactDelegate) Spacing() int {
	return 0
}

// Update handles delegate updates.
func (d compactDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

// Render renders a single compact list item.
func (d compactDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	prItem, ok := item.(PRItem)
	if !ok {
		return
	}

	pr := prItem.PR
	selected := index == m.Index()
	width := m.Width()

	// Build compact view
	view := d.renderCompactItem(pr, selected, width)
	_, _ = fmt.Fprint(w, view)
}

// renderCompactItem renders a single line PR item.
func (d compactDelegate) renderCompactItem(pr gh.PR, selected bool, width int) string {
	// Cursor
	cursor := "  "
	if selected {
		cursorStyle := lipgloss.NewStyle().Foreground(d.palette.Accent).Bold(true)
		cursor = cursorStyle.Render("> ")
	}

	// Number
	numberStyle := lipgloss.NewStyle().Foreground(d.palette.Muted)
	number := numberStyle.Render(fmt.Sprintf("#%-4d", pr.Number))

	// State icon
	var stateIcon string
	switch {
	case pr.IsDraft:
		stateIcon = lipgloss.NewStyle().Foreground(d.palette.Muted).Render("D")
	case pr.State == "merged":
		stateIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("#A371F7")).Render("M")
	case pr.State == "closed":
		stateIcon = lipgloss.NewStyle().Foreground(d.palette.Bad).Render("C")
	default:
		stateIcon = lipgloss.NewStyle().Foreground(d.palette.Good).Render("O")
	}

	// Title (truncated)
	titleWidth := width - 25 // Account for cursor, number, state, author
	if titleWidth < 10 {
		titleWidth = 10
	}
	title := ui.Truncate(pr.Title, titleWidth)

	titleStyle := lipgloss.NewStyle().Foreground(d.palette.Text)
	if selected {
		titleStyle = titleStyle.Bold(true)
	}
	titleRendered := titleStyle.Render(title)

	// Author
	authorStyle := lipgloss.NewStyle().Foreground(d.palette.Muted)
	author := authorStyle.Render(fmt.Sprintf("@%s", ui.Truncate(pr.Author.Login, 10)))

	return cursor + number + " " + stateIcon + " " + titleRendered + " " + author
}
