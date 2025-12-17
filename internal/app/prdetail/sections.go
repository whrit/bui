package prdetail

import (
	"fmt"
	"strings"

	"bui/internal/gh"
	"bui/internal/ui"

	"github.com/charmbracelet/lipgloss"
)

// =============================================================================
// Info Section
// =============================================================================

// renderInfoSection renders the basic PR information section.
func (m Model) renderInfoSection() string {
	if m.pr == nil {
		return "No PR data available"
	}

	var b strings.Builder

	// Title section
	titleStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true).
		MarginBottom(1)

	b.WriteString(titleStyle.Render(fmt.Sprintf("PR #%d: %s", m.pr.Number, m.pr.Title)))
	b.WriteString("\n\n")

	// Status and basic info
	labelStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		Width(12)

	valueStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text)

	// State badge
	state := ui.PRStateOpen
	if m.pr.IsDraft {
		state = ui.PRStateDraft
	} else if m.pr.State == "CLOSED" {
		state = ui.PRStateClosed
	} else if m.pr.State == "MERGED" {
		state = ui.PRStateMerged
	}
	badge := ui.NewStatusBadge(state, m.palette)

	b.WriteString(labelStyle.Render("Status:"))
	b.WriteString(badge.View())
	b.WriteString("\n")

	// Author
	b.WriteString(labelStyle.Render("Author:"))
	b.WriteString(valueStyle.Render(m.pr.Author.Login))
	b.WriteString("\n")

	// Branch info
	branchStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent)

	b.WriteString(labelStyle.Render("Branch:"))
	b.WriteString(branchStyle.Render(m.pr.HeadBranch))
	b.WriteString(valueStyle.Render(" -> "))
	b.WriteString(branchStyle.Render(m.pr.BaseBranch))
	b.WriteString("\n")

	// Mergeable status
	mergeableStyle := valueStyle
	mergeableText := m.pr.Mergeable
	switch m.pr.Mergeable {
	case "MERGEABLE":
		mergeableStyle = lipgloss.NewStyle().Foreground(m.palette.Good)
		mergeableText = "Mergeable"
	case "CONFLICTING":
		mergeableStyle = lipgloss.NewStyle().Foreground(m.palette.Bad)
		mergeableText = "Conflicts"
	case "UNKNOWN":
		mergeableStyle = lipgloss.NewStyle().Foreground(m.palette.Warn)
		mergeableText = "Unknown"
	}
	b.WriteString(labelStyle.Render("Mergeable:"))
	b.WriteString(mergeableStyle.Render(mergeableText))
	b.WriteString("\n")

	// Labels
	if len(m.pr.Labels) > 0 {
		b.WriteString(labelStyle.Render("Labels:"))
		for i, label := range m.pr.Labels {
			if i > 0 {
				b.WriteString(" ")
			}
			labelBadge := renderLabel(label.Name, label.Color, m.palette)
			b.WriteString(labelBadge)
		}
		b.WriteString("\n")
	}

	// Stats
	b.WriteString("\n")
	addStyle := lipgloss.NewStyle().Foreground(m.palette.Good)
	delStyle := lipgloss.NewStyle().Foreground(m.palette.Bad)

	b.WriteString(labelStyle.Render("Changes:"))
	b.WriteString(addStyle.Render(fmt.Sprintf("+%d", m.pr.Additions)))
	b.WriteString(" / ")
	b.WriteString(delStyle.Render(fmt.Sprintf("-%d", m.pr.Deletions)))
	b.WriteString("\n")

	// Timestamps
	b.WriteString(labelStyle.Render("Created:"))
	b.WriteString(valueStyle.Render(ui.RelativeTime(m.pr.CreatedAt)))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Updated:"))
	b.WriteString(valueStyle.Render(ui.RelativeTime(m.pr.UpdatedAt)))
	b.WriteString("\n")

	// URL
	b.WriteString("\n")
	urlStyle := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		Italic(true)
	b.WriteString(urlStyle.Render(m.pr.URL))

	return b.String()
}

// renderLabel creates a styled label badge.
func renderLabel(name, color string, p ui.Palette) string {
	// Try to use the label color if valid, otherwise use accent
	bg := p.Accent
	if color != "" && len(color) >= 6 {
		bg = lipgloss.Color("#" + color)
	}

	style := lipgloss.NewStyle().
		Background(bg).
		Foreground(lipgloss.Color("#FFFFFF")).
		Padding(0, 1).
		Bold(true)

	return style.Render(name)
}

// =============================================================================
// Description Section
// =============================================================================

// renderDescriptionSection renders the PR description with optional LLM summary.
func (m Model) renderDescriptionSection() string {
	var b strings.Builder

	// LLM Summary (if available)
	if m.summary != "" {
		summaryHeader := lipgloss.NewStyle().
			Foreground(m.palette.Accent).
			Bold(true).
			MarginBottom(1)

		summaryStyle := lipgloss.NewStyle().
			Foreground(m.palette.Text).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(m.palette.Accent).
			Padding(1, 2).
			MarginBottom(1)

		b.WriteString(summaryHeader.Render("AI Summary"))
		b.WriteString("\n")
		b.WriteString(summaryStyle.Render(m.summary))
		b.WriteString("\n\n")
	}

	// PR Body
	bodyHeader := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true).
		MarginBottom(1)

	b.WriteString(bodyHeader.Render("Description"))
	b.WriteString("\n\n")

	if m.pr == nil || m.pr.Body == "" {
		emptyStyle := lipgloss.NewStyle().
			Foreground(m.palette.Muted).
			Italic(true)
		b.WriteString(emptyStyle.Render("No description provided"))
	} else {
		// Render markdown body
		// For now, we render it as plain text with basic formatting
		// In production, we would use glamour for proper markdown rendering
		bodyStyle := lipgloss.NewStyle().
			Foreground(m.palette.Text)
		b.WriteString(bodyStyle.Render(m.pr.Body))
	}

	return b.String()
}

// =============================================================================
// Checks Section
// =============================================================================

// renderChecksSection renders the CI checks status section.
func (m Model) renderChecksSection() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true).
		MarginBottom(1)

	b.WriteString(headerStyle.Render("CI Checks"))
	b.WriteString("\n\n")

	if len(m.checks) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(m.palette.Muted).
			Italic(true)
		b.WriteString(emptyStyle.Render("No checks found"))
		return b.String()
	}

	// Summary
	passed, failed, pending := countCheckStatus(m.checks)
	summaryStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		MarginBottom(1)

	summary := fmt.Sprintf("%d passed, %d failed, %d pending",
		passed, failed, pending)
	b.WriteString(summaryStyle.Render(summary))
	b.WriteString("\n\n")

	// Individual checks
	for _, check := range m.checks {
		status := mapCheckStatus(check)
		indicator := ui.NewCheckIndicator(check.Name, status, m.palette)
		b.WriteString(indicator.View())

		// Add conclusion if available
		if check.Conclusion != "" && check.Conclusion != "success" {
			conclusionStyle := lipgloss.NewStyle().
				Foreground(m.palette.Muted).
				Italic(true)
			b.WriteString(conclusionStyle.Render(fmt.Sprintf(" (%s)", check.Conclusion)))
		}

		b.WriteString("\n")
	}

	return b.String()
}

// countCheckStatus counts checks by status.
func countCheckStatus(checks []gh.Check) (passed, failed, pending int) {
	for _, check := range checks {
		switch check.Conclusion {
		case "success":
			passed++
		case "failure", "timed_out", "cancelled":
			failed++
		default:
			if check.Status == "completed" {
				passed++ // neutral, skipped count as passed
			} else {
				pending++
			}
		}
	}
	return
}

// mapCheckStatus converts gh.Check status to ui.CheckStatus.
func mapCheckStatus(check gh.Check) ui.CheckStatus {
	// First check conclusion (for completed checks)
	switch check.Conclusion {
	case "success":
		return ui.CheckSuccess
	case "failure":
		return ui.CheckFailure
	case "skipped":
		return ui.CheckSkipped
	case "cancelled":
		return ui.CheckCancelled
	}

	// Then check status (for in-progress checks)
	switch check.Status {
	case "in_progress":
		return ui.CheckRunning
	case "queued":
		return ui.CheckPending
	case "completed":
		return ui.CheckSuccess // default for completed without conclusion
	default:
		return ui.CheckPending
	}
}

// =============================================================================
// Reviews Section
// =============================================================================

// renderReviewsSection renders the reviews section.
func (m Model) renderReviewsSection() string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text).
		Bold(true).
		MarginBottom(1)

	b.WriteString(headerStyle.Render("Reviews"))
	b.WriteString("\n\n")

	// Show requested reviewers first
	if m.pr != nil && len(m.pr.Reviewers) > 0 {
		subHeader := lipgloss.NewStyle().
			Foreground(m.palette.Muted).
			Bold(true)
		b.WriteString(subHeader.Render("Requested Reviewers"))
		b.WriteString("\n")

		for _, reviewer := range m.pr.Reviewers {
			b.WriteString(renderReviewer(reviewer.Login, reviewer.State, m.palette))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Show submitted reviews
	if len(m.reviews) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(m.palette.Muted).
			Italic(true)
		b.WriteString(emptyStyle.Render("No reviews submitted yet"))
		return b.String()
	}

	subHeader := lipgloss.NewStyle().
		Foreground(m.palette.Muted).
		Bold(true)
	b.WriteString(subHeader.Render("Submitted Reviews"))
	b.WriteString("\n")

	for _, review := range m.reviews {
		b.WriteString(renderReview(review, m.palette))
		b.WriteString("\n")
	}

	return b.String()
}

// renderReviewer renders a reviewer with their state.
func renderReviewer(login, state string, p ui.Palette) string {
	icon, style := getReviewStateStyle(state, p)

	nameStyle := lipgloss.NewStyle().Foreground(p.Text)

	return style.Render(icon) + " " + nameStyle.Render(login)
}

// renderReview renders a single review entry.
func renderReview(review gh.Review, p ui.Palette) string {
	icon, style := getReviewStateStyle(review.State, p)

	nameStyle := lipgloss.NewStyle().Foreground(p.Text)
	timeStyle := lipgloss.NewStyle().Foreground(p.Muted).Italic(true)

	result := style.Render(icon) + " " +
		nameStyle.Render(review.Author.Login) + " " +
		timeStyle.Render(ui.RelativeTime(review.CreatedAt))

	// Add review body snippet if present
	if review.Body != "" {
		bodyStyle := lipgloss.NewStyle().
			Foreground(p.Muted).
			PaddingLeft(3)
		body := ui.Truncate(review.Body, 80)
		result += "\n" + bodyStyle.Render(body)
	}

	return result
}

// getReviewStateStyle returns the icon and style for a review state.
func getReviewStateStyle(state string, p ui.Palette) (string, lipgloss.Style) {
	switch state {
	case "APPROVED":
		return "✓", lipgloss.NewStyle().Foreground(p.Good)
	case "CHANGES_REQUESTED":
		return "✗", lipgloss.NewStyle().Foreground(p.Bad)
	case "COMMENTED":
		return "◯", lipgloss.NewStyle().Foreground(p.Accent)
	case "PENDING":
		return "○", lipgloss.NewStyle().Foreground(p.Muted)
	case "DISMISSED":
		return "⊘", lipgloss.NewStyle().Foreground(p.Muted)
	default:
		return "?", lipgloss.NewStyle().Foreground(p.Muted)
	}
}
