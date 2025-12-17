package dashboard

import (
	"fmt"
	"strings"

	"bui/internal/ui"

	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model. It renders the dashboard view.
func (m Model) View() string {
	// Handle zero dimensions gracefully
	if m.width == 0 || m.height == 0 {
		return ""
	}

	var sections []string

	// Header
	header := m.renderHeader()
	sections = append(sections, header)

	// Main content based on state
	content := m.renderContent()
	sections = append(sections, content)

	// Footer with help
	footer := m.renderFooter()
	sections = append(sections, footer)

	// Join all sections
	view := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// Apply app style
	return m.styles.App.Width(m.width).Height(m.height).Render(view)
}

// =============================================================================
// Section Renderers
// =============================================================================

// renderHeader renders the header section.
func (m Model) renderHeader() string {
	// Title
	titleStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true).
		Padding(0, 1)

	title := titleStyle.Render("Pull Requests")

	// Filter info (if any filters are active)
	filterInfo := m.renderFilterInfo()

	// State indicator
	stateInfo := m.renderStateIndicator()

	// Combine header elements
	leftSide := lipgloss.JoinHorizontal(lipgloss.Left, title, filterInfo)
	rightSide := stateInfo

	// Calculate spacing
	leftWidth := lipgloss.Width(leftSide)
	rightWidth := lipgloss.Width(rightSide)
	spacerWidth := m.width - leftWidth - rightWidth - 2
	if spacerWidth < 0 {
		spacerWidth = 0
	}
	spacer := strings.Repeat(" ", spacerWidth)

	header := leftSide + spacer + rightSide

	// Add divider
	divider := ui.Divider(m.width, m.palette)

	return lipgloss.JoinVertical(lipgloss.Left, header, divider)
}

// renderFilterInfo renders the current filter information.
func (m Model) renderFilterInfo() string {
	var filters []string

	if m.filters.State != "" && m.filters.State != "all" {
		filters = append(filters, m.filters.State)
	}

	if m.filters.Author != "" {
		filters = append(filters, fmt.Sprintf("by %s", m.filters.Author))
	}

	if len(m.filters.Labels) > 0 {
		filters = append(filters, fmt.Sprintf("labels: %s", strings.Join(m.filters.Labels, ", ")))
	}

	if len(filters) == 0 {
		return ""
	}

	filterStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		Padding(0, 1)

	return filterStyle.Render(fmt.Sprintf("(%s)", strings.Join(filters, ", ")))
}

// renderStateIndicator renders the current state indicator.
func (m Model) renderStateIndicator() string {
	style := lipgloss.NewStyle().Padding(0, 1)

	switch m.state {
	case StateLoading:
		return style.Foreground(m.palette.Accent).Render("Loading...")
	case StateFiltering:
		return style.Foreground(m.palette.Accent).Render("Filtering...")
	case StateError:
		return style.Foreground(m.palette.Bad).Render("Error")
	default:
		// Show PR count
		count := len(m.prs)
		return style.Foreground(m.palette.Muted).Render(fmt.Sprintf("%d PRs", count))
	}
}

// renderContent renders the main content area based on current state.
func (m Model) renderContent() string {
	contentHeight := m.height - headerHeight - footerHeight
	if contentHeight < 1 {
		contentHeight = 1
	}

	switch m.state {
	case StateLoading:
		return m.renderLoading(contentHeight)
	case StateError:
		return m.renderError(contentHeight)
	case StateReady, StateFiltering:
		return m.renderList(contentHeight)
	default:
		return ""
	}
}

// renderLoading renders the loading state.
func (m Model) renderLoading(height int) string {
	// Center the spinner vertically
	spinnerView := m.spinner.View()

	// Calculate vertical padding
	spinnerHeight := 1
	topPadding := (height - spinnerHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	// Add padding
	padded := strings.Repeat("\n", topPadding) + spinnerView

	// Center horizontally
	centered := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(padded)

	return centered
}

// renderError renders the error state.
func (m Model) renderError(height int) string {
	if m.err == nil {
		return ""
	}

	errorDisplay := ui.NewErrorDisplay(m.err, m.palette, m.styles).
		WithTitle("Failed to load pull requests")

	errorView := errorDisplay.View()

	// Add helpful hint
	hintStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		MarginTop(1)
	hint := hintStyle.Render("Press 'r' to retry or 'q' to quit")

	content := lipgloss.JoinVertical(lipgloss.Center, errorView, hint)

	// Calculate vertical padding
	contentHeight := lipgloss.Height(content)
	topPadding := (height - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	padded := strings.Repeat("\n", topPadding) + content

	// Center horizontally
	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(padded)
}

// renderList renders the PR list.
func (m Model) renderList(height int) string {
	// Handle empty list
	if len(m.prs) == 0 {
		return m.renderEmptyList(height)
	}

	return m.list.View()
}

// renderEmptyList renders a message when there are no PRs.
func (m Model) renderEmptyList(height int) string {
	messageStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		Align(lipgloss.Center)

	message := messageStyle.Render("No pull requests found")

	// Hint to create new PR
	hintStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		MarginTop(1)
	hint := hintStyle.Render("Press 'c' to create a new pull request")

	content := lipgloss.JoinVertical(lipgloss.Center, message, hint)

	// Calculate vertical padding
	contentHeight := lipgloss.Height(content)
	topPadding := (height - contentHeight) / 2
	if topPadding < 0 {
		topPadding = 0
	}

	padded := strings.Repeat("\n", topPadding) + content

	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(padded)
}

// renderFooter renders the footer with help keys.
func (m Model) renderFooter() string {
	// Divider
	divider := ui.Divider(m.width, m.palette)

	// Help keys based on state
	helpKeys := m.getHelpKeys()

	helpStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		Padding(0, 1)

	keyStyle := m.styles.HelpKey
	descStyle := lipgloss.NewStyle().Foreground(m.palette.Muted)

	var helpItems []string
	for _, h := range helpKeys {
		item := keyStyle.Render(h.key) + descStyle.Render(" "+h.desc)
		helpItems = append(helpItems, item)
	}

	help := helpStyle.Render(strings.Join(helpItems, "  "))

	return lipgloss.JoinVertical(lipgloss.Left, divider, help)
}

// helpItem represents a help key-description pair.
type helpItem struct {
	key  string
	desc string
}

// getHelpKeys returns the help keys for the current state.
func (m Model) getHelpKeys() []helpItem {
	switch m.state {
	case StateLoading:
		return []helpItem{
			{"q", "quit"},
		}
	case StateError:
		return []helpItem{
			{"r", "retry"},
			{"q", "quit"},
		}
	case StateFiltering:
		return []helpItem{
			{"esc", "cancel"},
			{"enter", "apply"},
			{"q", "quit"},
		}
	case StateReady:
		return []helpItem{
			{"j/k", "navigate"},
			{"enter", "select"},
			{"r", "refresh"},
			{"/", "search"},
			{"c", "create"},
			{"q", "quit"},
		}
	default:
		return []helpItem{
			{"q", "quit"},
		}
	}
}

