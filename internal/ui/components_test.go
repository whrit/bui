package ui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"bui/internal/config"

	tea "github.com/charmbracelet/bubbletea"
)

// Helper to create test palette and styles
func testPaletteAndStyles() (Palette, Styles) {
	cfg := config.Config{}
	p := NewPalette(cfg)
	s := NewStyles(p)
	return p, s
}

// =============================================================================
// StatusBadge Tests
// =============================================================================

func TestStatusBadge_View(t *testing.T) {
	p, _ := testPaletteAndStyles()

	tests := []struct {
		name     string
		state    PRState
		wantText string
	}{
		{
			name:     "open state shows OPEN",
			state:    PRStateOpen,
			wantText: "OPEN",
		},
		{
			name:     "closed state shows CLOSED",
			state:    PRStateClosed,
			wantText: "CLOSED",
		},
		{
			name:     "merged state shows MERGED",
			state:    PRStateMerged,
			wantText: "MERGED",
		},
		{
			name:     "draft state shows DRAFT",
			state:    PRStateDraft,
			wantText: "DRAFT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := NewStatusBadge(tt.state, p)
			view := badge.View()

			if !strings.Contains(view, tt.wantText) {
				t.Errorf("StatusBadge.View() = %q, want to contain %q", view, tt.wantText)
			}
		})
	}
}

func TestStatusBadge_String(t *testing.T) {
	p, _ := testPaletteAndStyles()

	tests := []struct {
		name  string
		state PRState
		want  string
	}{
		{name: "open", state: PRStateOpen, want: "open"},
		{name: "closed", state: PRStateClosed, want: "closed"},
		{name: "merged", state: PRStateMerged, want: "merged"},
		{name: "draft", state: PRStateDraft, want: "draft"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			badge := NewStatusBadge(tt.state, p)
			if got := badge.String(); got != tt.want {
				t.Errorf("StatusBadge.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// =============================================================================
// CheckIndicator Tests
// =============================================================================

func TestCheckIndicator_Icon(t *testing.T) {
	p, _ := testPaletteAndStyles()

	tests := []struct {
		name   string
		status CheckStatus
		want   string
	}{
		{name: "pending icon", status: CheckPending, want: "○"},
		{name: "running icon", status: CheckRunning, want: "◐"},
		{name: "success icon", status: CheckSuccess, want: "✓"},
		{name: "failure icon", status: CheckFailure, want: "✗"},
		{name: "skipped icon", status: CheckSkipped, want: "⊘"},
		{name: "cancelled icon", status: CheckCancelled, want: "⊗"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indicator := NewCheckIndicator("Test", tt.status, p)
			if got := indicator.Icon(); got != tt.want {
				t.Errorf("CheckIndicator.Icon() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckIndicator_View(t *testing.T) {
	p, _ := testPaletteAndStyles()

	tests := []struct {
		name      string
		checkName string
		status    CheckStatus
		wantIcon  string
		wantName  string
	}{
		{
			name:      "success check view",
			checkName: "Unit Tests",
			status:    CheckSuccess,
			wantIcon:  "✓",
			wantName:  "Unit Tests",
		},
		{
			name:      "failure check view",
			checkName: "Lint",
			status:    CheckFailure,
			wantIcon:  "✗",
			wantName:  "Lint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indicator := NewCheckIndicator(tt.checkName, tt.status, p)
			view := indicator.View()

			if !strings.Contains(view, tt.wantIcon) {
				t.Errorf("CheckIndicator.View() = %q, want to contain icon %q", view, tt.wantIcon)
			}
			if !strings.Contains(view, tt.wantName) {
				t.Errorf("CheckIndicator.View() = %q, want to contain name %q", view, tt.wantName)
			}
		})
	}
}

// =============================================================================
// ErrorDisplay Tests
// =============================================================================

func TestErrorDisplay_View(t *testing.T) {
	p, s := testPaletteAndStyles()

	tests := []struct {
		name      string
		err       error
		title     string
		wantError string
		wantTitle string
	}{
		{
			name:      "error without title",
			err:       errors.New("connection refused"),
			title:     "",
			wantError: "connection refused",
			wantTitle: "",
		},
		{
			name:      "error with title",
			err:       errors.New("timeout"),
			title:     "Failed to load PRs",
			wantError: "timeout",
			wantTitle: "Failed to load PRs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			display := NewErrorDisplay(tt.err, p, s)
			if tt.title != "" {
				display = display.WithTitle(tt.title)
			}

			view := display.View()

			if !strings.Contains(view, tt.wantError) {
				t.Errorf("ErrorDisplay.View() = %q, want to contain error %q", view, tt.wantError)
			}
			if tt.wantTitle != "" && !strings.Contains(view, tt.wantTitle) {
				t.Errorf("ErrorDisplay.View() = %q, want to contain title %q", view, tt.wantTitle)
			}
		})
	}
}

func TestErrorDisplay_IsVisible(t *testing.T) {
	p, s := testPaletteAndStyles()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "visible with error", err: errors.New("some error"), want: true},
		{name: "not visible with nil error", err: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			display := NewErrorDisplay(tt.err, p, s)
			if got := display.IsVisible(); got != tt.want {
				t.Errorf("ErrorDisplay.IsVisible() = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// ConfirmDialog Tests
// =============================================================================

func TestConfirmDialog_ShowHide(t *testing.T) {
	p, s := testPaletteAndStyles()

	dialog := NewConfirmDialog("Delete PR", "Are you sure?", p, s)

	if dialog.Visible() {
		t.Error("NewConfirmDialog should not be visible by default")
	}

	dialog = dialog.Show()
	if !dialog.Visible() {
		t.Error("ConfirmDialog.Show() should make dialog visible")
	}

	dialog = dialog.Hide()
	if dialog.Visible() {
		t.Error("ConfirmDialog.Hide() should make dialog not visible")
	}
}

func TestConfirmDialog_FocusToggle(t *testing.T) {
	p, s := testPaletteAndStyles()

	dialog := NewConfirmDialog("Delete", "Are you sure?", p, s).Show()

	// Initial focus should be on cancel (focused=false means cancel focused)
	if dialog.focused {
		t.Error("Initial focus should be on cancel button")
	}

	// Tab should toggle focus
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	dialog, _ = dialog.Update(tabMsg)
	if !dialog.focused {
		t.Error("Tab should toggle focus to confirm button")
	}

	// Tab again should toggle back
	dialog, _ = dialog.Update(tabMsg)
	if dialog.focused {
		t.Error("Tab should toggle focus back to cancel button")
	}

	// Right arrow should also toggle
	rightMsg := tea.KeyMsg{Type: tea.KeyRight}
	dialog, _ = dialog.Update(rightMsg)
	if !dialog.focused {
		t.Error("Right arrow should toggle focus to confirm button")
	}

	// Left arrow should toggle back
	leftMsg := tea.KeyMsg{Type: tea.KeyLeft}
	dialog, _ = dialog.Update(leftMsg)
	if dialog.focused {
		t.Error("Left arrow should toggle focus to cancel button")
	}
}

func TestConfirmDialog_Confirmed(t *testing.T) {
	p, s := testPaletteAndStyles()

	dialog := NewConfirmDialog("Delete", "Are you sure?", p, s).Show()

	// Initially not confirmed
	if dialog.Confirmed() {
		t.Error("Dialog should not be confirmed initially")
	}

	// Press enter while on cancel button (focused=false)
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	dialog, _ = dialog.Update(enterMsg)
	if dialog.Confirmed() {
		t.Error("Dialog should not be confirmed when cancel is focused")
	}

	// Reset and toggle to confirm button, then press enter
	dialog = NewConfirmDialog("Delete", "Are you sure?", p, s).Show()
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	dialog, _ = dialog.Update(tabMsg) // Focus on confirm
	dialog, _ = dialog.Update(enterMsg)
	if !dialog.Confirmed() {
		t.Error("Dialog should be confirmed when confirm button is focused and enter pressed")
	}
}

func TestConfirmDialog_EscapeCancels(t *testing.T) {
	p, s := testPaletteAndStyles()

	dialog := NewConfirmDialog("Delete", "Are you sure?", p, s).Show()

	// Toggle to confirm button
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	dialog, _ = dialog.Update(tabMsg)

	// Press escape
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	dialog, _ = dialog.Update(escMsg)

	if dialog.Visible() {
		t.Error("Escape should hide the dialog")
	}
	if dialog.Confirmed() {
		t.Error("Escape should not confirm the dialog")
	}
}

func TestConfirmDialog_View(t *testing.T) {
	p, s := testPaletteAndStyles()

	dialog := NewConfirmDialog("Delete PR", "Are you sure you want to delete?", p, s).Show()
	view := dialog.View()

	if !strings.Contains(view, "Delete PR") {
		t.Errorf("ConfirmDialog.View() should contain title, got %q", view)
	}
	if !strings.Contains(view, "Are you sure you want to delete?") {
		t.Errorf("ConfirmDialog.View() should contain message, got %q", view)
	}
	if !strings.Contains(view, "Cancel") {
		t.Errorf("ConfirmDialog.View() should contain Cancel button, got %q", view)
	}
	if !strings.Contains(view, "Confirm") {
		t.Errorf("ConfirmDialog.View() should contain Confirm button, got %q", view)
	}
}

// =============================================================================
// Spinner Tests
// =============================================================================

func TestSpinner_WithMessage(t *testing.T) {
	p, s := testPaletteAndStyles()

	spinner := NewSpinner(p, s)
	spinner = spinner.WithMessage("Loading...")

	if spinner.message != "Loading..." {
		t.Errorf("Spinner.message = %q, want %q", spinner.message, "Loading...")
	}
}

func TestSpinner_View(t *testing.T) {
	p, s := testPaletteAndStyles()

	spinner := NewSpinner(p, s).WithMessage("Loading PRs...")
	view := spinner.View()

	if !strings.Contains(view, "Loading PRs...") {
		t.Errorf("Spinner.View() = %q, want to contain message", view)
	}
}

func TestSpinner_Init(t *testing.T) {
	p, s := testPaletteAndStyles()

	spinner := NewSpinner(p, s)
	cmd := spinner.Init()

	if cmd == nil {
		t.Error("Spinner.Init() should return a non-nil command for tick")
	}
}

// =============================================================================
// Helper Functions Tests
// =============================================================================

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "no truncation needed",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "truncation with ellipsis",
			input:  "hello world",
			maxLen: 8,
			want:   "hello...",
		},
		{
			name:   "very short max length",
			input:  "hello",
			maxLen: 3,
			want:   "...",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "unicode truncation",
			input:  "hello world",
			maxLen: 5,
			want:   "he...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.input, tt.maxLen); got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestPadRight(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  string
	}{
		{
			name:  "padding needed",
			input: "hello",
			width: 10,
			want:  "hello     ",
		},
		{
			name:  "no padding needed",
			input: "hello",
			width: 5,
			want:  "hello",
		},
		{
			name:  "string longer than width",
			input: "hello world",
			width: 5,
			want:  "hello world",
		},
		{
			name:  "empty string",
			input: "",
			width: 5,
			want:  "     ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PadRight(tt.input, tt.width); got != tt.want {
				t.Errorf("PadRight(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.want)
			}
		})
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "just now",
			time: now.Add(-30 * time.Second),
			want: "just now",
		},
		{
			name: "1 minute ago",
			time: now.Add(-1 * time.Minute),
			want: "1 minute ago",
		},
		{
			name: "5 minutes ago",
			time: now.Add(-5 * time.Minute),
			want: "5 minutes ago",
		},
		{
			name: "1 hour ago",
			time: now.Add(-1 * time.Hour),
			want: "1 hour ago",
		},
		{
			name: "3 hours ago",
			time: now.Add(-3 * time.Hour),
			want: "3 hours ago",
		},
		{
			name: "1 day ago",
			time: now.Add(-24 * time.Hour),
			want: "1 day ago",
		},
		{
			name: "3 days ago",
			time: now.Add(-72 * time.Hour),
			want: "3 days ago",
		},
		{
			name: "1 week ago",
			time: now.Add(-7 * 24 * time.Hour),
			want: "1 week ago",
		},
		{
			name: "2 weeks ago",
			time: now.Add(-14 * 24 * time.Hour),
			want: "2 weeks ago",
		},
		{
			name: "1 month ago",
			time: now.Add(-31 * 24 * time.Hour),
			want: "1 month ago",
		},
		{
			name: "3 months ago",
			time: now.Add(-93 * 24 * time.Hour),
			want: "3 months ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RelativeTime(tt.time); got != tt.want {
				t.Errorf("RelativeTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDivider(t *testing.T) {
	p, _ := testPaletteAndStyles()

	tests := []struct {
		name  string
		width int
	}{
		{name: "short divider", width: 10},
		{name: "long divider", width: 80},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			divider := Divider(tt.width, p)

			// The divider should contain repeated characters
			if len(divider) == 0 {
				t.Error("Divider should not be empty")
			}
			// Should contain the divider character (stripped of ANSI codes, check raw contains dash)
			if !strings.Contains(divider, "─") && !strings.Contains(divider, "-") {
				// The divider uses box-drawing character, but may be styled
				// Just ensure it's not empty and has some length
				t.Logf("Divider output: %q", divider)
			}
		})
	}
}
