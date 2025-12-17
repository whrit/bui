package dashboard

import (
	"fmt"
	"time"

	"bui/internal/config"
	"bui/internal/gh"
	"bui/internal/ui"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Constants
// =============================================================================

const (
	// autoRefreshInterval is how often to refresh PRs when auto-refresh is enabled.
	autoRefreshInterval = 30 * time.Second

	// headerHeight is the height of the header section.
	headerHeight = 3

	// footerHeight is the height of the footer/help section.
	footerHeight = 2

	// defaultPRLimit is the default number of PRs to fetch.
	defaultPRLimit = 50
)

// =============================================================================
// State
// =============================================================================

// State represents the current state of the dashboard.
type State int

const (
	// StateLoading indicates PRs are being loaded.
	StateLoading State = iota
	// StateReady indicates PRs are loaded and ready for interaction.
	StateReady
	// StateError indicates an error occurred.
	StateError
	// StateFiltering indicates the user is filtering/searching.
	StateFiltering
)

// String returns a string representation of the state.
func (s State) String() string {
	switch s {
	case StateLoading:
		return "loading"
	case StateReady:
		return "ready"
	case StateError:
		return "error"
	case StateFiltering:
		return "filtering"
	default:
		return "unknown"
	}
}

// =============================================================================
// PRItem
// =============================================================================

// PRItem wraps gh.PR to implement the list.Item interface.
type PRItem struct {
	PR gh.PR
}

// Title returns the PR title for display in the list.
func (i PRItem) Title() string {
	return i.PR.Title
}

// Description returns a short description for display in the list.
func (i PRItem) Description() string {
	return fmt.Sprintf("#%d by %s", i.PR.Number, i.PR.Author.Login)
}

// FilterValue returns the value used for filtering/searching.
func (i PRItem) FilterValue() string {
	return i.PR.Title
}

// =============================================================================
// Model
// =============================================================================

// Model is the Bubble Tea model for the dashboard screen.
type Model struct {
	// Dependencies
	cfg      config.Config
	ghClient *gh.Client

	// UI state
	state   State
	list    list.Model
	search  textinput.Model
	spinner ui.Spinner
	err     error

	// Data
	prs     []gh.PR
	filters gh.PRFilters

	// Styling
	palette ui.Palette
	styles  ui.Styles
	keymap  ui.KeyMap

	// Dimensions
	width, height int

	// Auto-refresh
	autoRefresh bool
}

// New creates a new dashboard model.
func New(cfg config.Config) Model {
	palette := ui.NewPalette(cfg)
	styles := ui.NewStyles(palette)
	keymap := ui.NewKeyMap(cfg)

	// Initialize list with empty items (will be populated on PR load)
	delegate := NewDelegate(palette, styles)
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Pull Requests"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false) // We render our own help
	l.Styles.Title = styles.Header

	// Initialize search input
	ti := textinput.New()
	ti.Placeholder = "Search PRs..."
	ti.CharLimit = 100

	// Initialize spinner
	spinner := ui.NewSpinner(palette, styles).WithMessage("Loading pull requests...")

	// Default filters
	filters := gh.PRFilters{
		State: "open",
		Limit: defaultPRLimit,
	}

	return Model{
		cfg:         cfg,
		ghClient:    gh.New(),
		state:       StateLoading,
		list:        l,
		search:      ti,
		spinner:     spinner,
		filters:     filters,
		palette:     palette,
		styles:      styles,
		keymap:      keymap,
		autoRefresh: cfg.Behavior.AutoRefresh,
	}
}

// SelectedPR returns the currently selected PR, or nil if none is selected.
func (m Model) SelectedPR() *gh.PR {
	if len(m.list.Items()) == 0 {
		return nil
	}

	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}

	prItem, ok := item.(PRItem)
	if !ok {
		return nil
	}

	return &prItem.PR
}

// Init implements tea.Model. It returns the initial command to load PRs.
func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		loadPRs(m.ghClient, m.filters),
		m.spinner.Init(),
	}
	return tea.Batch(cmds...)
}

// SetSize updates the model dimensions and recalculates list size.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	// Calculate list dimensions accounting for header and footer
	listHeight := height - headerHeight - footerHeight
	if listHeight < 1 {
		listHeight = 1
	}

	m.list.SetSize(width, listHeight)
}

// =============================================================================
// Key Bindings
// =============================================================================

// DashboardKeyBindings returns the key bindings for the dashboard.
func DashboardKeyBindings() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("up/k", "up")),
		key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("down/j", "down")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "create")),
		key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}
