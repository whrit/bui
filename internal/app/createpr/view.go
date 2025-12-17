package createpr

import (
	"fmt"
	"strings"

	"bui/internal/ui"

	"github.com/charmbracelet/lipgloss"
)

// View implements tea.Model. It renders the create PR wizard.
func (m Model) View() string {
	switch m.step {
	case StepBranchSelect:
		return m.renderBranchSelectView()
	case StepDraftGeneration:
		return m.renderDraftGenerationView()
	case StepEditTitle:
		return m.renderEditTitleView()
	case StepEditBody:
		return m.renderEditBodyView()
	case StepReview:
		return m.renderReviewView()
	case StepSubmitting:
		return m.renderSubmittingView()
	case StepComplete:
		return m.renderCompleteView()
	case StepError:
		return m.renderErrorView()
	default:
		return m.renderBranchSelectView()
	}
}

// =============================================================================
// Header and Footer
// =============================================================================

// renderHeader renders the wizard header with step indicator.
func (m Model) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true)

	stepStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	accentStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true)

	stepNum := m.currentStepNumber()
	total := m.totalSteps()

	var stepIndicator string
	if stepNum > 0 {
		stepIndicator = stepStyle.Render(fmt.Sprintf(" - Step %d of %d", stepNum, total))
	}

	title := titleStyle.Render("Create Pull Request")
	stepName := accentStyle.Render(m.step.DisplayName())

	return title + stepIndicator + "\n" + stepName + "\n"
}

// renderFooter renders the help footer with available key bindings.
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

	switch m.step {
	case StepBranchSelect:
		items = append(items, keyStyle.Render("j/k")+" "+descStyle.Render("navigate"))
		items = append(items, keyStyle.Render("enter")+" "+descStyle.Render("select"))
		items = append(items, keyStyle.Render("tab")+" "+descStyle.Render("switch base/head"))
		items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("cancel"))

	case StepDraftGeneration:
		items = append(items, keyStyle.Render("s")+" "+descStyle.Render("skip"))
		items = append(items, keyStyle.Render("ctrl+x")+" "+descStyle.Render("cancel"))

	case StepEditTitle:
		items = append(items, keyStyle.Render("tab/enter")+" "+descStyle.Render("next"))
		items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("back"))

	case StepEditBody:
		items = append(items, keyStyle.Render("ctrl+s/tab")+" "+descStyle.Render("review"))
		items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("back"))

	case StepReview:
		items = append(items, keyStyle.Render("enter")+" "+descStyle.Render("submit"))
		items = append(items, keyStyle.Render("d")+" "+descStyle.Render("toggle draft"))
		items = append(items, keyStyle.Render("e")+" "+descStyle.Render("edit"))
		items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("back"))

	case StepComplete:
		items = append(items, keyStyle.Render("enter/esc")+" "+descStyle.Render("done"))

	case StepError:
		items = append(items, keyStyle.Render("r")+" "+descStyle.Render("retry"))
		items = append(items, keyStyle.Render("esc")+" "+descStyle.Render("back"))
	}

	return "\n" + strings.Join(items, sep)
}

// =============================================================================
// Branch Selection View
// =============================================================================

// renderBranchSelectView renders the branch selection step.
func (m Model) renderBranchSelectView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Loading state
	if !m.branchesReady {
		b.WriteString(m.spinner.View())
		b.WriteString("\n")
		return b.String()
	}

	// Branch info
	infoStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text)

	labelStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		Width(15)

	selectedStyle := lipgloss.NewStyle().
		Foreground(m.palette.Good).
		Bold(true)

	// Show current selections
	baseFocus := ""
	headFocus := ""
	if m.branchFocus == FocusBaseBranch {
		baseFocus = " [selecting]"
	} else {
		headFocus = " [selecting]"
	}

	b.WriteString(infoStyle.Render(
		labelStyle.Render("Base Branch:") + " " +
			selectedStyle.Render(m.baseBranch) +
			lipgloss.NewStyle().Foreground(m.palette.Accent).Render(baseFocus),
	))
	b.WriteString("\n")

	b.WriteString(infoStyle.Render(
		labelStyle.Render("Head Branch:") + " " +
			selectedStyle.Render(m.headBranch) +
			lipgloss.NewStyle().Foreground(m.palette.Accent).Render(headFocus),
	))
	b.WriteString("\n\n")

	// Commit count
	if len(m.commits) > 0 {
		commitStyle := lipgloss.NewStyle().
			Foreground(m.palette.Muted)

		b.WriteString(commitStyle.Render(fmt.Sprintf("Commits to merge: %d", len(m.commits))))
		b.WriteString("\n")

		// Show a few commit subjects
		maxShow := 5
		if len(m.commits) < maxShow {
			maxShow = len(m.commits)
		}
		for i := 0; i < maxShow; i++ {
			bulletStyle := lipgloss.NewStyle().Foreground(m.palette.Accent)
			subjectStyle := lipgloss.NewStyle().Foreground(m.palette.Text)
			subject := m.commits[i].Subject
			if len(subject) > 60 {
				subject = subject[:57] + "..."
			}
			b.WriteString(bulletStyle.Render("  * ") + subjectStyle.Render(subject) + "\n")
		}
		if len(m.commits) > maxShow {
			b.WriteString(lipgloss.NewStyle().Foreground(m.palette.Muted).Render(
				fmt.Sprintf("  ... and %d more\n", len(m.commits)-maxShow),
			))
		}
		b.WriteString("\n")
	}

	// Branch list
	listStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Border).
		Padding(0, 1)

	b.WriteString(listStyle.Render(m.branchList.View()))

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// =============================================================================
// Draft Generation View
// =============================================================================

// renderDraftGenerationView renders the LLM draft generation step.
func (m Model) renderDraftGenerationView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Spinner and progress
	contentStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Accent).
		Padding(1, 2).
		Width(m.width - 4)

	var content strings.Builder
	content.WriteString(m.spinner.View())
	content.WriteString("\n\n")

	if m.streamedText != "" {
		// Show streamed content with cursor
		streamStyle := lipgloss.NewStyle().
			Foreground(m.palette.Text)

		// Truncate if too long for display
		displayText := m.streamedText
		maxLen := 500
		if len(displayText) > maxLen {
			displayText = displayText[len(displayText)-maxLen:]
			displayText = "..." + displayText
		}

		content.WriteString(streamStyle.Render(displayText))
		content.WriteString(lipgloss.NewStyle().
			Foreground(m.palette.Accent).
			Blink(true).
			Render("|"))
	} else {
		hintStyle := lipgloss.NewStyle().
			Foreground(m.palette.Muted).
			Italic(true)
		content.WriteString(hintStyle.Render("Analyzing commits and generating PR description..."))
	}

	b.WriteString(contentStyle.Render(content.String()))

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// =============================================================================
// Edit Title View
// =============================================================================

// renderEditTitleView renders the title editing step.
func (m Model) renderEditTitleView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Label
	labelStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true)

	b.WriteString(labelStyle.Render("Title:"))
	b.WriteString("\n\n")

	// Input
	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Accent).
		Padding(0, 1)

	b.WriteString(inputStyle.Render(m.titleInput.View()))
	b.WriteString("\n")

	// Character count
	countStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	titleLen := len(m.titleInput.Value())
	b.WriteString(countStyle.Render(fmt.Sprintf("%d/%d characters", titleLen, maxTitleLength)))
	b.WriteString("\n")

	// Preview hint
	if m.generatedBody != "" {
		hintStyle := lipgloss.NewStyle().
			Foreground(m.palette.Muted).
			Italic(true)
		b.WriteString("\n")
		b.WriteString(hintStyle.Render("Body preview (next step):"))
		b.WriteString("\n")

		previewStyle := lipgloss.NewStyle().
			Foreground(m.palette.Text).
			MaxWidth(m.width - 4)

		preview := m.generatedBody
		if len(preview) > 200 {
			preview = preview[:197] + "..."
		}
		b.WriteString(previewStyle.Render(preview))
	}

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// =============================================================================
// Edit Body View
// =============================================================================

// renderEditBodyView renders the body editing step.
func (m Model) renderEditBodyView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Title preview
	titleLabelStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	titleValueStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true)

	b.WriteString(titleLabelStyle.Render("Title: "))
	b.WriteString(titleValueStyle.Render(m.title))
	b.WriteString("\n\n")

	// Body label
	labelStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true)

	b.WriteString(labelStyle.Render("Description:"))
	b.WriteString("\n\n")

	// Textarea
	textareaStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Accent).
		Padding(0, 1)

	b.WriteString(textareaStyle.Render(m.bodyInput.View()))

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// =============================================================================
// Review View
// =============================================================================

// renderReviewView renders the review step before submission.
func (m Model) renderReviewView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Content box
	contentStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Border).
		Padding(1, 2).
		Width(m.width - 4)

	var content strings.Builder

	labelStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted)

	valueStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text)

	accentStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Bold(true)

	// Branch info
	content.WriteString(labelStyle.Render("Merge: "))
	content.WriteString(accentStyle.Render(m.headBranch))
	content.WriteString(labelStyle.Render(" -> "))
	content.WriteString(accentStyle.Render(m.baseBranch))
	content.WriteString("\n\n")

	// Title
	content.WriteString(labelStyle.Render("Title:"))
	content.WriteString("\n")
	content.WriteString(valueStyle.Render(m.title))
	content.WriteString("\n\n")

	// Body
	content.WriteString(labelStyle.Render("Description:"))
	content.WriteString("\n")
	bodyPreview := m.body
	maxBodyLen := 400
	if len(bodyPreview) > maxBodyLen {
		bodyPreview = bodyPreview[:maxBodyLen-3] + "..."
	}
	content.WriteString(valueStyle.Render(bodyPreview))
	content.WriteString("\n\n")

	// Draft mode
	draftLabel := "Draft Mode: "
	draftValue := "Off"
	draftStyle := lipgloss.NewStyle().Foreground(m.palette.Good)
	if m.isDraft {
		draftValue = "On"
		draftStyle = lipgloss.NewStyle().Foreground(m.palette.Warn)
	}
	content.WriteString(labelStyle.Render(draftLabel))
	content.WriteString(draftStyle.Render(draftValue))

	b.WriteString(contentStyle.Render(content.String()))

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// =============================================================================
// Submitting View
// =============================================================================

// renderSubmittingView renders the submission in progress.
func (m Model) renderSubmittingView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Spinner
	contentStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Accent).
		Padding(2, 4).
		Width(m.width - 4).
		Align(lipgloss.Center)

	b.WriteString(contentStyle.Render(m.spinner.View()))

	return b.String()
}

// =============================================================================
// Complete View
// =============================================================================

// renderCompleteView renders the successful completion view.
func (m Model) renderCompleteView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Success message
	successStyle := lipgloss.NewStyle().
		Foreground(m.palette.Good).
		Bold(true)

	contentStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Good).
		Padding(1, 2).
		Width(m.width - 4)

	var content strings.Builder

	content.WriteString(successStyle.Render("Pull Request Created Successfully!"))
	content.WriteString("\n\n")

	if m.createdPR != nil {
		labelStyle := lipgloss.NewStyle().
			Foreground(m.palette.Muted)

		valueStyle := lipgloss.NewStyle().
			Foreground(m.palette.Text)

		linkStyle := lipgloss.NewStyle().
			Foreground(m.palette.Accent).
			Underline(true)

		content.WriteString(labelStyle.Render("PR #"))
		content.WriteString(valueStyle.Render(fmt.Sprintf("%d", m.createdPR.Number)))
		content.WriteString("\n")

		content.WriteString(labelStyle.Render("Title: "))
		content.WriteString(valueStyle.Render(m.createdPR.Title))
		content.WriteString("\n")

		if m.createdPR.URL != "" {
			content.WriteString(labelStyle.Render("URL: "))
			content.WriteString(linkStyle.Render(m.createdPR.URL))
		}
	}

	b.WriteString(contentStyle.Render(content.String()))

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// =============================================================================
// Error View
// =============================================================================

// renderErrorView renders the error state.
func (m Model) renderErrorView() string {
	var b strings.Builder

	// Header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Divider
	b.WriteString(ui.Divider(m.width, m.palette))
	b.WriteString("\n\n")

	// Error display
	errorDisplay := ui.NewErrorDisplay(m.err, m.palette, m.styles).
		WithTitle("Error")

	centeredStyle := lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center)

	b.WriteString(centeredStyle.Render(errorDisplay.View()))

	// Footer
	b.WriteString(m.renderFooter())

	return b.String()
}
