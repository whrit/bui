package diffview

import (
	"fmt"
	"strings"

	"bui/internal/syntax"
	"bui/internal/ui"

	"github.com/charmbracelet/lipgloss"
)

// view renders the complete diff view screen.
func (m Model) view() string {
	switch m.state {
	case StateLoading:
		return m.renderLoading()
	case StateError:
		return m.renderError()
	case StateReady:
		return m.renderReady()
	default:
		return ""
	}
}

// =============================================================================
// State Renderers
// =============================================================================

// renderLoading renders the loading state.
func (m Model) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	content := m.spinner.View()

	return style.Render(content)
}

// renderError renders the error state.
func (m Model) renderError() string {
	errDisplay := ui.NewErrorDisplay(m.err, m.palette, m.styles).
		WithTitle("Failed to load diff")

	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	helpText := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		Render("\nPress 'esc' to go back")

	return style.Render(errDisplay.View() + helpText)
}

// renderReady renders the main split-panel view when data is loaded.
func (m Model) renderReady() string {
	// Build header
	header := m.renderHeader()

	// Build main content (split panel)
	mainContent := m.renderSplitPanel()

	// Build footer
	footer := m.renderFooter()

	// Combine all sections
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		mainContent,
		footer,
	)
}

// =============================================================================
// Header
// =============================================================================

// renderHeader renders the top header bar.
func (m Model) renderHeader() string {
	// PR info
	titleStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true)

	mutedStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	// Calculate stats
	stats := CalculateStats(m.files)
	statsStr := FormatStats(stats, m.palette)

	title := titleStyle.Render(fmt.Sprintf("PR #%d Diff", m.prNumber))

	headerStyle := lipgloss.NewStyle().
		Width(m.width).
		Padding(0, 1).
		Background(m.palette.Panel)

	// Show file/total info
	fileInfo := mutedStyle.Render(fmt.Sprintf("(%d/%d)", m.selectedFile+1, len(m.files)))

	leftSide := title + " " + fileInfo
	rightSide := statsStr

	// Calculate spacing
	spacing := m.width - lipgloss.Width(leftSide) - lipgloss.Width(rightSide) - 4
	if spacing < 1 {
		spacing = 1
	}

	return headerStyle.Render(leftSide + strings.Repeat(" ", spacing) + rightSide)
}

// =============================================================================
// Split Panel
// =============================================================================

// renderSplitPanel renders the main split panel with file tree and diff content.
func (m Model) renderSplitPanel() string {
	fileTree := m.renderFileTree()
	diffContent := m.renderDiffPanel()

	// Add border between panels
	borderStyle := lipgloss.NewStyle().
		Foreground(m.palette.Border)

	separator := borderStyle.Render(strings.Repeat("|", m.height-footerHeight-headerHeight))

	return lipgloss.JoinHorizontal(lipgloss.Top,
		fileTree,
		separator,
		diffContent,
	)
}

// renderFileTree renders the file tree panel.
func (m Model) renderFileTree() string {
	// Panel style with focus indicator
	panelStyle := lipgloss.NewStyle().
		Width(m.fileListWidth).
		Height(m.height-footerHeight-headerHeight).
		Padding(0, 1)

	// Add focus border if focused
	if m.focus == FocusFileTree {
		panelStyle = panelStyle.
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(m.palette.Accent).
			BorderLeft(true).
			BorderRight(false).
			BorderTop(false).
			BorderBottom(false)
	}

	return panelStyle.Render(m.fileList.View())
}

// renderDiffPanel renders the diff content panel.
func (m Model) renderDiffPanel() string {
	// Calculate panel width
	panelWidth := m.width - m.fileListWidth - borderWidth - 2

	panelStyle := lipgloss.NewStyle().
		Width(panelWidth).
		Height(m.height-footerHeight-headerHeight).
		Padding(0, 1)

	// Add focus border if focused
	if m.focus == FocusDiff {
		panelStyle = panelStyle.
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(m.palette.Accent).
			BorderLeft(true).
			BorderRight(false).
			BorderTop(false).
			BorderBottom(false)
	}

	return panelStyle.Render(m.diffViewport.View())
}

// =============================================================================
// Diff Content Rendering
// =============================================================================

// renderCurrentFileDiff renders the diff content for the currently selected file.
func (m *Model) renderCurrentFileDiff() string {
	file := m.CurrentFile()
	if file == nil {
		return m.renderEmptyDiff()
	}

	var sb strings.Builder

	// File header
	sb.WriteString(m.renderFileHeader(file))
	sb.WriteString("\n\n")

	// Check for binary file
	if file.IsBinary {
		sb.WriteString(m.renderBinaryFile())
		return sb.String()
	}

	// Render each hunk
	for hunkIdx, hunk := range file.Hunks {
		sb.WriteString(m.renderHunk(file.DisplayPath(), hunkIdx, hunk))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderEmptyDiff renders placeholder for when no file is selected.
func (m *Model) renderEmptyDiff() string {
	style := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		Italic(true)

	return style.Render("No file selected")
}

// renderBinaryFile renders a placeholder for binary files.
func (m *Model) renderBinaryFile() string {
	boxStyle := lipgloss.NewStyle().
		Foreground(m.palette.Warn).
		Background(m.palette.Panel).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Warn)

	return boxStyle.Render("[Binary file - no diff available]")
}

// renderFileHeader renders the header for a file diff.
func (m *Model) renderFileHeader(file *FileDiff) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true)

	pathStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent)

	mutedStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	statusStyle := m.getStatusStyle(file.Status)

	header := headerStyle.Render(FileStatusSymbol(file.Status)) +
		" " + pathStyle.Render(file.DisplayPath())

	// Add renamed info if applicable
	if file.Status == "renamed" && file.OldPath != file.NewPath {
		header += " " + mutedStyle.Render(fmt.Sprintf("(from %s)", file.OldPath))
	}

	// Add stats
	addStyle := lipgloss.NewStyle().Foreground(m.palette.Good)
	delStyle := lipgloss.NewStyle().Foreground(m.palette.Bad)

	stats := ""
	if file.Additions > 0 {
		stats += " " + addStyle.Render(fmt.Sprintf("+%d", file.Additions))
	}
	if file.Deletions > 0 {
		stats += " " + delStyle.Render(fmt.Sprintf("-%d", file.Deletions))
	}

	return header + stats + "\n" + statusStyle.Render(strings.Repeat("-", 40))
}

// renderHunk renders a single hunk with optional line numbers.
func (m *Model) renderHunk(filePath string, hunkIdx int, hunk Hunk) string {
	var sb strings.Builder

	// Check if hunk is selected (for LLM context)
	isSelected := m.IsHunkSelected(filePath, hunkIdx)
	// Check if this is the current (focused) hunk
	isCurrent := hunkIdx == m.currentHunkIndex && m.focus == FocusDiff

	// Current hunk indicator
	if isCurrent {
		currentStyle := lipgloss.NewStyle().
			Foreground(m.palette.Accent).
			Bold(true)
		sb.WriteString(currentStyle.Render("> "))
	}

	// Selection indicator
	if isSelected {
		selectStyle := lipgloss.NewStyle().
			Foreground(m.palette.Good).
			Bold(true)
		sb.WriteString(selectStyle.Render("[SELECTED] "))
	}

	// Hunk header styling - highlight if current
	hunkHeaderStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		Background(m.palette.Panel)

	if isCurrent {
		hunkHeaderStyle = hunkHeaderStyle.
			Foreground(m.palette.Accent).
			Bold(true)
	}

	sb.WriteString(hunkHeaderStyle.Render(hunk.Header))
	sb.WriteString("\n")

	// Get lexer for syntax highlighting
	lexerName := syntax.LexerForFile(filePath)
	themeName := syntax.ResolveTheme(m.cfg.UI.SyntaxTheme)

	// Render each line in the hunk
	for _, line := range hunk.Lines {
		if line.Type == LineHeader {
			continue // Skip header lines as we already rendered the hunk header
		}
		sb.WriteString(m.renderDiffLine(line, lexerName, themeName))
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderDiffLine renders a single diff line with optional line numbers and syntax highlighting.
func (m *Model) renderDiffLine(line DiffLine, lexerName, themeName string) string {
	var sb strings.Builder

	// Line numbers (if enabled)
	if m.showLineNums {
		sb.WriteString(m.renderLineNumbers(line))
		sb.WriteString(" ")
	}

	// Apply syntax highlighting to the line content
	highlighted := m.highlightLine(line, lexerName, themeName)

	// Apply search highlighting if there's an active search
	if len(m.searchMatches) > 0 && m.searchQuery != "" {
		highlighted = m.highlightSearchMatches(highlighted, line.Content)
	}

	sb.WriteString(highlighted)

	return sb.String()
}

// highlightSearchMatches highlights search matches in the line content.
func (m *Model) highlightSearchMatches(rendered, original string) string {
	if m.searchQuery == "" {
		return rendered
	}

	lowerOriginal := strings.ToLower(original)
	lowerQuery := strings.ToLower(m.searchQuery)

	// Check if there are any matches in this line
	if !strings.Contains(lowerOriginal, lowerQuery) {
		return rendered
	}

	// Create highlight style
	highlightStyle := lipgloss.NewStyle().
		Background(m.palette.Warn).
		Foreground(m.palette.Panel).
		Bold(true)

	// For simplicity, we highlight in the original content and re-apply styling
	// This is a basic implementation that works with plain text
	result := rendered

	// Find and highlight all occurrences (case-insensitive)
	lowerRendered := strings.ToLower(result)
	startIdx := 0
	var parts []string
	lastEnd := 0

	for {
		idx := strings.Index(lowerRendered[startIdx:], lowerQuery)
		if idx == -1 {
			break
		}
		matchStart := startIdx + idx
		matchEnd := matchStart + len(m.searchQuery)

		// Add text before match
		if matchStart > lastEnd {
			parts = append(parts, result[lastEnd:matchStart])
		}

		// Add highlighted match
		matchText := result[matchStart:matchEnd]
		parts = append(parts, highlightStyle.Render(matchText))

		lastEnd = matchEnd
		startIdx = matchEnd
	}

	// Add remaining text
	if lastEnd < len(result) {
		parts = append(parts, result[lastEnd:])
	}

	if len(parts) > 0 {
		return strings.Join(parts, "")
	}

	return rendered
}

// renderLineNumbers renders the line number gutter.
func (m *Model) renderLineNumbers(line DiffLine) string {
	numStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		Width(4).
		Align(lipgloss.Right)

	switch line.Type {
	case LineAdded:
		return numStyle.Render("") + " " + numStyle.Render(fmt.Sprintf("%d", line.NewNum))
	case LineRemoved:
		return numStyle.Render(fmt.Sprintf("%d", line.OldNum)) + " " + numStyle.Render("")
	case LineContext:
		oldNum := ""
		newNum := ""
		if line.OldNum > 0 {
			oldNum = fmt.Sprintf("%d", line.OldNum)
		}
		if line.NewNum > 0 {
			newNum = fmt.Sprintf("%d", line.NewNum)
		}
		return numStyle.Render(oldNum) + " " + numStyle.Render(newNum)
	default:
		return numStyle.Render("") + " " + numStyle.Render("")
	}
}

// highlightLine applies syntax highlighting to a diff line.
func (m *Model) highlightLine(line DiffLine, lexerName, themeName string) string {
	// Use the syntax package to highlight the diff line
	highlighted, err := syntax.HighlightDiff(line.Content, themeName, lexerName)
	if err != nil {
		// Fall back to basic coloring
		return m.colorDiffLine(line)
	}
	return highlighted
}

// colorDiffLine applies basic diff coloring to a line (fallback).
func (m *Model) colorDiffLine(line DiffLine) string {
	var style lipgloss.Style

	switch line.Type {
	case LineAdded:
		style = lipgloss.NewStyle().Foreground(m.palette.Good)
	case LineRemoved:
		style = lipgloss.NewStyle().Foreground(m.palette.Bad)
	case LineContext:
		style = lipgloss.NewStyle().Foreground(m.palette.Text)
	case LineHeader:
		style = lipgloss.NewStyle().Foreground(m.palette.Muted)
	default:
		style = lipgloss.NewStyle().Foreground(m.palette.Text)
	}

	return style.Render(line.Content)
}

// getStatusStyle returns a style for a file status.
func (m *Model) getStatusStyle(status string) lipgloss.Style {
	switch status {
	case "added":
		return lipgloss.NewStyle().Foreground(m.palette.Good)
	case "deleted":
		return lipgloss.NewStyle().Foreground(m.palette.Bad)
	case "renamed":
		return lipgloss.NewStyle().Foreground(m.palette.Warn)
	default:
		return lipgloss.NewStyle().Foreground(m.palette.Accent)
	}
}

// =============================================================================
// Footer
// =============================================================================

// renderFooter renders the bottom footer/help bar.
func (m Model) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(footerHeight).
		Padding(0, 1).
		Background(m.palette.Panel)

	keyStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	// If in search mode, show search input
	if m.searchMode {
		searchPrompt := keyStyle.Render("/") + descStyle.Render(m.searchQuery) + keyStyle.Render("_")
		helpText := descStyle.Render("Enter to search, Esc to cancel")
		return footerStyle.Render(searchPrompt + "  " + helpText)
	}

	// Build help text based on current focus
	var helpItems []string

	helpItems = append(helpItems,
		keyStyle.Render("tab")+descStyle.Render(" switch"),
		keyStyle.Render("j/k")+descStyle.Render(" scroll"),
	)

	if m.focus == FocusFileTree {
		helpItems = append(helpItems,
			keyStyle.Render("enter")+descStyle.Render(" select"),
		)
	} else {
		helpItems = append(helpItems,
			keyStyle.Render("[/]")+descStyle.Render(" files"),
			keyStyle.Render("{/}")+descStyle.Render(" hunks"),
			keyStyle.Render("s")+descStyle.Render(" select"),
			keyStyle.Render("/")+descStyle.Render(" search"),
		)
	}

	// Show search navigation help if search is active
	if len(m.searchMatches) > 0 {
		helpItems = append(helpItems,
			keyStyle.Render("n/N")+descStyle.Render(" next/prev match"),
		)
	} else {
		helpItems = append(helpItems,
			keyStyle.Render("n")+descStyle.Render(" line nums"),
		)
	}

	helpItems = append(helpItems,
		keyStyle.Render("esc")+descStyle.Render(" back"),
	)

	helpText := strings.Join(helpItems, "  ")

	// Add scroll position indicator and selection info
	scrollInfo := ""
	if m.focus == FocusDiff {
		file := m.CurrentFile()
		if file != nil && len(file.Hunks) > 0 {
			// Show hunk position
			hunkInfo := fmt.Sprintf("Hunk %d/%d", m.currentHunkIndex+1, len(file.Hunks))
			scrollInfo = descStyle.Render(hunkInfo)
		}

		// Show search match count
		if len(m.searchMatches) > 0 {
			searchStyle := lipgloss.NewStyle().Foreground(m.palette.Warn)
			scrollInfo += " " + searchStyle.Render(fmt.Sprintf("[%d/%d matches]", m.CurrentMatchDisplay(), m.SearchMatchCount()))
		}

		// Show selection count
		totalSelected := len(m.SelectedHunks())
		if totalSelected > 0 {
			selectStyle := lipgloss.NewStyle().Foreground(m.palette.Good)
			scrollInfo += " " + selectStyle.Render(fmt.Sprintf("[%d selected]", totalSelected))
		}

		scrollPct := m.diffViewport.ScrollPercent()
		scrollInfo += descStyle.Render(fmt.Sprintf(" %d%%", int(scrollPct*100)))
	}

	// Calculate spacing
	spacing := m.width - lipgloss.Width(helpText) - lipgloss.Width(scrollInfo) - 4
	if spacing < 1 {
		spacing = 1
	}

	return footerStyle.Render(helpText + strings.Repeat(" ", spacing) + scrollInfo)
}
