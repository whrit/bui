// Package root implements the root/shell model for the bui application.
// It manages screen navigation between Dashboard, PR Detail, Diff Viewer,
// Composer, Create PR, and Review screens.
package root

import (
	tea "github.com/charmbracelet/bubbletea"

	"bui/internal/app/createpr"
	"bui/internal/app/dashboard"
	"bui/internal/app/diffview"
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
	default:
		return "unknown"
	}
}

// =============================================================================
// Model
// =============================================================================

// Model is the root Bubble Tea model that orchestrates screen navigation.
type Model struct {
	// Configuration
	cfg     config.Config
	cfgPath string

	// Screen state
	screen Screen

	// Screen models
	dashboard dashboard.Model
	prdetail  prdetail.Model
	diffview  diffview.Model
	createpr  createpr.Model
	review    review.Model

	// Dependencies
	ghClient    *gh.Client
	gitRepo     *git.Repo
	llmProvider llm.Provider

	// Dimensions
	width, height int

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

	// Navigation messages from Dashboard
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

	// Navigation messages from PR Detail
	case prdetail.BackToDashboardMsg:
		m.screen = ScreenDashboard
		// Dashboard maintains its state; no need to reinitialize
		return m, nil

	case prdetail.ViewDiffMsg:
		m.screen = ScreenDiffView
		m.diffview = diffview.New(m.cfg, msg.PRNumber)
		m.diffview.SetSize(m.width, m.height)
		return m, m.diffview.Init()

	// Navigation messages from Diff View
	case diffview.BackToPRDetailMsg:
		m.screen = ScreenPRDetail
		// PR detail should still have the PR loaded, so just switch back
		// If we wanted to ensure the PR is reloaded, we would reinitialize here
		return m, nil

	// Navigation messages from Create PR
	case createpr.BackToDashboardMsg:
		m.screen = ScreenDashboard
		// Dashboard maintains its state; no need to reinitialize
		return m, nil

	case createpr.PRCreatedMsg:
		// Navigate to PR detail for the newly created PR
		m.screen = ScreenPRDetail
		m.prdetail = prdetail.New(m.cfg, msg.PR.Number)
		m.prdetail.SetSize(m.width, m.height)
		return m, m.prdetail.Init()

	// Navigation messages from PR Detail to Review
	case prdetail.StartReviewMsg:
		// Navigate to review screen
		m.screen = ScreenReview
		m.review = review.New(m.cfg, m.ghClient, m.llmProvider, msg.PRNumber, msg.PRTitle, msg.Diff)
		m.review.SetSize(m.width, m.height)
		return m, m.review.Init()

	// Navigation messages from Review
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
		return m, m.prdetail.Init()
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
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model. It renders the active screen.
func (m Model) View() string {
	switch m.screen {
	case ScreenDashboard:
		return m.dashboard.View()
	case ScreenPRDetail:
		return m.prdetail.View()
	case ScreenDiffView:
		return m.diffview.View()
	case ScreenCreatePR:
		return m.createpr.View()
	case ScreenReview:
		return m.review.View()
	default:
		return "Unknown screen"
	}
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
	}
}
