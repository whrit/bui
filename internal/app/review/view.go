package review

import (
	"fmt"
	"strings"

	"bui/internal/ui"

	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model. It renders the review screen.
func (m Model) View() string {
	switch m.state {
	case StateTypeSelect:
		return m.renderTypeSelectView()
	case StateCompose:
		return m.renderComposeView()
	case StatePreview:
		return m.renderPreviewView()
	case StateSubmitting:
		return m.renderSubmittingView()
	case StateComplete:
		return m.renderCompleteView()
	case StateError:
		return m.renderErrorView()
	default:
		return m.renderTypeSelectView()
	}
}

// =============================================================================
// Type Select View
// =============================================================================

// renderTypeSelectView renders the review type selection view.
func (m Model) renderTypeSelectView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader("Submit Review"))
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// PR info
	prInfoStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)
	b.WriteString(prInfoStyle.Render(fmt.Sprintf("PR #%d: %s", m.prNumber, m.prTitle)))
	b.WriteString("\n\n")

	// Selection prompt
	promptStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true)
	b.WriteString(promptStyle.Render("Select Review Type:"))
	b.WriteString("\n\n")

	// Options
	optionStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		PaddingLeft(2)

	selectedKeyStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	// Approve option
	approveIcon := lipgloss.NewStyle().Foreground(m.palette.Good).Render("+")
	b.WriteString(optionStyle.Render(
		selectedKeyStyle.Render("[1]") + " " + approveIcon + " " +
			lipgloss.NewStyle().Foreground(m.palette.Text).Render("Approve") +
			descStyle.Render(" - Approve this pull request"),
	))
	b.WriteString("\n\n")

	// Request Changes option
	changesIcon := lipgloss.NewStyle().Foreground(m.palette.Bad).Render("x")
	b.WriteString(optionStyle.Render(
		selectedKeyStyle.Render("[2]") + " " + changesIcon + " " +
			lipgloss.NewStyle().Foreground(m.palette.Text).Render("Request Changes") +
			descStyle.Render(" - Request changes before merging"),
	))
	b.WriteString("\n\n")

	// Comment option
	commentIcon := lipgloss.NewStyle().Foreground(m.palette.Accent).Render("#")
	b.WriteString(optionStyle.Render(
		selectedKeyStyle.Render("[3]") + " " + commentIcon + " " +
			lipgloss.NewStyle().Foreground(m.palette.Text).Render("Comment") +
			descStyle.Render(" - Leave a comment without approval"),
	))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(m.renderTypeSelectFooter())

	return b.String()
}

// =============================================================================
// Compose View
// =============================================================================

// renderComposeView renders the comment composition view.
func (m Model) renderComposeView() string {
	var b strings.Builder

	// Header
	headerTitle := fmt.Sprintf("Review: %s", m.reviewType.DisplayName())
	b.WriteString(m.renderHeader(headerTitle))
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// PR info
	prInfoStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)
	b.WriteString(prInfoStyle.Render(fmt.Sprintf("PR #%d: %s", m.prNumber, m.prTitle)))
	b.WriteString("\n\n")

	// If generating, show spinner and streamed content
	if m.isGenerating {
		contentStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(m.palette.Accent).
			Padding(1, 2).
			Width(m.width - 4)

		var content strings.Builder
		content.WriteString(m.spinner.View())
		content.WriteString("\n\n")

		if m.generatedComment != "" {
			streamStyle := lipgloss.NewStyle().
				Foreground(m.palette.Text)
			content.WriteString(streamStyle.Render(m.generatedComment))
			content.WriteString(lipgloss.NewStyle().
				Foreground(m.palette.Accent).
				Blink(true).
				Render("|"))
		}

		b.WriteString(contentStyle.Render(content.String()))
		b.WriteString("\n\n")

		// Generating footer
		b.WriteString(m.renderGeneratingFooter())
	} else {
		// Show validation message if needed
		if !m.canProceedToPreview() && m.reviewType != ReviewTypeApprove {
			warnStyle := lipgloss.NewStyle().
				Foreground(m.palette.Warn)
			b.WriteString(warnStyle.Render("A comment is required for this review type."))
			b.WriteString("\n\n")
		}

		// Textarea
		textareaStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(m.palette.Border).
			Padding(0, 1)

		b.WriteString(textareaStyle.Render(m.textarea.View()))
		b.WriteString("\n\n")

		// Footer
		b.WriteString(m.renderComposeFooter())
	}

	return b.String()
}

// =============================================================================
// Preview View
// =============================================================================

// renderPreviewView renders the preview view before submission.
func (m Model) renderPreviewView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader("Preview Review"))
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// PR info
	prInfoStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)
	b.WriteString(prInfoStyle.Render(fmt.Sprintf("PR #%d: %s", m.prNumber, m.prTitle)))
	b.WriteString("\n\n")

	// Review type badge
	badgeStyle := m.getReviewTypeBadgeStyle()
	b.WriteString(badgeStyle.Render(m.reviewType.DisplayName()))
	b.WriteString("\n\n")

	// Comment preview
	if m.comment != "" {
		labelStyle := lipgloss.NewStyle().
			Foreground(m.palette.Muted).
			Bold(true)
		b.WriteString(labelStyle.Render("Comment:"))
		b.WriteString("\n")

		commentStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(m.palette.Border).
			Padding(1, 2).
			Width(m.width - 4)

		b.WriteString(commentStyle.Render(m.comment))
	} else {
		emptyStyle := lipgloss.NewStyle().
			Foreground(m.palette.Muted).
			Italic(true)
		b.WriteString(emptyStyle.Render("(No comment provided)"))
	}
	b.WriteString("\n\n")

	// Footer
	b.WriteString(m.renderPreviewFooter())

	return b.String()
}

// =============================================================================
// Submitting View
// =============================================================================

// renderSubmittingView renders the view during submission.
func (m Model) renderSubmittingView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader("Submitting Review"))
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Spinner
	spinnerStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Padding(2, 0)

	b.WriteString(spinnerStyle.Render(m.spinner.View()))
	b.WriteString("\n")

	return b.String()
}

// =============================================================================
// Complete View
// =============================================================================

// renderCompleteView renders the view after successful submission.
func (m Model) renderCompleteView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader("Review Submitted"))
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Success message
	successStyle := lipgloss.NewStyle().
		Foreground(m.palette.Good).
		Bold(true)

	centerStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Padding(2, 0)

	successMsg := fmt.Sprintf("Successfully submitted %s review for PR #%d",
		m.reviewType.DisplayName(), m.prNumber)
	b.WriteString(centerStyle.Render(successStyle.Render(successMsg)))
	b.WriteString("\n\n")

	// Footer
	b.WriteString(m.renderCompleteFooter())

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
		WithTitle("Submission Error")

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

	prStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent)

	return titleStyle.Render(title) + "  " + prStyle.Render(fmt.Sprintf("PR #%d", m.prNumber))
}

// =============================================================================
// Footers
// =============================================================================

// renderTypeSelectFooter renders the footer for type selection state.
func (m Model) renderTypeSelectFooter() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	sepStyle := lipgloss.NewStyle().
		Foreground(m.palette.Border)

	sep := sepStyle.Render(" | ")

	var items []string
	items = append(items, keyStyle.Render("1/a")+" "+descStyle.Render("approve"))
	items = append(items, keyStyle.Render("2/c")+" "+descStyle.Render("changes"))
	items = append(items, keyStyle.Render("3/m")+" "+descStyle.Render("comment"))
	items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("cancel"))

	return strings.Join(items, sep)
}

// renderComposeFooter renders the footer for compose state.
func (m Model) renderComposeFooter() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	sepStyle := lipgloss.NewStyle().
		Foreground(m.palette.Border)

	sep := sepStyle.Render(" | ")

	var items []string
	items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("back"))

	// LLM actions (if provider available)
	if m.llmProvider != nil {
		items = append(items, keyStyle.Render("ctrl+g")+" "+descStyle.Render("generate"))
	}

	items = append(items, keyStyle.Render("ctrl+s")+" "+descStyle.Render("preview"))

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
	items = append(items, keyStyle.Render("ctrl+x")+" "+descStyle.Render("cancel generation"))
	items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("cancel"))

	return strings.Join(items, sep)
}

// renderPreviewFooter renders the footer for preview state.
func (m Model) renderPreviewFooter() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	sepStyle := lipgloss.NewStyle().
		Foreground(m.palette.Border)

	sep := sepStyle.Render(" | ")

	var items []string
	items = append(items, keyStyle.Render("enter")+" "+descStyle.Render("submit"))
	items = append(items, keyStyle.Render("e")+" "+descStyle.Render("edit"))
	items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("back"))

	return strings.Join(items, sep)
}

// renderCompleteFooter renders the footer for complete state.
func (m Model) renderCompleteFooter() string {
	keyStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	return keyStyle.Render("enter") + " " + descStyle.Render("back to PR")
}

// renderErrorFooter renders the footer for error state.
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
	items = append(items, keyStyle.Render("r")+" "+descStyle.Render("retry"))
	items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("back"))

	return strings.Join(items, sep)
}

// =============================================================================
// Helpers
// =============================================================================

// getReviewTypeBadgeStyle returns the badge style for the current review type.
func (m Model) getReviewTypeBadgeStyle() lipgloss.Style {
	baseStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true)

	switch m.reviewType {
	case ReviewTypeApprove:
		return baseStyle.
			Background(m.palette.Good).
			Foreground(lipgloss.Color("#FFFFFF"))
	case ReviewTypeRequestChanges:
		return baseStyle.
			Background(m.palette.Bad).
			Foreground(lipgloss.Color("#FFFFFF"))
	case ReviewTypeComment:
		return baseStyle.
			Background(m.palette.Accent).
			Foreground(lipgloss.Color("#FFFFFF"))
	default:
		return baseStyle.
			Background(m.palette.Muted).
			Foreground(lipgloss.Color("#FFFFFF"))
	}
}
