package ui

import (
	"strings"
	"testing"

	"bui/internal/config"

	"github.com/charmbracelet/lipgloss"
)

// =============================================================================
// Header Tests
// =============================================================================

func TestNewHeader(t *testing.T) {
	p, s := testPaletteAndStyles()

	header := NewHeader(p, s)

	if header == nil {
		t.Fatal("NewHeader() returned nil")
	}
	if header.Title != "" {
		t.Errorf("NewHeader().Title = %q, want empty string", header.Title)
	}
	if header.Subtitle != "" {
		t.Errorf("NewHeader().Subtitle = %q, want empty string", header.Subtitle)
	}
}

func TestHeader_SetTitle(t *testing.T) {
	p, s := testPaletteAndStyles()

	header := NewHeader(p, s).SetTitle("Dashboard")

	if header.Title != "Dashboard" {
		t.Errorf("Header.Title = %q, want %q", header.Title, "Dashboard")
	}
}

func TestHeader_SetSubtitle(t *testing.T) {
	p, s := testPaletteAndStyles()

	header := NewHeader(p, s).SetSubtitle("PR #123")

	if header.Subtitle != "PR #123" {
		t.Errorf("Header.Subtitle = %q, want %q", header.Subtitle, "PR #123")
	}
}

func TestHeader_SetWidth(t *testing.T) {
	p, s := testPaletteAndStyles()

	header := NewHeader(p, s).SetWidth(80)

	if header.Width != 80 {
		t.Errorf("Header.Width = %d, want %d", header.Width, 80)
	}
}

func TestHeader_MethodChaining(t *testing.T) {
	p, s := testPaletteAndStyles()

	header := NewHeader(p, s).
		SetTitle("Dashboard").
		SetSubtitle("PR #123").
		SetWidth(100)

	if header.Title != "Dashboard" {
		t.Errorf("Header.Title = %q, want %q", header.Title, "Dashboard")
	}
	if header.Subtitle != "PR #123" {
		t.Errorf("Header.Subtitle = %q, want %q", header.Subtitle, "PR #123")
	}
	if header.Width != 100 {
		t.Errorf("Header.Width = %d, want %d", header.Width, 100)
	}
}

func TestHeader_Render(t *testing.T) {
	p, s := testPaletteAndStyles()

	tests := []struct {
		name          string
		title         string
		subtitle      string
		width         int
		wantTitle     bool
		wantSubtitle  bool
		wantNonEmpty  bool
	}{
		{
			name:         "with title only",
			title:        "bui",
			subtitle:     "",
			width:        80,
			wantTitle:    true,
			wantSubtitle: false,
			wantNonEmpty: true,
		},
		{
			name:         "with title and subtitle",
			title:        "bui",
			subtitle:     "PR #42",
			width:        80,
			wantTitle:    true,
			wantSubtitle: true,
			wantNonEmpty: true,
		},
		{
			name:         "empty header",
			title:        "",
			subtitle:     "",
			width:        80,
			wantTitle:    false,
			wantSubtitle: false,
			wantNonEmpty: true,
		},
		{
			name:         "narrow width",
			title:        "bui",
			subtitle:     "Dashboard",
			width:        20,
			wantTitle:    true,
			wantSubtitle: true,
			wantNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := NewHeader(p, s).
				SetTitle(tt.title).
				SetSubtitle(tt.subtitle).
				SetWidth(tt.width)

			rendered := header.Render()

			if tt.wantNonEmpty && rendered == "" {
				t.Error("Header.Render() returned empty string")
			}
			if tt.wantTitle && !strings.Contains(rendered, tt.title) {
				t.Errorf("Header.Render() = %q, want to contain title %q", rendered, tt.title)
			}
			if tt.wantSubtitle && !strings.Contains(rendered, tt.subtitle) {
				t.Errorf("Header.Render() = %q, want to contain subtitle %q", rendered, tt.subtitle)
			}
		})
	}
}

func TestHeader_RenderEdgeCases(t *testing.T) {
	p, s := testPaletteAndStyles()

	tests := []struct {
		name  string
		width int
	}{
		{name: "zero width", width: 0},
		{name: "negative width", width: -10},
		{name: "very small width", width: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := NewHeader(p, s).
				SetTitle("Test").
				SetWidth(tt.width)

			// Should not panic
			rendered := header.Render()

			// Should return something, even if empty
			_ = rendered
		})
	}
}

// =============================================================================
// Footer Tests
// =============================================================================

func TestNewFooter(t *testing.T) {
	p, s := testPaletteAndStyles()

	footer := NewFooter(p, s)

	if footer == nil {
		t.Fatal("NewFooter() returned nil")
	}
	if footer.HelpText != "" {
		t.Errorf("NewFooter().HelpText = %q, want empty string", footer.HelpText)
	}
	if footer.StatusText != "" {
		t.Errorf("NewFooter().StatusText = %q, want empty string", footer.StatusText)
	}
}

func TestFooter_SetHelp(t *testing.T) {
	p, s := testPaletteAndStyles()

	footer := NewFooter(p, s).SetHelp("j/k: navigate")

	if footer.HelpText != "j/k: navigate" {
		t.Errorf("Footer.HelpText = %q, want %q", footer.HelpText, "j/k: navigate")
	}
}

func TestFooter_SetStatus(t *testing.T) {
	p, s := testPaletteAndStyles()

	footer := NewFooter(p, s).SetStatus("5 PRs")

	if footer.StatusText != "5 PRs" {
		t.Errorf("Footer.StatusText = %q, want %q", footer.StatusText, "5 PRs")
	}
}

func TestFooter_SetWidth(t *testing.T) {
	p, s := testPaletteAndStyles()

	footer := NewFooter(p, s).SetWidth(120)

	if footer.Width != 120 {
		t.Errorf("Footer.Width = %d, want %d", footer.Width, 120)
	}
}

func TestFooter_MethodChaining(t *testing.T) {
	p, s := testPaletteAndStyles()

	footer := NewFooter(p, s).
		SetHelp("q: quit").
		SetStatus("online").
		SetWidth(80)

	if footer.HelpText != "q: quit" {
		t.Errorf("Footer.HelpText = %q, want %q", footer.HelpText, "q: quit")
	}
	if footer.StatusText != "online" {
		t.Errorf("Footer.StatusText = %q, want %q", footer.StatusText, "online")
	}
	if footer.Width != 80 {
		t.Errorf("Footer.Width = %d, want %d", footer.Width, 80)
	}
}

func TestFooter_Render(t *testing.T) {
	p, s := testPaletteAndStyles()

	tests := []struct {
		name       string
		help       string
		status     string
		width      int
		wantHelp   bool
		wantStatus bool
	}{
		{
			name:       "help only",
			help:       "j/k: navigate | enter: select | q: quit",
			status:     "",
			width:      80,
			wantHelp:   true,
			wantStatus: false,
		},
		{
			name:       "help and status",
			help:       "q: quit",
			status:     "Connected",
			width:      80,
			wantHelp:   true,
			wantStatus: true,
		},
		{
			name:       "status only",
			help:       "",
			status:     "5 items",
			width:      80,
			wantHelp:   false,
			wantStatus: true,
		},
		{
			name:       "empty footer",
			help:       "",
			status:     "",
			width:      80,
			wantHelp:   false,
			wantStatus: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			footer := NewFooter(p, s).
				SetHelp(tt.help).
				SetStatus(tt.status).
				SetWidth(tt.width)

			rendered := footer.Render()

			if tt.wantHelp && !strings.Contains(rendered, tt.help) {
				t.Errorf("Footer.Render() = %q, want to contain help %q", rendered, tt.help)
			}
			if tt.wantStatus && !strings.Contains(rendered, tt.status) {
				t.Errorf("Footer.Render() = %q, want to contain status %q", rendered, tt.status)
			}
		})
	}
}

// =============================================================================
// ScreenFrame Tests
// =============================================================================

func TestNewScreenFrame(t *testing.T) {
	p, s := testPaletteAndStyles()

	frame := NewScreenFrame(p, s)

	if frame == nil {
		t.Fatal("NewScreenFrame() returned nil")
	}
	if frame.header == nil {
		t.Error("NewScreenFrame().header should not be nil")
	}
	if frame.footer == nil {
		t.Error("NewScreenFrame().footer should not be nil")
	}
}

func TestScreenFrame_SetSize(t *testing.T) {
	p, s := testPaletteAndStyles()

	frame := NewScreenFrame(p, s).SetSize(100, 50)

	if frame.Width != 100 {
		t.Errorf("ScreenFrame.Width = %d, want %d", frame.Width, 100)
	}
	if frame.Height != 50 {
		t.Errorf("ScreenFrame.Height = %d, want %d", frame.Height, 50)
	}
}

func TestScreenFrame_SetHeader(t *testing.T) {
	p, s := testPaletteAndStyles()

	frame := NewScreenFrame(p, s).SetHeader("App Title", "Context")

	if frame.header.Title != "App Title" {
		t.Errorf("ScreenFrame.header.Title = %q, want %q", frame.header.Title, "App Title")
	}
	if frame.header.Subtitle != "Context" {
		t.Errorf("ScreenFrame.header.Subtitle = %q, want %q", frame.header.Subtitle, "Context")
	}
}

func TestScreenFrame_SetFooter(t *testing.T) {
	p, s := testPaletteAndStyles()

	frame := NewScreenFrame(p, s).SetFooter("Help text", "Status")

	if frame.footer.HelpText != "Help text" {
		t.Errorf("ScreenFrame.footer.HelpText = %q, want %q", frame.footer.HelpText, "Help text")
	}
	if frame.footer.StatusText != "Status" {
		t.Errorf("ScreenFrame.footer.StatusText = %q, want %q", frame.footer.StatusText, "Status")
	}
}

func TestScreenFrame_ContentHeight(t *testing.T) {
	p, s := testPaletteAndStyles()

	tests := []struct {
		name       string
		height     int
		wantMinPos bool // content height should be positive or zero
	}{
		{
			name:       "standard height",
			height:     50,
			wantMinPos: true,
		},
		{
			name:       "small height",
			height:     5,
			wantMinPos: true,
		},
		{
			name:       "zero height",
			height:     0,
			wantMinPos: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := NewScreenFrame(p, s).SetSize(80, tt.height)
			contentHeight := frame.ContentHeight()

			if tt.wantMinPos && contentHeight < 0 {
				t.Errorf("ScreenFrame.ContentHeight() = %d, want >= 0", contentHeight)
			}
			// Content height should be less than total height (accounting for header/footer)
			if tt.height > 0 && contentHeight >= tt.height {
				t.Errorf("ScreenFrame.ContentHeight() = %d, should be less than total height %d", contentHeight, tt.height)
			}
		})
	}
}

func TestScreenFrame_Render(t *testing.T) {
	p, s := testPaletteAndStyles()

	tests := []struct {
		name        string
		title       string
		subtitle    string
		help        string
		status      string
		content     string
		width       int
		height      int
		wantTitle   bool
		wantContent bool
		wantHelp    bool
	}{
		{
			name:        "full frame",
			title:       "bui",
			subtitle:    "Dashboard",
			help:        "q: quit",
			status:      "Ready",
			content:     "Main content here",
			width:       80,
			height:      24,
			wantTitle:   true,
			wantContent: true,
			wantHelp:    true,
		},
		{
			name:        "minimal frame",
			title:       "App",
			subtitle:    "",
			help:        "",
			status:      "",
			content:     "Content",
			width:       40,
			height:      10,
			wantTitle:   true,
			wantContent: true,
			wantHelp:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := NewScreenFrame(p, s).
				SetSize(tt.width, tt.height).
				SetHeader(tt.title, tt.subtitle).
				SetFooter(tt.help, tt.status)

			rendered := frame.Render(tt.content)

			if tt.wantTitle && !strings.Contains(rendered, tt.title) {
				t.Errorf("ScreenFrame.Render() should contain title %q", tt.title)
			}
			if tt.wantContent && !strings.Contains(rendered, tt.content) {
				t.Errorf("ScreenFrame.Render() should contain content %q", tt.content)
			}
			if tt.wantHelp && !strings.Contains(rendered, tt.help) {
				t.Errorf("ScreenFrame.Render() should contain help %q", tt.help)
			}
		})
	}
}

func TestScreenFrame_MethodChaining(t *testing.T) {
	p, s := testPaletteAndStyles()

	frame := NewScreenFrame(p, s).
		SetSize(100, 50).
		SetHeader("Title", "Subtitle").
		SetFooter("Help", "Status")

	if frame.Width != 100 {
		t.Errorf("ScreenFrame.Width = %d, want %d", frame.Width, 100)
	}
	if frame.header.Title != "Title" {
		t.Errorf("ScreenFrame.header.Title = %q, want %q", frame.header.Title, "Title")
	}
	if frame.footer.HelpText != "Help" {
		t.Errorf("ScreenFrame.footer.HelpText = %q, want %q", frame.footer.HelpText, "Help")
	}
}

// =============================================================================
// Toast Tests
// =============================================================================

func TestNewToast(t *testing.T) {
	p, s := testPaletteAndStyles()

	toast := NewToast(p, s)

	if toast == nil {
		t.Fatal("NewToast() returned nil")
	}
	if !toast.IsEmpty() {
		t.Error("NewToast() should be empty")
	}
}

func TestToast_Info(t *testing.T) {
	p, s := testPaletteAndStyles()

	toast := NewToast(p, s).Info("Information message")

	if toast.Message != "Information message" {
		t.Errorf("Toast.Message = %q, want %q", toast.Message, "Information message")
	}
	if toast.Type != ToastInfo {
		t.Errorf("Toast.Type = %v, want %v", toast.Type, ToastInfo)
	}
	if toast.IsEmpty() {
		t.Error("Toast should not be empty after Info()")
	}
}

func TestToast_Success(t *testing.T) {
	p, s := testPaletteAndStyles()

	toast := NewToast(p, s).Success("Operation successful")

	if toast.Message != "Operation successful" {
		t.Errorf("Toast.Message = %q, want %q", toast.Message, "Operation successful")
	}
	if toast.Type != ToastSuccess {
		t.Errorf("Toast.Type = %v, want %v", toast.Type, ToastSuccess)
	}
}

func TestToast_Warning(t *testing.T) {
	p, s := testPaletteAndStyles()

	toast := NewToast(p, s).Warning("Warning message")

	if toast.Message != "Warning message" {
		t.Errorf("Toast.Message = %q, want %q", toast.Message, "Warning message")
	}
	if toast.Type != ToastWarning {
		t.Errorf("Toast.Type = %v, want %v", toast.Type, ToastWarning)
	}
}

func TestToast_Error(t *testing.T) {
	p, s := testPaletteAndStyles()

	toast := NewToast(p, s).Error("Error occurred")

	if toast.Message != "Error occurred" {
		t.Errorf("Toast.Message = %q, want %q", toast.Message, "Error occurred")
	}
	if toast.Type != ToastError {
		t.Errorf("Toast.Type = %v, want %v", toast.Type, ToastError)
	}
}

func TestToast_SetWidth(t *testing.T) {
	p, s := testPaletteAndStyles()

	toast := NewToast(p, s).SetWidth(60)

	if toast.Width != 60 {
		t.Errorf("Toast.Width = %d, want %d", toast.Width, 60)
	}
}

func TestToast_IsEmpty(t *testing.T) {
	p, s := testPaletteAndStyles()

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{name: "empty message", message: "", want: true},
		{name: "non-empty message", message: "Hello", want: false},
		{name: "whitespace only", message: "   ", want: false}, // whitespace is still a message
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toast := NewToast(p, s)
			toast.Message = tt.message

			if got := toast.IsEmpty(); got != tt.want {
				t.Errorf("Toast.IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToast_Render(t *testing.T) {
	p, s := testPaletteAndStyles()

	tests := []struct {
		name        string
		toastType   ToastType
		message     string
		width       int
		wantMessage bool
	}{
		{
			name:        "info toast",
			toastType:   ToastInfo,
			message:     "Info message",
			width:       80,
			wantMessage: true,
		},
		{
			name:        "success toast",
			toastType:   ToastSuccess,
			message:     "Success!",
			width:       80,
			wantMessage: true,
		},
		{
			name:        "warning toast",
			toastType:   ToastWarning,
			message:     "Warning!",
			width:       80,
			wantMessage: true,
		},
		{
			name:        "error toast",
			toastType:   ToastError,
			message:     "Error occurred",
			width:       80,
			wantMessage: true,
		},
		{
			name:        "empty toast",
			toastType:   ToastInfo,
			message:     "",
			width:       80,
			wantMessage: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toast := NewToast(p, s).SetWidth(tt.width)
			toast.Message = tt.message
			toast.Type = tt.toastType

			rendered := toast.Render()

			if tt.wantMessage {
				if !strings.Contains(rendered, tt.message) {
					t.Errorf("Toast.Render() = %q, want to contain message %q", rendered, tt.message)
				}
			} else {
				if rendered != "" {
					t.Errorf("Toast.Render() = %q, want empty string for empty toast", rendered)
				}
			}
		})
	}
}

func TestToast_MethodChaining(t *testing.T) {
	p, s := testPaletteAndStyles()

	toast := NewToast(p, s).
		SetWidth(100).
		Error("Something went wrong")

	if toast.Width != 100 {
		t.Errorf("Toast.Width = %d, want %d", toast.Width, 100)
	}
	if toast.Message != "Something went wrong" {
		t.Errorf("Toast.Message = %q, want %q", toast.Message, "Something went wrong")
	}
	if toast.Type != ToastError {
		t.Errorf("Toast.Type = %v, want %v", toast.Type, ToastError)
	}
}

// =============================================================================
// Modal Tests
// =============================================================================

func TestNewModal(t *testing.T) {
	p, s := testPaletteAndStyles()

	modal := NewModal(p, s)

	if modal == nil {
		t.Fatal("NewModal() returned nil")
	}
	if modal.Title != "" {
		t.Errorf("NewModal().Title = %q, want empty string", modal.Title)
	}
	if modal.Content != "" {
		t.Errorf("NewModal().Content = %q, want empty string", modal.Content)
	}
}

func TestModal_SetTitle(t *testing.T) {
	p, s := testPaletteAndStyles()

	modal := NewModal(p, s).SetTitle("Help")

	if modal.Title != "Help" {
		t.Errorf("Modal.Title = %q, want %q", modal.Title, "Help")
	}
}

func TestModal_SetContent(t *testing.T) {
	p, s := testPaletteAndStyles()

	content := "This is the modal content.\nWith multiple lines."
	modal := NewModal(p, s).SetContent(content)

	if modal.Content != content {
		t.Errorf("Modal.Content = %q, want %q", modal.Content, content)
	}
}

func TestModal_SetSize(t *testing.T) {
	p, s := testPaletteAndStyles()

	modal := NewModal(p, s).SetSize(60, 20)

	if modal.Width != 60 {
		t.Errorf("Modal.Width = %d, want %d", modal.Width, 60)
	}
	if modal.Height != 20 {
		t.Errorf("Modal.Height = %d, want %d", modal.Height, 20)
	}
}

func TestModal_MethodChaining(t *testing.T) {
	p, s := testPaletteAndStyles()

	modal := NewModal(p, s).
		SetTitle("Settings").
		SetContent("Modal content here").
		SetSize(70, 25)

	if modal.Title != "Settings" {
		t.Errorf("Modal.Title = %q, want %q", modal.Title, "Settings")
	}
	if modal.Content != "Modal content here" {
		t.Errorf("Modal.Content = %q, want %q", modal.Content, "Modal content here")
	}
	if modal.Width != 70 {
		t.Errorf("Modal.Width = %d, want %d", modal.Width, 70)
	}
	if modal.Height != 25 {
		t.Errorf("Modal.Height = %d, want %d", modal.Height, 25)
	}
}

func TestModal_Render(t *testing.T) {
	p, s := testPaletteAndStyles()

	tests := []struct {
		name        string
		title       string
		content     string
		width       int
		height      int
		wantTitle   bool
		wantContent bool
	}{
		{
			name:        "full modal",
			title:       "Help",
			content:     "Press j/k to navigate",
			width:       50,
			height:      10,
			wantTitle:   true,
			wantContent: true,
		},
		{
			name:        "no title",
			title:       "",
			content:     "Some content",
			width:       40,
			height:      8,
			wantTitle:   false,
			wantContent: true,
		},
		{
			name:        "multiline content",
			title:       "Info",
			content:     "Line 1\nLine 2\nLine 3",
			width:       60,
			height:      15,
			wantTitle:   true,
			wantContent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modal := NewModal(p, s).
				SetTitle(tt.title).
				SetContent(tt.content).
				SetSize(tt.width, tt.height)

			rendered := modal.Render()

			if rendered == "" {
				t.Error("Modal.Render() returned empty string")
			}
			if tt.wantTitle && tt.title != "" && !strings.Contains(rendered, tt.title) {
				t.Errorf("Modal.Render() = %q, want to contain title %q", rendered, tt.title)
			}
			if tt.wantContent && !strings.Contains(rendered, strings.Split(tt.content, "\n")[0]) {
				t.Errorf("Modal.Render() should contain content")
			}
		})
	}
}

func TestModal_RenderOverlay(t *testing.T) {
	p, s := testPaletteAndStyles()

	tests := []struct {
		name            string
		title           string
		content         string
		modalWidth      int
		modalHeight     int
		backgroundWidth int
		backgroundHeight int
	}{
		{
			name:            "centered modal",
			title:           "Help",
			content:         "Help content",
			modalWidth:      40,
			modalHeight:     10,
			backgroundWidth: 80,
			backgroundHeight: 24,
		},
		{
			name:            "modal same size as background",
			title:           "Full",
			content:         "Full screen modal",
			modalWidth:      80,
			modalHeight:     24,
			backgroundWidth: 80,
			backgroundHeight: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modal := NewModal(p, s).
				SetTitle(tt.title).
				SetContent(tt.content).
				SetSize(tt.modalWidth, tt.modalHeight)

			// Create a simple background
			bgStyle := lipgloss.NewStyle().
				Width(tt.backgroundWidth).
				Height(tt.backgroundHeight)
			background := bgStyle.Render("Background content here")

			rendered := modal.RenderOverlay(background)

			if rendered == "" {
				t.Error("Modal.RenderOverlay() returned empty string")
			}
			// The modal should contain its title
			if !strings.Contains(rendered, tt.title) {
				t.Errorf("Modal.RenderOverlay() should contain modal title %q", tt.title)
			}
		})
	}
}

func TestModal_RenderOverlay_EdgeCases(t *testing.T) {
	p, s := testPaletteAndStyles()

	tests := []struct {
		name       string
		modalW     int
		modalH     int
		background string
	}{
		{
			name:       "empty background",
			modalW:     40,
			modalH:     10,
			background: "",
		},
		{
			name:       "single line background",
			modalW:     20,
			modalH:     5,
			background: "Single line",
		},
		{
			name:       "very small modal",
			modalW:     10,
			modalH:     3,
			background: "Some background\ncontent\nhere",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modal := NewModal(p, s).
				SetTitle("Test").
				SetContent("Content").
				SetSize(tt.modalW, tt.modalH)

			// Should not panic
			rendered := modal.RenderOverlay(tt.background)

			// Should return something
			if rendered == "" {
				t.Error("Modal.RenderOverlay() returned empty string")
			}
		})
	}
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestLayout_ScreenFrameWithToast(t *testing.T) {
	p, s := testPaletteAndStyles()

	// Create a screen frame
	frame := NewScreenFrame(p, s).
		SetSize(80, 24).
		SetHeader("bui", "Dashboard").
		SetFooter("q: quit | ?: help", "5 PRs")

	// Render with some content
	content := "Main content area"
	rendered := frame.Render(content)

	// Create a toast
	toast := NewToast(p, s).
		SetWidth(80).
		Success("Changes saved!")

	toastRendered := toast.Render()

	// Both should render without errors
	if rendered == "" {
		t.Error("ScreenFrame.Render() returned empty")
	}
	if toastRendered == "" {
		t.Error("Toast.Render() returned empty")
	}
}

func TestLayout_ModalOverScreenFrame(t *testing.T) {
	p, s := testPaletteAndStyles()

	// Create screen frame
	frame := NewScreenFrame(p, s).
		SetSize(80, 24).
		SetHeader("bui", "").
		SetFooter("q: quit", "")

	background := frame.Render("Content underneath")

	// Create modal
	modal := NewModal(p, s).
		SetTitle("Help").
		SetContent("j/k: navigate\nenter: select\nq: quit").
		SetSize(40, 10)

	// Render modal over the screen frame
	rendered := modal.RenderOverlay(background)

	if rendered == "" {
		t.Error("Modal.RenderOverlay() over ScreenFrame returned empty")
	}
	if !strings.Contains(rendered, "Help") {
		t.Error("Modal should be visible in overlay")
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkHeader_Render(b *testing.B) {
	cfg := config.Config{}
	p := NewPalette(cfg)
	s := NewStyles(p)

	header := NewHeader(p, s).
		SetTitle("bui").
		SetSubtitle("Dashboard").
		SetWidth(80)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = header.Render()
	}
}

func BenchmarkFooter_Render(b *testing.B) {
	cfg := config.Config{}
	p := NewPalette(cfg)
	s := NewStyles(p)

	footer := NewFooter(p, s).
		SetHelp("q: quit | ?: help | j/k: navigate").
		SetStatus("5 PRs").
		SetWidth(80)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = footer.Render()
	}
}

func BenchmarkScreenFrame_Render(b *testing.B) {
	cfg := config.Config{}
	p := NewPalette(cfg)
	s := NewStyles(p)

	frame := NewScreenFrame(p, s).
		SetSize(80, 24).
		SetHeader("bui", "Dashboard").
		SetFooter("q: quit", "Ready")

	content := "Main content\nWith multiple\nLines of text"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = frame.Render(content)
	}
}

func BenchmarkToast_Render(b *testing.B) {
	cfg := config.Config{}
	p := NewPalette(cfg)
	s := NewStyles(p)

	toast := NewToast(p, s).
		SetWidth(80).
		Success("Operation completed successfully!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = toast.Render()
	}
}

func BenchmarkModal_RenderOverlay(b *testing.B) {
	cfg := config.Config{}
	p := NewPalette(cfg)
	s := NewStyles(p)

	modal := NewModal(p, s).
		SetTitle("Help").
		SetContent("j/k: navigate\nenter: select\nq: quit").
		SetSize(40, 10)

	background := strings.Repeat(strings.Repeat(".", 80)+"\n", 24)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = modal.RenderOverlay(background)
	}
}
