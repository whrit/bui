package prdetail

import (
	"fmt"
	"strings"

	"bui/internal/ui"

	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model. It renders the PR detail screen.
func (m Model) View() string {
	// Handle loading state
	if m.state == StateLoading {
		return m.renderLoadingView()
	}

	// Handle error state
	if m.state == StateError {
		return m.renderErrorView()
	}

	// Handle confirmation dialog
	if m.state == StateConfirming {
		return m.renderConfirmView()
	}

	// Handle merging/closing states
	if m.state == StateMerging || m.state == StateClosing {
		return m.renderActionView()
	}

	// Render normal view
	return m.renderReadyView()
}

// =============================================================================
// Loading View
// =============================================================================

// renderLoadingView renders the loading state.
func (m Model) renderLoadingView() string {
	var b strings.Builder

	// Center the spinner
	spinnerView := m.spinner.View()

	centeredStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	b.WriteString(centeredStyle.Render(spinnerView))

	return b.String()
}

// =============================================================================
// Error View
// =============================================================================

// renderErrorView renders the error state.
func (m Model) renderErrorView() string {
	var b strings.Builder

	errorDisplay := ui.NewErrorDisplay(m.err, m.palette, m.styles).
		WithTitle("Error Loading PR")

	helpStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		MarginTop(2)

	helpText := helpStyle.Render("Press 'r' to retry, 'esc' to go back, 'q' to quit")

	content := lipgloss.JoinVertical(lipgloss.Center,
		errorDisplay.View(),
		helpText,
	)

	centeredStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	b.WriteString(centeredStyle.Render(content))

	return b.String()
}

// =============================================================================
// Confirm View
// =============================================================================

// renderConfirmView renders the confirmation dialog overlay.
func (m Model) renderConfirmView() string {
	// Render the dialog
	dialogView := m.confirmDialog.View()

	// Create the overlay with centered dialog
	overlayStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return overlayStyle.Render(dialogView)
}

// =============================================================================
// Action View
// =============================================================================

// renderActionView renders the view during merge/close operations.
func (m Model) renderActionView() string {
	spinnerView := m.spinner.View()

	centeredStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return centeredStyle.Render(spinnerView)
}

// =============================================================================
// Ready View
// =============================================================================

// renderReadyView renders the normal PR detail view.
func (m Model) renderReadyView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Tab bar
	b.WriteString(m.renderTabBar())
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n")

	// Main content (viewport)
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// =============================================================================
// Header
// =============================================================================

// renderHeader renders the PR header section.
func (m Model) renderHeader() string {
	if m.pr == nil {
		return ""
	}

	var b strings.Builder

	// Title line
	titleStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true)

	prNum := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true).
		Render(fmt.Sprintf("#%d", m.pr.Number))

	b.WriteString(prNum)
	b.WriteString(" ")
	b.WriteString(titleStyle.Render(ui.Truncate(m.pr.Title, m.width-10)))
	b.WriteString("\n")

	// Status line
	state := ui.PRStateOpen
	if m.pr.IsDraft {
		state = ui.PRStateDraft
	} else if m.pr.State == "CLOSED" {
		state = ui.PRStateClosed
	} else if m.pr.State == "MERGED" {
		state = ui.PRStateMerged
	}
	badge := ui.NewStatusBadge(state, m.palette)

	infoStyle := lipgloss.NewStyle().Foreground(m.palette.Muted)
	branchStyle := lipgloss.NewStyle().Foreground(m.palette.Accent)

	b.WriteString(badge.View())
	b.WriteString("  ")
	b.WriteString(infoStyle.Render(m.pr.Author.Login + " wants to merge "))
	b.WriteString(branchStyle.Render(m.pr.HeadBranch))
	b.WriteString(infoStyle.Render(" into "))
	b.WriteString(branchStyle.Render(m.pr.BaseBranch))
	b.WriteString("\n")

	// Stats line
	addStyle := lipgloss.NewStyle().Foreground(m.palette.Good)
	delStyle := lipgloss.NewStyle().Foreground(m.palette.Bad)
	timeStyle := lipgloss.NewStyle().Foreground(m.palette.Muted)

	b.WriteString(addStyle.Render(fmt.Sprintf("+%d", m.pr.Additions)))
	b.WriteString(" ")
	b.WriteString(delStyle.Render(fmt.Sprintf("-%d", m.pr.Deletions)))
	b.WriteString("  ")
	b.WriteString(timeStyle.Render("Updated " + ui.RelativeTime(m.pr.UpdatedAt)))

	// Add check summary if available
	if len(m.checks) > 0 {
		passed, failed, pending := countCheckStatus(m.checks)
		checkStyle := m.palette.Good
		if failed > 0 {
			checkStyle = m.palette.Bad
		} else if pending > 0 {
			checkStyle = m.palette.Warn
		}

		checkSummary := lipgloss.NewStyle().
			Foreground(checkStyle).
			Render(fmt.Sprintf("  Checks: %d/%d", passed, len(m.checks)))
		b.WriteString(checkSummary)
	}

	return b.String()
}

// =============================================================================
// Tab Bar
// =============================================================================

// renderTabBar renders the section tab bar.
func (m Model) renderTabBar() string {
	var tabs []string

	sections := []Section{SectionInfo, SectionDescription, SectionChecks, SectionReviews}

	for i, section := range sections {
		name := fmt.Sprintf("%d:%s", i+1, section.String())

		var style lipgloss.Style
		if section == m.activeSection {
			style = lipgloss.NewStyle().
				Foreground(m.palette.Accent).
				Bold(true).
				Underline(true)
		} else {
			style = lipgloss.NewStyle().
				Foreground(m.palette.Muted)
		}

		tabs = append(tabs, style.Render(name))
	}

	tabBar := strings.Join(tabs, "  ")

	// Add loading indicator if summarizing
	if m.state == StateSummarizing {
		loadingStyle := lipgloss.NewStyle().
			Foreground(m.palette.Accent).
			Italic(true)
		tabBar += "  " + loadingStyle.Render(m.spinner.View())
	}

	return tabBar
}

// =============================================================================
// Footer
// =============================================================================

// renderFooter renders the help footer.
func (m Model) renderFooter() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	sepStyle := lipgloss.NewStyle().
		Foreground(m.palette.Border)

	sep := sepStyle.Render(" | ")

	var items []string

	// Navigation
	items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("back"))
	items = append(items, keyStyle.Render("tab")+" "+descStyle.Render("sections"))
	items = append(items, keyStyle.Render("j/k")+" "+descStyle.Render("scroll"))

	// Actions
	items = append(items, keyStyle.Render("d")+" "+descStyle.Render("diff"))
	items = append(items, keyStyle.Render("o")+" "+descStyle.Render("browser"))

	// LLM (if provider available)
	if m.llmProvider != nil {
		items = append(items, keyStyle.Render("g")+" "+descStyle.Render("summarize"))
	}

	// Conditional actions
	if m.canMerge() {
		items = append(items, keyStyle.Render("m")+" "+descStyle.Render("merge"))
	}
	if m.canClose() {
		items = append(items, keyStyle.Render("x")+" "+descStyle.Render("close"))
	}

	// Refresh
	items = append(items, keyStyle.Render("r")+" "+descStyle.Render("refresh"))

	return strings.Join(items, sep)
}

// =============================================================================
// Viewport Content (delegated to sections.go)
// =============================================================================

// The updateViewportContent method is defined in model.go and calls
// the section renderers defined in sections.go.
