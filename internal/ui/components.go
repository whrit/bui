package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// =============================================================================
// PR State and StatusBadge
// =============================================================================

// PRState represents the state of a pull request.
type PRState string

const (
	PRStateOpen   PRState = "open"
	PRStateClosed PRState = "closed"
	PRStateMerged PRState = "merged"
	PRStateDraft  PRState = "draft"
)

// StatusBadge renders a visual badge for PR state.
type StatusBadge struct {
	state   PRState
	palette Palette
}

// NewStatusBadge creates a new status badge for the given PR state.
func NewStatusBadge(state PRState, p Palette) StatusBadge {
	return StatusBadge{
		state:   state,
		palette: p,
	}
}

// View renders the status badge with appropriate styling.
func (b StatusBadge) View() string {
	var style lipgloss.Style
	var text string

	baseStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true)

	switch b.state {
	case PRStateOpen:
		style = baseStyle.
			Background(b.palette.Good).
			Foreground(lipgloss.Color("#FFFFFF"))
		text = "OPEN"
	case PRStateClosed:
		style = baseStyle.
			Background(b.palette.Bad).
			Foreground(lipgloss.Color("#FFFFFF"))
		text = "CLOSED"
	case PRStateMerged:
		// Purple for merged
		style = baseStyle.
			Background(lipgloss.Color("#A371F7")).
			Foreground(lipgloss.Color("#FFFFFF"))
		text = "MERGED"
	case PRStateDraft:
		style = baseStyle.
			Background(b.palette.Muted).
			Foreground(lipgloss.Color("#FFFFFF"))
		text = "DRAFT"
	default:
		style = baseStyle.
			Background(b.palette.Muted).
			Foreground(lipgloss.Color("#FFFFFF"))
		text = strings.ToUpper(string(b.state))
	}

	return style.Render(text)
}

// String returns the string representation of the state.
func (b StatusBadge) String() string {
	return string(b.state)
}

// =============================================================================
// Check Status and CheckIndicator
// =============================================================================

// CheckStatus represents the status of a CI check.
type CheckStatus string

const (
	CheckPending   CheckStatus = "pending"
	CheckRunning   CheckStatus = "running"
	CheckSuccess   CheckStatus = "success"
	CheckFailure   CheckStatus = "failure"
	CheckSkipped   CheckStatus = "skipped"
	CheckCancelled CheckStatus = "cancelled"
)

// CheckIndicator renders a visual indicator for CI check status.
type CheckIndicator struct {
	name    string
	status  CheckStatus
	palette Palette
}

// NewCheckIndicator creates a new check indicator.
func NewCheckIndicator(name string, status CheckStatus, p Palette) CheckIndicator {
	return CheckIndicator{
		name:    name,
		status:  status,
		palette: p,
	}
}

// Icon returns the status icon/symbol for the check.
func (c CheckIndicator) Icon() string {
	switch c.status {
	case CheckPending:
		return "○"
	case CheckRunning:
		return "◐"
	case CheckSuccess:
		return "✓"
	case CheckFailure:
		return "✗"
	case CheckSkipped:
		return "⊘"
	case CheckCancelled:
		return "⊗"
	default:
		return "?"
	}
}

// View renders the check indicator with icon and name.
func (c CheckIndicator) View() string {
	icon := c.Icon()
	var iconStyle lipgloss.Style

	switch c.status {
	case CheckPending:
		iconStyle = lipgloss.NewStyle().Foreground(c.palette.Muted)
	case CheckRunning:
		iconStyle = lipgloss.NewStyle().Foreground(c.palette.Accent)
	case CheckSuccess:
		iconStyle = lipgloss.NewStyle().Foreground(c.palette.Good)
	case CheckFailure:
		iconStyle = lipgloss.NewStyle().Foreground(c.palette.Bad)
	case CheckSkipped, CheckCancelled:
		iconStyle = lipgloss.NewStyle().Foreground(c.palette.Muted)
	default:
		iconStyle = lipgloss.NewStyle().Foreground(c.palette.Muted)
	}

	nameStyle := lipgloss.NewStyle().Foreground(c.palette.Text)

	return iconStyle.Render(icon) + " " + nameStyle.Render(c.name)
}

// =============================================================================
// Spinner
// =============================================================================

// Spinner is a loading spinner with optional message.
type Spinner struct {
	spinner spinner.Model
	message string
	styles  Styles
	palette Palette
}

// NewSpinner creates a new spinner component.
func NewSpinner(p Palette, s Styles) Spinner {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(p.Accent)

	return Spinner{
		spinner: sp,
		message: "",
		styles:  s,
		palette: p,
	}
}

// WithMessage sets the spinner message and returns a new Spinner.
func (s Spinner) WithMessage(msg string) Spinner {
	s.message = msg
	return s
}

// Init returns the initial command to start the spinner animation.
func (s Spinner) Init() tea.Cmd {
	return s.spinner.Tick
}

// Update handles spinner tick messages.
func (s Spinner) Update(msg tea.Msg) (Spinner, tea.Cmd) {
	var cmd tea.Cmd
	s.spinner, cmd = s.spinner.Update(msg)
	return s, cmd
}

// View renders the spinner with its message.
func (s Spinner) View() string {
	if s.message == "" {
		return s.spinner.View()
	}
	return s.spinner.View() + " " + s.message
}

// Tick returns the spinner tick message type for external use.
func (s Spinner) Tick() tea.Msg {
	return s.spinner.Tick()
}

// =============================================================================
// ErrorDisplay
// =============================================================================

// ErrorDisplay shows error messages inline with styling.
type ErrorDisplay struct {
	err     error
	title   string
	palette Palette
	styles  Styles
}

// NewErrorDisplay creates a new error display component.
func NewErrorDisplay(err error, p Palette, s Styles) ErrorDisplay {
	return ErrorDisplay{
		err:     err,
		title:   "",
		palette: p,
		styles:  s,
	}
}

// WithTitle sets the error title and returns a new ErrorDisplay.
func (e ErrorDisplay) WithTitle(title string) ErrorDisplay {
	e.title = title
	return e
}

// View renders the error display in a bordered box.
func (e ErrorDisplay) View() string {
	if e.err == nil {
		return ""
	}

	errorStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(e.palette.Bad).
		Padding(0, 1).
		Foreground(e.palette.Text)

	titleStyle := lipgloss.NewStyle().
		Foreground(e.palette.Bad).
		Bold(true)

	messageStyle := lipgloss.NewStyle().
		Foreground(e.palette.Text)

	var content string
	if e.title != "" {
		content = titleStyle.Render(e.title) + "\n" + messageStyle.Render(e.err.Error())
	} else {
		content = messageStyle.Render(e.err.Error())
	}

	return errorStyle.Render(content)
}

// IsVisible returns true if the error display has an error to show.
func (e ErrorDisplay) IsVisible() bool {
	return e.err != nil
}

// =============================================================================
// ConfirmDialog
// =============================================================================

// ConfirmDialog is a modal-style confirmation dialog.
type ConfirmDialog struct {
	title     string
	message   string
	confirmed bool
	focused   bool // true = confirm button focused, false = cancel focused
	visible   bool
	palette   Palette
	styles    Styles
}

// NewConfirmDialog creates a new confirmation dialog.
func NewConfirmDialog(title, message string, p Palette, s Styles) ConfirmDialog {
	return ConfirmDialog{
		title:     title,
		message:   message,
		confirmed: false,
		focused:   false, // Start with cancel focused
		visible:   false,
		palette:   p,
		styles:    s,
	}
}

// Show makes the dialog visible.
func (d ConfirmDialog) Show() ConfirmDialog {
	d.visible = true
	d.confirmed = false
	d.focused = false
	return d
}

// Hide hides the dialog.
func (d ConfirmDialog) Hide() ConfirmDialog {
	d.visible = false
	return d
}

// Visible returns whether the dialog is visible.
func (d ConfirmDialog) Visible() bool {
	return d.visible
}

// Confirmed returns whether the user confirmed the action.
func (d ConfirmDialog) Confirmed() bool {
	return d.confirmed
}

// Update handles key messages for the dialog.
func (d ConfirmDialog) Update(msg tea.Msg) (ConfirmDialog, tea.Cmd) {
	if !d.visible {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab, tea.KeyRight, tea.KeyLeft:
			// Toggle focus between cancel and confirm
			d.focused = !d.focused
		case tea.KeyEnter:
			if d.focused {
				// Confirm button is focused
				d.confirmed = true
			}
			d.visible = false
		case tea.KeyEscape:
			d.confirmed = false
			d.visible = false
		}
	}

	return d, nil
}

// View renders the confirmation dialog.
func (d ConfirmDialog) View() string {
	if !d.visible {
		return ""
	}

	// Title style
	titleStyle := lipgloss.NewStyle().
		Foreground(d.palette.Text).
		Bold(true).
		MarginBottom(1)

	// Message style
	messageStyle := lipgloss.NewStyle().
		Foreground(d.palette.Text).
		MarginBottom(1)

	// Button styles
	buttonStyle := lipgloss.NewStyle().
		Padding(0, 2).
		MarginRight(1)

	cancelStyle := buttonStyle
	confirmStyle := buttonStyle

	if d.focused {
		// Confirm is focused
		confirmStyle = confirmStyle.
			Background(d.palette.Accent).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
		cancelStyle = cancelStyle.
			Background(d.palette.Muted).
			Foreground(d.palette.Text)
	} else {
		// Cancel is focused
		cancelStyle = cancelStyle.
			Background(d.palette.Accent).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
		confirmStyle = confirmStyle.
			Background(d.palette.Muted).
			Foreground(d.palette.Text)
	}

	// Build content
	title := titleStyle.Render(d.title)
	message := messageStyle.Render(d.message)
	buttons := cancelStyle.Render("Cancel") + confirmStyle.Render("Confirm")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		message,
		buttons,
	)

	// Dialog box style
	dialogStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(d.palette.Border).
		Background(d.palette.Panel).
		Padding(1, 2)

	return dialogStyle.Render(content)
}

// =============================================================================
// Helper Functions
// =============================================================================

// Truncate truncates a string to maxLen characters, adding ellipsis if needed.
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// PadRight pads a string to the specified width with spaces.
func PadRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// RelativeTime formats a time as a relative time string (e.g., "2 hours ago").
func RelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "just now"
	case diff < 2*time.Minute:
		return "1 minute ago"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d minutes ago", minutes)
	case diff < 2*time.Hour:
		return "1 hour ago"
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("%d hours ago", hours)
	case diff < 48*time.Hour:
		return "1 day ago"
	case diff < 7*24*time.Hour:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d days ago", days)
	case diff < 14*24*time.Hour:
		return "1 week ago"
	case diff < 30*24*time.Hour:
		weeks := int(diff.Hours() / 24 / 7)
		return fmt.Sprintf("%d weeks ago", weeks)
	case diff < 60*24*time.Hour:
		return "1 month ago"
	default:
		months := int(diff.Hours() / 24 / 30)
		return fmt.Sprintf("%d months ago", months)
	}
}

// Divider renders a horizontal divider line of the specified width.
func Divider(width int, p Palette) string {
	dividerChar := "─"
	line := strings.Repeat(dividerChar, width)

	style := lipgloss.NewStyle().
		Foreground(p.Border)

	return style.Render(line)
}
