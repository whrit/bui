// Package root implements the root/shell model for the bui application.
// It manages screen navigation between Dashboard, PR Detail, Diff Viewer,
// Composer, Create PR, Review, and Help screens.
package root

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"bui/internal/app/createpr"
	"bui/internal/app/dashboard"
	"bui/internal/app/diffview"
	"bui/internal/app/help"
	"bui/internal/app/prdetail"
	"bui/internal/app/review"
	"bui/internal/config"
	"bui/internal/gh"
	"bui/internal/git"
	"bui/internal/llm"
	"bui/internal/ui"
)

// =============================================================================
// Screen Type
// =============================================================================

// Screen represents the active screen in the application.
type Screen int

const (
	// ScreenDashboard is the main PR listing screen.
	ScreenDashboard Screen = iota
	// ScreenPRDetail shows details for a single PR.
	ScreenPRDetail
	// ScreenDiffView shows the diff for a PR.
	ScreenDiffView
	// ScreenCreatePR shows the create PR wizard.
	ScreenCreatePR
	// ScreenReview shows the review submission screen.
	ScreenReview
	// ScreenHelp shows the help/keybindings screen.
	ScreenHelp
)

// String returns a string representation of the screen.
func (s Screen) String() string {
	switch s {
	case ScreenDashboard:
		return "dashboard"
	case ScreenPRDetail:
		return "prdetail"
	case ScreenDiffView:
		return "diffview"
	case ScreenCreatePR:
		return "createpr"
	case ScreenReview:
		return "review"
	case ScreenHelp:
		return "help"
	default:
		return "unknown"
	}
}

// =============================================================================
// Toast Message Types
// =============================================================================

// ShowToastMsg displays a toast notification.
type ShowToastMsg struct {
	Message string
	Type    ui.ToastType
}

// ClearToastMsg clears the current toast.
type ClearToastMsg struct{}

// ToastTickMsg is sent for toast auto-dismiss countdown.
type ToastTickMsg struct{}

// toastDuration is the number of ticks before a toast auto-dismisses.
// At 100ms per tick, 30 ticks = 3 seconds.
const toastDuration = 30

// =============================================================================
// Model
// =============================================================================

// Model is the root Bubble Tea model that orchestrates screen navigation.
type Model struct {
	// Configuration
	cfg     config.Config
	cfgPath string

	// Screen state
	screen         Screen
	previousScreen Screen // To return to after help closes

	// Screen models
	dashboard dashboard.Model
	prdetail  prdetail.Model
	diffview  diffview.Model
	createpr  createpr.Model
	review    review.Model
	help      help.Model

	// Dependencies
	ghClient    *gh.Client
	gitRepo     *git.Repo
	llmProvider llm.Provider

	// Dimensions
	width, height int

	// Toast state
	toast      *ui.Toast
	toastTimer int // Countdown ticks until toast clears

	// Styling
	palette ui.Palette
	styles  ui.Styles
	keymap  ui.KeyMap
}

// New creates a new root model with the dashboard as the starting screen.
func New(cfg config.Config, cfgPath string) Model {
	pal := ui.NewPalette(cfg)
	styles := ui.NewStyles(pal)
	keymap := ui.NewKeyMap(cfg)

	// Initialize dashboard as the starting screen
	dash := dashboard.New(cfg)

	// Initialize dependencies
	ghClient := gh.New()
	gitRepo := git.New()

	// Initialize LLM provider if configured
	var llmProvider llm.Provider
	if cfg.LLM.Provider != "" {
		llmProvider, _ = llm.NewProvider(cfg)
	}

	return Model{
		cfg:         cfg,
		cfgPath:     cfgPath,
		screen:      ScreenDashboard,
		dashboard:   dash,
		ghClient:    ghClient,
		gitRepo:     gitRepo,
		llmProvider: llmProvider,
		palette:     pal,
		styles:      styles,
		keymap:      keymap,
	}
}

// Init implements tea.Model. It initializes the active screen.
func (m Model) Init() tea.Cmd {
	// Initialize the dashboard (the starting screen)
	return m.dashboard.Init()
}

// Update implements tea.Model. It handles messages and routes them to the active screen.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.updateScreenSizes()

	case tea.KeyMsg:
		// Global quit handler - check before routing to screens
		if msg.String() == m.cfg.Keys.Quit || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Global help handler - open help from any screen (except help itself)
		if msg.String() == m.cfg.Keys.Help && m.screen != ScreenHelp {
			// Cleanup prdetail if we're navigating away from it
			if m.screen == ScreenPRDetail {
				m.prdetail.Cleanup()
			}
			m.previousScreen = m.screen
			m.screen = ScreenHelp
			m.help = help.New(m.cfg, screenToHelpContext(m.previousScreen))
			m.help.SetSize(m.width, m.height)
			return m, m.help.Init()
		}

	// ==========================================================================
	// Toast Messages
	// ==========================================================================

	case ShowToastMsg:
		m.toast = ui.NewToast(m.palette, m.styles)
		switch msg.Type {
		case ui.ToastSuccess:
			m.toast.Success(msg.Message)
		case ui.ToastError:
			m.toast.Error(msg.Message)
		case ui.ToastWarning:
			m.toast.Warning(msg.Message)
		default:
			m.toast.Info(msg.Message)
		}
		m.toastTimer = toastDuration
		return m, toastTick()

	case ToastTickMsg:
		if m.toastTimer > 0 {
			m.toastTimer--
			if m.toastTimer == 0 {
				m.toast = nil
				return m, nil
			}
			return m, toastTick()
		}
		return m, nil

	case ClearToastMsg:
		m.toast = nil
		m.toastTimer = 0
		return m, nil

	// ==========================================================================
	// Help Screen Navigation
	// ==========================================================================

	case help.CloseHelpMsg:
		// Return to the previous screen
		m.screen = m.previousScreen
		return m, nil

	// ==========================================================================
	// Navigation messages from Dashboard
	// ==========================================================================

	case dashboard.OpenPRDetailMsg:
		m.screen = ScreenPRDetail
		m.prdetail = prdetail.New(m.cfg, msg.PR.Number)
		m.prdetail.SetSize(m.width, m.height)
		return m, m.prdetail.Init()

	case dashboard.CreatePRMsg:
		// Navigate to create PR wizard
		m.screen = ScreenCreatePR
		m.createpr = createpr.New(m.cfg, m.ghClient, m.gitRepo, m.llmProvider)
		m.createpr.SetSize(m.width, m.height)
		return m, m.createpr.Init()

	// ==========================================================================
	// Navigation messages from PR Detail
	// ==========================================================================

	case prdetail.BackToDashboardMsg:
		// Cleanup is already called by prdetail before sending this message,
		// but we call it here as a safety net in case of direct navigation
		m.prdetail.Cleanup()
		m.screen = ScreenDashboard
		// Dashboard maintains its state; no need to reinitialize
		return m, nil

	case prdetail.ViewDiffMsg:
		// Cleanup is already called by prdetail before sending this message,
		// but we call it here as a safety net in case of direct navigation
		m.prdetail.Cleanup()
		m.screen = ScreenDiffView
		m.diffview = diffview.New(m.cfg, msg.PRNumber)
		m.diffview.SetSize(m.width, m.height)
		return m, m.diffview.Init()

	case prdetail.MergeDoneMsg:
		// Show success toast when PR is merged
		return m, showToast("PR merged successfully", ui.ToastSuccess)

	case prdetail.MergeErrorMsg:
		// Show error toast when merge fails
		return m, showToast(fmt.Sprintf("Merge failed: %v", msg.Err), ui.ToastError)

	// ==========================================================================
	// Navigation messages from Diff View
	// ==========================================================================

	case diffview.BackToPRDetailMsg:
		m.screen = ScreenPRDetail
		// PR detail should still have the PR loaded, so just switch back
		// If we wanted to ensure the PR is reloaded, we would reinitialize here
		return m, nil

	// ==========================================================================
	// Navigation messages from Create PR
	// ==========================================================================

	case createpr.BackToDashboardMsg:
		m.screen = ScreenDashboard
		// Dashboard maintains its state; no need to reinitialize
		return m, nil

	case createpr.PRCreatedMsg:
		// Navigate to PR detail for the newly created PR
		m.screen = ScreenPRDetail
		m.prdetail = prdetail.New(m.cfg, msg.PR.Number)
		m.prdetail.SetSize(m.width, m.height)
		// Show success toast and initialize PR detail
		return m, tea.Batch(
			m.prdetail.Init(),
			showToast(fmt.Sprintf("PR #%d created", msg.PR.Number), ui.ToastSuccess),
		)

	case createpr.PRCreateErrorMsg:
		// Show error toast when PR creation fails
		return m, showToast(fmt.Sprintf("Failed to create PR: %v", msg.Err), ui.ToastError)

	// ==========================================================================
	// Navigation messages from PR Detail to Review
	// ==========================================================================

	case prdetail.StartReviewMsg:
		// Cleanup is already called by prdetail before sending this message,
		// but we call it here as a safety net in case of direct navigation
		m.prdetail.Cleanup()
		// Navigate to review screen
		m.screen = ScreenReview
		m.review = review.New(m.cfg, m.ghClient, m.llmProvider, msg.PRNumber, msg.PRTitle, msg.Diff)
		m.review.SetSize(m.width, m.height)
		return m, m.review.Init()

	// ==========================================================================
	// Navigation messages from Review
	// ==========================================================================

	case review.BackToPRDetailMsg:
		m.screen = ScreenPRDetail
		// PR detail should still have the PR loaded, so just switch back
		return m, nil

	case review.ReviewSubmittedMsg:
		// Navigate back to PR detail and refresh
		m.screen = ScreenPRDetail
		// Reinitialize to refresh the PR data after review submission
		m.prdetail = prdetail.New(m.cfg, m.prdetail.PRNumber())
		m.prdetail.SetSize(m.width, m.height)
		// Show success toast and initialize PR detail
		return m, tea.Batch(
			m.prdetail.Init(),
			showToast("Review submitted", ui.ToastSuccess),
		)
	}

	// Route update to active screen
	switch m.screen {
	case ScreenDashboard:
		var newModel tea.Model
		newModel, cmd = m.dashboard.Update(msg)
		m.dashboard = newModel.(dashboard.Model)
		cmds = append(cmds, cmd)

	case ScreenPRDetail:
		var newModel tea.Model
		newModel, cmd = m.prdetail.Update(msg)
		m.prdetail = newModel.(prdetail.Model)
		cmds = append(cmds, cmd)

	case ScreenDiffView:
		var newModel tea.Model
		newModel, cmd = m.diffview.Update(msg)
		m.diffview = newModel.(diffview.Model)
		cmds = append(cmds, cmd)

	case ScreenCreatePR:
		var newModel tea.Model
		newModel, cmd = m.createpr.Update(msg)
		m.createpr = newModel.(createpr.Model)
		cmds = append(cmds, cmd)

	case ScreenReview:
		var newModel tea.Model
		newModel, cmd = m.review.Update(msg)
		m.review = newModel.(review.Model)
		cmds = append(cmds, cmd)

	case ScreenHelp:
		var newModel tea.Model
		newModel, cmd = m.help.Update(msg)
		m.help = newModel.(help.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model. It renders the active screen.
func (m Model) View() string {
	var view string

	// Render active screen
	switch m.screen {
	case ScreenDashboard:
		view = m.dashboard.View()
	case ScreenPRDetail:
		view = m.prdetail.View()
	case ScreenDiffView:
		view = m.diffview.View()
	case ScreenCreatePR:
		view = m.createpr.View()
	case ScreenReview:
		view = m.review.View()
	case ScreenHelp:
		view = m.help.View()
	default:
		view = "Unknown screen"
	}

	// Overlay toast if present
	if m.toast != nil && !m.toast.IsEmpty() {
		toastView := m.toast.SetWidth(m.width - 4).Render()
		// Position toast at the bottom center of the screen
		view = lipgloss.JoinVertical(lipgloss.Center, view, toastView)
	}

	return view
}

// =============================================================================
// Accessors
// =============================================================================

// CurrentScreen returns the active screen.
func (m Model) CurrentScreen() Screen {
	return m.screen
}

// SetScreen changes the active screen. This is primarily useful for testing.
func (m *Model) SetScreen(s Screen) {
	m.screen = s
}

// Width returns the current terminal width.
func (m Model) Width() int {
	return m.width
}

// Height returns the current terminal height.
func (m Model) Height() int {
	return m.height
}

// Dashboard returns the dashboard model. Useful for testing.
func (m Model) Dashboard() dashboard.Model {
	return m.dashboard
}

// PRDetail returns the PR detail model. Useful for testing.
func (m Model) PRDetail() prdetail.Model {
	return m.prdetail
}

// DiffView returns the diff view model. Useful for testing.
func (m Model) DiffView() diffview.Model {
	return m.diffview
}

// CreatePR returns the create PR wizard model. Useful for testing.
func (m Model) CreatePR() createpr.Model {
	return m.createpr
}

// Review returns the review model. Useful for testing.
func (m Model) Review() review.Model {
	return m.review
}

// Help returns the help model. Useful for testing.
func (m Model) Help() help.Model {
	return m.help
}

// PreviousScreen returns the screen that was active before opening help.
// Useful for testing.
func (m Model) PreviousScreen() Screen {
	return m.previousScreen
}

// Toast returns the current toast. Useful for testing.
func (m Model) Toast() *ui.Toast {
	return m.toast
}

// ToastTimer returns the current toast countdown timer. Useful for testing.
func (m Model) ToastTimer() int {
	return m.toastTimer
}

// =============================================================================
// Internal Helpers
// =============================================================================

// updateScreenSizes updates the dimensions of all screens.
// Only the active screen needs to be sized, but we do all for consistency
// when switching screens.
func (m *Model) updateScreenSizes() {
	switch m.screen {
	case ScreenDashboard:
		m.dashboard.SetSize(m.width, m.height)
	case ScreenPRDetail:
		m.prdetail.SetSize(m.width, m.height)
	case ScreenDiffView:
		m.diffview.SetSize(m.width, m.height)
	case ScreenCreatePR:
		m.createpr.SetSize(m.width, m.height)
	case ScreenReview:
		m.review.SetSize(m.width, m.height)
	case ScreenHelp:
		m.help.SetSize(m.width, m.height)
	}
}

// screenToHelpContext maps a screen to its corresponding help context.
func screenToHelpContext(s Screen) help.HelpContext {
	switch s {
	case ScreenDashboard:
		return help.HelpContextDashboard
	case ScreenPRDetail:
		return help.HelpContextPRDetail
	case ScreenDiffView:
		return help.HelpContextDiffView
	case ScreenCreatePR:
		return help.HelpContextCreatePR
	case ScreenReview:
		return help.HelpContextReview
	default:
		return help.HelpContextGlobal
	}
}

// showToast creates a command that displays a toast notification.
func showToast(msg string, toastType ui.ToastType) tea.Cmd {
	return func() tea.Msg {
		return ShowToastMsg{Message: msg, Type: toastType}
	}
}

// toastTick creates a command that sends a tick for toast countdown.
func toastTick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return ToastTickMsg{}
	})
}
