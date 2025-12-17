package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// =============================================================================
// Header Component
// =============================================================================

// Header is a reusable header component that shows an app title on the left
// and optional context (subtitle) on the right.
type Header struct {
	Title    string
	Subtitle string
	Width    int
	palette  Palette
	styles   Styles
}

// NewHeader creates a new header component with the given palette and styles.
func NewHeader(palette Palette, styles Styles) *Header {
	return &Header{
		Title:    "",
		Subtitle: "",
		Width:    0,
		palette:  palette,
		styles:   styles,
	}
}

// SetTitle sets the header title (left-aligned) and returns the header for chaining.
func (h *Header) SetTitle(title string) *Header {
	h.Title = title
	return h
}

// SetSubtitle sets the header subtitle (right-aligned context) and returns the header for chaining.
func (h *Header) SetSubtitle(subtitle string) *Header {
	h.Subtitle = subtitle
	return h
}

// SetWidth sets the header width and returns the header for chaining.
func (h *Header) SetWidth(width int) *Header {
	h.Width = width
	return h
}

// Render renders the header as a styled string.
func (h *Header) Render() string {
	width := h.Width
	if width <= 0 {
		width = 80 // default width
	}

	// Account for horizontal padding (1 on each side = 2 total)
	const horizontalPadding = 2
	innerWidth := width - horizontalPadding
	if innerWidth < 1 {
		innerWidth = 1
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(h.palette.Text)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(h.palette.Muted)

	title := titleStyle.Render(h.Title)
	subtitle := subtitleStyle.Render(h.Subtitle)

	// Calculate available space for padding between title and subtitle
	titleWidth := lipgloss.Width(title)
	subtitleWidth := lipgloss.Width(subtitle)

	// Create the header line
	var headerContent string
	if h.Subtitle != "" {
		// Space needed between title and subtitle
		padding := innerWidth - titleWidth - subtitleWidth
		if padding < 1 {
			padding = 1
		}
		headerContent = title + strings.Repeat(" ", padding) + subtitle
	} else {
		headerContent = title
	}

	// Apply header container style
	headerStyle := lipgloss.NewStyle().
		Width(width).
		Background(h.palette.Panel).
		Foreground(h.palette.Text).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(h.palette.Border)

	return headerStyle.Render(headerContent)
}

// =============================================================================
// Footer Component
// =============================================================================

// Footer is a reusable footer/status bar component that shows help hints on the left
// and optional status info on the right.
type Footer struct {
	HelpText   string
	StatusText string
	Width      int
	palette    Palette
	styles     Styles
}

// NewFooter creates a new footer component with the given palette and styles.
func NewFooter(palette Palette, styles Styles) *Footer {
	return &Footer{
		HelpText:   "",
		StatusText: "",
		Width:      0,
		palette:    palette,
		styles:     styles,
	}
}

// SetHelp sets the help text (left-aligned) and returns the footer for chaining.
func (f *Footer) SetHelp(help string) *Footer {
	f.HelpText = help
	return f
}

// SetStatus sets the status text (right-aligned) and returns the footer for chaining.
func (f *Footer) SetStatus(status string) *Footer {
	f.StatusText = status
	return f
}

// SetWidth sets the footer width and returns the footer for chaining.
func (f *Footer) SetWidth(width int) *Footer {
	f.Width = width
	return f
}

// Render renders the footer as a styled string.
func (f *Footer) Render() string {
	width := f.Width
	if width <= 0 {
		width = 80 // default width
	}

	helpStyle := lipgloss.NewStyle().
		Foreground(f.palette.Muted)

	statusStyle := lipgloss.NewStyle().
		Foreground(f.palette.Muted)

	help := helpStyle.Render(f.HelpText)
	status := statusStyle.Render(f.StatusText)

	// Calculate widths
	helpWidth := lipgloss.Width(help)
	statusWidth := lipgloss.Width(status)

	// Build footer content
	var footerContent string
	if f.HelpText != "" && f.StatusText != "" {
		// Both help and status
		padding := width - helpWidth - statusWidth - 2 // -2 for container padding
		if padding < 1 {
			padding = 1
		}
		footerContent = help + strings.Repeat(" ", padding) + status
	} else if f.HelpText != "" {
		footerContent = help
	} else if f.StatusText != "" {
		// Right-align the status
		padding := width - statusWidth - 2 // -2 for container padding
		if padding < 0 {
			padding = 0
		}
		footerContent = strings.Repeat(" ", padding) + status
	}

	// Apply footer container style
	footerStyle := lipgloss.NewStyle().
		Width(width).
		Foreground(f.palette.Muted).
		Padding(0, 1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(f.palette.Border)

	return footerStyle.Render(footerContent)
}

// =============================================================================
// ScreenFrame Component
// =============================================================================

// ScreenFrame is a wrapper component that combines header, content area, and footer
// into a complete screen layout.
type ScreenFrame struct {
	header  *Header
	footer  *Footer
	Width   int
	Height  int
	palette Palette
	styles  Styles
}

// NewScreenFrame creates a new screen frame with the given palette and styles.
func NewScreenFrame(palette Palette, styles Styles) *ScreenFrame {
	return &ScreenFrame{
		header:  NewHeader(palette, styles),
		footer:  NewFooter(palette, styles),
		Width:   0,
		Height:  0,
		palette: palette,
		styles:  styles,
	}
}

// SetSize sets the screen frame dimensions and returns the frame for chaining.
func (sf *ScreenFrame) SetSize(width, height int) *ScreenFrame {
	sf.Width = width
	sf.Height = height
	sf.header.SetWidth(width)
	sf.footer.SetWidth(width)
	return sf
}

// SetHeader sets the header title and subtitle and returns the frame for chaining.
func (sf *ScreenFrame) SetHeader(title, subtitle string) *ScreenFrame {
	sf.header.SetTitle(title).SetSubtitle(subtitle)
	return sf
}

// SetFooter sets the footer help and status text and returns the frame for chaining.
func (sf *ScreenFrame) SetFooter(help, status string) *ScreenFrame {
	sf.footer.SetHelp(help).SetStatus(status)
	return sf
}

// ContentHeight returns the available height for content after accounting for header and footer.
// The header and footer each take approximately 2 lines (1 content + 1 border).
func (sf *ScreenFrame) ContentHeight() int {
	// Header: 1 line content + 1 line border = 2 lines
	// Footer: 1 line content + 1 line border = 2 lines
	// Total overhead: 4 lines
	const headerFooterOverhead = 4

	available := sf.Height - headerFooterOverhead
	if available < 0 {
		return 0
	}
	return available
}

// Render renders the screen frame with the given content wrapped by header and footer.
func (sf *ScreenFrame) Render(content string) string {
	width := sf.Width
	if width <= 0 {
		width = 80
	}

	// Render header
	header := sf.header.Render()

	// Render footer
	footer := sf.footer.Render()

	// Calculate content height and constrain if necessary
	contentHeight := sf.ContentHeight()

	// Style the content area
	contentStyle := lipgloss.NewStyle().
		Width(width).
		Height(contentHeight).
		Foreground(sf.palette.Text)

	styledContent := contentStyle.Render(content)

	// Join header, content, and footer vertically
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		styledContent,
		footer,
	)
}

// =============================================================================
// Toast Component
// =============================================================================

// ToastType represents the type/severity of a toast notification.
type ToastType int

const (
	// ToastInfo is for informational messages.
	ToastInfo ToastType = iota
	// ToastSuccess is for success messages.
	ToastSuccess
	// ToastWarning is for warning messages.
	ToastWarning
	// ToastError is for error messages.
	ToastError
)

// Toast is a component for showing transient notification messages.
type Toast struct {
	Message string
	Type    ToastType
	Width   int
	palette Palette
	styles  Styles
}

// NewToast creates a new toast component with the given palette and styles.
func NewToast(palette Palette, styles Styles) *Toast {
	return &Toast{
		Message: "",
		Type:    ToastInfo,
		Width:   0,
		palette: palette,
		styles:  styles,
	}
}

// Info sets the toast as an info message and returns the toast for chaining.
func (t *Toast) Info(msg string) *Toast {
	t.Message = msg
	t.Type = ToastInfo
	return t
}

// Success sets the toast as a success message and returns the toast for chaining.
func (t *Toast) Success(msg string) *Toast {
	t.Message = msg
	t.Type = ToastSuccess
	return t
}

// Warning sets the toast as a warning message and returns the toast for chaining.
func (t *Toast) Warning(msg string) *Toast {
	t.Message = msg
	t.Type = ToastWarning
	return t
}

// Error sets the toast as an error message and returns the toast for chaining.
func (t *Toast) Error(msg string) *Toast {
	t.Message = msg
	t.Type = ToastError
	return t
}

// SetWidth sets the toast width and returns the toast for chaining.
func (t *Toast) SetWidth(width int) *Toast {
	t.Width = width
	return t
}

// IsEmpty returns true if the toast has no message.
func (t *Toast) IsEmpty() bool {
	return t.Message == ""
}

// Render renders the toast as a styled string.
// Returns an empty string if the toast is empty.
func (t *Toast) Render() string {
	if t.IsEmpty() {
		return ""
	}

	// Determine border color based on toast type
	var borderColor lipgloss.Color
	var icon string
	switch t.Type {
	case ToastInfo:
		borderColor = t.palette.Accent
		icon = "i"
	case ToastSuccess:
		borderColor = t.palette.Good
		icon = "+"
	case ToastWarning:
		borderColor = t.palette.Warn
		icon = "!"
	case ToastError:
		borderColor = t.palette.Bad
		icon = "x"
	default:
		borderColor = t.palette.Accent
		icon = "i"
	}

	// Create icon style
	iconStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Bold(true)

	// Create message style
	messageStyle := lipgloss.NewStyle().
		Foreground(t.palette.Text)

	// Build toast content
	content := iconStyle.Render("["+icon+"]") + " " + messageStyle.Render(t.Message)

	// Create toast container style
	toastStyle := lipgloss.NewStyle().
		Padding(0, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor)

	if t.Width > 0 {
		toastStyle = toastStyle.MaxWidth(t.Width)
	}

	return toastStyle.Render(content)
}

// =============================================================================
// Modal Component
// =============================================================================

// Modal is a component for rendering overlay dialogs.
type Modal struct {
	Title   string
	Content string
	Width   int
	Height  int
	palette Palette
	styles  Styles
}

// NewModal creates a new modal component with the given palette and styles.
func NewModal(palette Palette, styles Styles) *Modal {
	return &Modal{
		Title:   "",
		Content: "",
		Width:   0,
		Height:  0,
		palette: palette,
		styles:  styles,
	}
}

// SetTitle sets the modal title and returns the modal for chaining.
func (m *Modal) SetTitle(title string) *Modal {
	m.Title = title
	return m
}

// SetContent sets the modal content and returns the modal for chaining.
func (m *Modal) SetContent(content string) *Modal {
	m.Content = content
	return m
}

// SetSize sets the modal dimensions and returns the modal for chaining.
func (m *Modal) SetSize(width, height int) *Modal {
	m.Width = width
	m.Height = height
	return m
}

// Render renders the modal as a styled bordered box.
func (m *Modal) Render() string {
	width := m.Width
	if width <= 0 {
		width = 50 // default width
	}
	height := m.Height
	if height <= 0 {
		height = 15 // default height
	}

	// Title style
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.palette.Text).
		MarginBottom(1)

	// Content style
	contentStyle := lipgloss.NewStyle().
		Foreground(m.palette.Text)

	// Build modal content
	var modalContent string
	if m.Title != "" {
		modalContent = lipgloss.JoinVertical(lipgloss.Left,
			titleStyle.Render(m.Title),
			contentStyle.Render(m.Content),
		)
	} else {
		modalContent = contentStyle.Render(m.Content)
	}

	// Modal container style with border
	modalStyle := lipgloss.NewStyle().
		Width(width-4).   // Account for border and padding
		Height(height-4). // Account for border and padding
		Padding(1, 2).
		Background(m.palette.Panel).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Border)

	return modalStyle.Render(modalContent)
}

// RenderOverlay renders the modal centered over the given background content.
func (m *Modal) RenderOverlay(background string) string {
	// Get background dimensions
	bgWidth := lipgloss.Width(background)
	bgHeight := lipgloss.Height(background)

	// Use explicit dimensions if set, otherwise use defaults
	modalWidth := m.Width
	if modalWidth <= 0 {
		modalWidth = 50
	}
	modalHeight := m.Height
	if modalHeight <= 0 {
		modalHeight = 15
	}

	// Ensure background has minimum dimensions
	if bgWidth < modalWidth {
		bgWidth = modalWidth
	}
	if bgHeight < modalHeight {
		bgHeight = modalHeight
	}

	// Render the modal
	modal := m.Render()

	// Center the modal within the background space using lipgloss.Place
	return lipgloss.Place(
		bgWidth, bgHeight,
		lipgloss.Center, lipgloss.Center,
		modal,
		lipgloss.WithWhitespaceBackground(m.palette.Bg),
	)
}
