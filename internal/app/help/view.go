package help

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// keyColumnWidth is the width of the key column in help content.
	keyColumnWidth = 14
)

// =============================================================================
// View
// =============================================================================

// View implements tea.Model. It renders the help screen.
func (m Model) View() string {
	// Render header
	header := m.renderHeader()

	// Render viewport content
	content := m.viewport.View()

	// Render footer
	footer := m.renderFooter()

	// Join all parts
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		content,
		footer,
	)
}

// renderHeader renders the help screen header.
func (m Model) renderHeader() string {
	width := m.width
	if width <= 0 {
		width = 80
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.palette.Text).
		Background(m.palette.Panel).
		Padding(0, 1)

	contextStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		Background(m.palette.Panel)

	var title string
	if m.context == HelpContextGlobal {
		title = "Help - bui"
	} else {
		title = fmt.Sprintf("Help - %s", m.context.String())
	}

	titleText := titleStyle.Render(title)

	// Calculate padding for full width
	titleWidth := lipgloss.Width(titleText)
	padding := width - titleWidth - 2
	if padding < 0 {
		padding = 0
	}

	headerStyle := lipgloss.NewStyle().
		Width(width).
		Background(m.palette.Panel).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(m.palette.Border)

	headerContent := titleText + contextStyle.Render(strings.Repeat(" ", padding))

	return headerStyle.Render(headerContent)
}

// renderFooter renders the help screen footer with key hints.
func (m Model) renderFooter() string {
	width := m.width
	if width <= 0 {
		width = 80
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	hints := []string{
		keyStyle.Render("j/k") + descStyle.Render(" scroll"),
		keyStyle.Render("q/esc") + descStyle.Render(" close"),
	}

	// Add scroll position if content is scrollable
	if m.viewport.TotalLineCount() > m.viewport.Height {
		scrollPercent := int(m.viewport.ScrollPercent() * 100)
		hints = append(hints, descStyle.Render(fmt.Sprintf("%d%%", scrollPercent)))
	}

	footerContent := strings.Join(hints, descStyle.Render("  "))

	footerStyle := lipgloss.NewStyle().
		Width(width).
		Foreground(m.palette.Muted).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(m.palette.Border)

	return footerStyle.Render(footerContent)
}

// =============================================================================
// Help Content Generation
// =============================================================================

// generateHelpContent generates the formatted help content string.
func (m Model) generateHelpContent() string {
	var sections []string

	// Global keys section
	sections = append(sections, m.renderGlobalSection())

	// Navigation keys section
	sections = append(sections, m.renderNavigationSection())

	// Context-specific section (if not global)
	if m.context != HelpContextGlobal {
		contextSection := m.renderContextSection()
		if contextSection != "" {
			sections = append(sections, contextSection)
		}
	} else {
		// For global context, show all context sections
		sections = append(sections, m.renderAllContextsSection())
	}

	return strings.Join(sections, "\n\n")
}

// renderGlobalSection renders the global keybindings section.
func (m Model) renderGlobalSection() string {
	sectionStyle := m.getSectionStyle()
	headerStyle := m.getSectionHeaderStyle()

	bindings := []KeyBinding{
		{Key: m.keymap.Quit.Help().Key, Description: "Quit application"},
		{Key: m.keymap.Help.Help().Key, Description: "Show help"},
		{Key: m.keymap.Confirm.Help().Key, Description: "Confirm action"},
		{Key: m.keymap.Cancel.Help().Key, Description: "Cancel / Go back"},
	}

	lines := []string{
		headerStyle.Render("GLOBAL"),
	}

	for _, b := range bindings {
		lines = append(lines, m.formatKeyBinding(b.Key, b.Description))
	}

	return sectionStyle.Render(strings.Join(lines, "\n"))
}

// renderNavigationSection renders the navigation keybindings section.
func (m Model) renderNavigationSection() string {
	sectionStyle := m.getSectionStyle()
	headerStyle := m.getSectionHeaderStyle()

	bindings := []KeyBinding{
		{Key: "j / down", Description: "Move down"},
		{Key: "k / up", Description: "Move up"},
		{Key: "g / Home", Description: "Go to top"},
		{Key: "G / End", Description: "Go to bottom"},
		{Key: "pgdn", Description: "Page down"},
		{Key: "pgup", Description: "Page up"},
	}

	lines := []string{
		headerStyle.Render("NAVIGATION"),
	}

	for _, b := range bindings {
		lines = append(lines, m.formatKeyBinding(b.Key, b.Description))
	}

	return sectionStyle.Render(strings.Join(lines, "\n"))
}

// renderContextSection renders the context-specific keybindings section.
func (m Model) renderContextSection() string {
	bindings := m.getContextBindings()
	if len(bindings) == 0 {
		return ""
	}

	sectionStyle := m.getSectionStyle()
	headerStyle := m.getSectionHeaderStyle()

	lines := []string{
		headerStyle.Render(strings.ToUpper(m.context.String())),
	}

	for _, b := range bindings {
		lines = append(lines, m.formatKeyBinding(b.Key, b.Description))
	}

	return sectionStyle.Render(strings.Join(lines, "\n"))
}

// renderAllContextsSection renders all context-specific sections for global help.
func (m Model) renderAllContextsSection() string {
	var sections []string

	contexts := []HelpContext{
		HelpContextDashboard,
		HelpContextPRDetail,
		HelpContextDiffView,
		HelpContextCreatePR,
		HelpContextReview,
		HelpContextComposer,
	}

	for _, ctx := range contexts {
		// Temporarily switch context to get bindings
		originalContext := m.context
		m.context = ctx
		bindings := m.getContextBindings()
		m.context = originalContext

		if len(bindings) == 0 {
			continue
		}

		sectionStyle := m.getSectionStyle()
		headerStyle := m.getSectionHeaderStyle()

		lines := []string{
			headerStyle.Render(strings.ToUpper(ctx.String())),
		}

		for _, b := range bindings {
			lines = append(lines, m.formatKeyBinding(b.Key, b.Description))
		}

		sections = append(sections, sectionStyle.Render(strings.Join(lines, "\n")))
	}

	return strings.Join(sections, "\n\n")
}

// =============================================================================
// Formatting Helpers
// =============================================================================

// formatKeyBinding formats a single key binding line.
func (m Model) formatKeyBinding(key, description string) string {
	keyStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true).
		Width(keyColumnWidth)

	descStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text)

	return "  " + keyStyle.Render(key) + descStyle.Render(description)
}

// getSectionStyle returns the style for help sections.
func (m Model) getSectionStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.palette.Text)
}

// getSectionHeaderStyle returns the style for section headers.
func (m Model) getSectionHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		Bold(true).
		MarginBottom(1)
}
