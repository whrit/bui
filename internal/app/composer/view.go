package composer

import (
	"strings"

	"bui/internal/ui"

	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model. It renders the composer screen.
func (m Model) View() string {
	switch m.state {
	case StateEditing:
		return m.renderEditingView()
	case StateGenerating:
		return m.renderGeneratingView()
	case StateReady:
		return m.renderReadyView()
	case StateError:
		return m.renderErrorView()
	default:
		return m.renderEditingView()
	}
}

// =============================================================================
// Editing View
// =============================================================================

// renderEditingView renders the textarea editing view.
func (m Model) renderEditingView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader("Compose"))
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Textarea
	textareaStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Border).
		Padding(0, 1)

	b.WriteString(textareaStyle.Render(m.textarea.View()))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(m.renderEditingFooter())

	return b.String()
}

// =============================================================================
// Generating View
// =============================================================================

// renderGeneratingView renders the LLM generation view.
func (m Model) renderGeneratingView() string {
	var b strings.Builder

	// Header with generating indicator
	b.WriteString(m.renderHeader("Generating..."))
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Spinner and partial content
	contentStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Accent).
		Padding(1, 2).
		Width(m.width - 4)

	var content strings.Builder
	content.WriteString(m.spinner.View())
	content.WriteString("\n\n")

	if m.streamedText != "" {
		// Show streamed content with a cursor indicator
		streamStyle := lipgloss.NewStyle().
			Foreground(m.palette.Text)
		content.WriteString(streamStyle.Render(m.streamedText))
		content.WriteString(lipgloss.NewStyle().
			Foreground(m.palette.Accent).
			Blink(true).
			Render("|"))
	}

	b.WriteString(contentStyle.Render(content.String()))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(m.renderGeneratingFooter())

	return b.String()
}

// =============================================================================
// Ready View
// =============================================================================

// renderReadyView renders the view when generation is complete.
func (m Model) renderReadyView() string {
	var b strings.Builder

	// Header with ready indicator
	b.WriteString(m.renderHeader("Generated"))
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Show generated content or textarea for editing
	contentStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Good).
		Padding(0, 1)

	b.WriteString(contentStyle.Render(m.textarea.View()))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(m.renderReadyFooter())

	return b.String()
}

// =============================================================================
// Error View
// =============================================================================

// renderErrorView renders the error state.
func (m Model) renderErrorView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader("Error"))
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Error display
	errorDisplay := ui.NewErrorDisplay(m.err, m.palette, m.styles).
		WithTitle("Generation Error")

	centeredStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center)

	b.WriteString(centeredStyle.Render(errorDisplay.View()))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(m.renderErrorFooter())

	return b.String()
}

// =============================================================================
// Header
// =============================================================================

// renderHeader renders the header with a title and state indicator.
func (m Model) renderHeader(title string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true)

	stateStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	var stateIndicator string
	switch m.state {
	case StateEditing:
		stateIndicator = stateStyle.Render("[editing]")
	case StateGenerating:
		stateIndicator = lipgloss.NewStyle().
			Foreground(m.palette.Accent).
			Italic(true).
			Render("[generating]")
	case StateReady:
		stateIndicator = lipgloss.NewStyle().
			Foreground(m.palette.Good).
			Render("[ready]")
	case StateError:
		stateIndicator = lipgloss.NewStyle().
			Foreground(m.palette.Bad).
			Render("[error]")
	}

	return titleStyle.Render(title) + "  " + stateIndicator
}

// =============================================================================
// Footers
// =============================================================================

// renderEditingFooter renders the footer for editing state.
func (m Model) renderEditingFooter() string {
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
	items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("cancel"))

	// LLM actions (if provider available)
	if m.llmProvider != nil {
		items = append(items, keyStyle.Render("ctrl+g")+" "+descStyle.Render("generate"))
	}

	// Submit
	items = append(items, keyStyle.Render("ctrl+s")+" "+descStyle.Render("submit"))

	return strings.Join(items, sep)
}

// renderGeneratingFooter renders the footer during generation.
func (m Model) renderGeneratingFooter() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	sepStyle := lipgloss.NewStyle().
		Foreground(m.palette.Border)

	sep := sepStyle.Render(" | ")

	var items []string

	// Cancel generation
	items = append(items, keyStyle.Render("ctrl+x")+" "+descStyle.Render("cancel generation"))
	items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("cancel"))

	return strings.Join(items, sep)
}

// renderReadyFooter renders the footer when generation is complete.
func (m Model) renderReadyFooter() string {
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
	items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("cancel"))

	// Edit
	items = append(items, keyStyle.Render("e")+" "+descStyle.Render("edit"))

	// Regenerate (if provider available)
	if m.llmProvider != nil {
		items = append(items, keyStyle.Render("ctrl+g")+" "+descStyle.Render("regenerate"))
	}

	// Submit
	items = append(items, keyStyle.Render("ctrl+s")+" "+descStyle.Render("accept"))

	return strings.Join(items, sep)
}

// renderErrorFooter renders the footer during error state.
func (m Model) renderErrorFooter() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	sepStyle := lipgloss.NewStyle().
		Foreground(m.palette.Border)

	sep := sepStyle.Render(" | ")

	var items []string

	// Retry
	items = append(items, keyStyle.Render("r")+" "+descStyle.Render("retry"))

	// Back
	items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("back"))

	return strings.Join(items, sep)
}
