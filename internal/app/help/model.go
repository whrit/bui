package help

import (
	"bui/internal/config"
	"bui/internal/ui"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// HelpContext Enum
// =============================================================================

// HelpContext represents the screen context for which help is being displayed.
// Different contexts show different relevant keybindings.
type HelpContext int

const (
	// HelpContextGlobal shows all keybindings across all screens.
	HelpContextGlobal HelpContext = iota
	// HelpContextDashboard shows keybindings relevant to the dashboard screen.
	HelpContextDashboard
	// HelpContextPRDetail shows keybindings relevant to the PR detail screen.
	HelpContextPRDetail
	// HelpContextDiffView shows keybindings relevant to the diff viewer.
	HelpContextDiffView
	// HelpContextCreatePR shows keybindings relevant to the PR creation flow.
	HelpContextCreatePR
	// HelpContextReview shows keybindings relevant to the review screen.
	HelpContextReview
	// HelpContextComposer shows keybindings relevant to the text composer.
	HelpContextComposer
)

// String returns a human-readable name for the HelpContext.
func (c HelpContext) String() string {
	switch c {
	case HelpContextGlobal:
		return "Global"
	case HelpContextDashboard:
		return "Dashboard"
	case HelpContextPRDetail:
		return "PR Detail"
	case HelpContextDiffView:
		return "Diff View"
	case HelpContextCreatePR:
		return "Create PR"
	case HelpContextReview:
		return "Review"
	case HelpContextComposer:
		return "Composer"
	default:
		return "Unknown"
	}
}

// =============================================================================
// KeyBinding
// =============================================================================

// KeyBinding represents a single key binding with its key and description.
type KeyBinding struct {
	Key         string
	Description string
}

// =============================================================================
// Model
// =============================================================================

// Model is the Bubble Tea model for the help screen.
type Model struct {
	// Context determines which keybindings to display.
	context HelpContext

	// UI components
	viewport viewport.Model

	// Keymap reference for generating help content.
	keymap ui.KeyMap

	// Styling
	palette ui.Palette
	styles  ui.Styles

	// Dimensions
	width, height int
}

// New creates a new help screen model.
func New(cfg config.Config, context HelpContext) Model {
	palette := ui.NewPalette(cfg)
	styles := ui.NewStyles(palette)
	keymap := ui.NewKeyMap(cfg)

	// Initialize viewport with default size (will be resized on first WindowSizeMsg)
	vp := viewport.New(80, 20)
	vp.Style = styles.App

	m := Model{
		context:  context,
		viewport: vp,
		keymap:   keymap,
		palette:  palette,
		styles:   styles,
		width:    80,
		height:   24,
	}

	// Generate initial content
	m.viewport.SetContent(m.generateHelpContent())

	return m
}

// Init implements tea.Model. Returns nil as no startup commands are needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// SetSize updates the model dimensions and recalculates viewport size.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate viewport dimensions accounting for header and footer
	// Header: ~3 lines (title + border)
	// Footer: ~2 lines (hints + border)
	const headerFooterOverhead = 6

	viewportHeight := height - headerFooterOverhead
	if viewportHeight < 5 {
		viewportHeight = 5
	}

	viewportWidth := width - 4 // Account for border padding
	if viewportWidth < 20 {
		viewportWidth = 20
	}

	m.viewport.Width = viewportWidth
	m.viewport.Height = viewportHeight
	m.viewport.SetContent(m.generateHelpContent())
}

// Context returns the current help context.
func (m Model) Context() HelpContext {
	return m.context
}

// getContextBindings returns keybindings specific to the current context.
func (m Model) getContextBindings() []KeyBinding {
	switch m.context {
	case HelpContextDashboard:
		return []KeyBinding{
			{Key: m.keymap.Refresh.Help().Key, Description: "Refresh PR list"},
			{Key: m.keymap.Search.Help().Key, Description: "Search/filter PRs"},
			{Key: m.keymap.Filter.Help().Key, Description: "Filter by status"},
			{Key: m.keymap.CreatePR.Help().Key, Description: "Create new PR"},
			{Key: m.keymap.Select.Help().Key, Description: "Open PR details"},
			{Key: m.keymap.OpenInBrowser.Help().Key, Description: "Open in browser"},
		}

	case HelpContextPRDetail:
		return []KeyBinding{
			{Key: m.keymap.ViewDiff.Help().Key, Description: "View diff"},
			{Key: m.keymap.MergePR.Help().Key, Description: "Merge PR"},
			{Key: m.keymap.ClosePR.Help().Key, Description: "Close PR"},
			{Key: m.keymap.ApprovePR.Help().Key, Description: "Approve PR"},
			{Key: m.keymap.RequestChanges.Help().Key, Description: "Request changes"},
			{Key: m.keymap.GenerateWithLLM.Help().Key, Description: "Generate AI summary"},
			{Key: m.keymap.OpenInBrowser.Help().Key, Description: "Open in browser"},
			{Key: m.keymap.Refresh.Help().Key, Description: "Refresh data"},
			{Key: "tab", Description: "Switch sections"},
			{Key: "1-4", Description: "Jump to section"},
		}

	case HelpContextDiffView:
		return []KeyBinding{
			{Key: m.keymap.NextItem.Help().Key, Description: "Next file/hunk"},
			{Key: m.keymap.PrevItem.Help().Key, Description: "Previous file/hunk"},
			{Key: m.keymap.Left.Help().Key, Description: "Collapse file"},
			{Key: m.keymap.Right.Help().Key, Description: "Expand file"},
			{Key: "n", Description: "Toggle line numbers"},
			{Key: m.keymap.OpenInBrowser.Help().Key, Description: "Open in browser"},
		}

	case HelpContextCreatePR:
		return []KeyBinding{
			{Key: "tab", Description: "Next field"},
			{Key: "shift+tab", Description: "Previous field"},
			{Key: m.keymap.GenerateWithLLM.Help().Key, Description: "Generate with AI"},
			{Key: m.keymap.Submit.Help().Key, Description: "Create PR"},
			{Key: "d", Description: "Toggle draft mode"},
		}

	case HelpContextReview:
		return []KeyBinding{
			{Key: m.keymap.ApprovePR.Help().Key, Description: "Approve"},
			{Key: m.keymap.RequestChanges.Help().Key, Description: "Request changes"},
			{Key: "c", Description: "Comment only"},
			{Key: m.keymap.Submit.Help().Key, Description: "Submit review"},
			{Key: m.keymap.GenerateWithLLM.Help().Key, Description: "Generate review"},
		}

	case HelpContextComposer:
		return []KeyBinding{
			{Key: m.keymap.GenerateWithLLM.Help().Key, Description: "Generate with AI"},
			{Key: m.keymap.RegenerateLLM.Help().Key, Description: "Regenerate"},
			{Key: m.keymap.AcceptLLM.Help().Key, Description: "Accept suggestion"},
			{Key: m.keymap.EditLLM.Help().Key, Description: "Edit suggestion"},
			{Key: m.keymap.Submit.Help().Key, Description: "Submit"},
			{Key: m.keymap.ClearAll.Help().Key, Description: "Clear all"},
		}

	case HelpContextGlobal:
		// Global context shows no additional context-specific bindings
		// as all bindings are shown in separate sections
		return []KeyBinding{}

	default:
		return []KeyBinding{}
	}
}
