package dashboard

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model. It handles all incoming messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	// Window size change
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil

	// PRs loaded successfully
	case PRsLoadedMsg:
		return m.handlePRsLoaded(msg)

	// Error loading PRs
	case PRsErrorMsg:
		return m.handlePRsError(msg)

	// Manual refresh request
	case RefreshMsg:
		return m.handleRefresh()

	// Auto-refresh tick
	case AutoRefreshTickMsg:
		return m.handleAutoRefreshTick()

	// Spinner tick
	case spinner.TickMsg:
		if m.state == StateLoading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)

	// Keyboard input
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	}

	// Pass remaining messages to the list
	if m.state == StateReady || m.state == StateFiltering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// =============================================================================
// Message Handlers
// =============================================================================

// handlePRsLoaded processes successfully loaded PRs.
func (m Model) handlePRsLoaded(msg PRsLoadedMsg) (tea.Model, tea.Cmd) {
	m.state = StateReady
	m.err = nil
	m.prs = msg.PRs

	// Convert PRs to list items
	items := make([]list.Item, len(msg.PRs))
	for i, pr := range msg.PRs {
		items[i] = PRItem{PR: pr}
	}
	m.list.SetItems(items)

	// Start auto-refresh timer if enabled
	var cmd tea.Cmd
	if m.autoRefresh {
		cmd = tickAutoRefresh(autoRefreshInterval)
	}

	return m, cmd
}

// handlePRsError processes PR loading errors.
func (m Model) handlePRsError(msg PRsErrorMsg) (tea.Model, tea.Cmd) {
	m.state = StateError
	m.err = msg.Err
	return m, nil
}

// handleRefresh triggers a PR list refresh.
func (m Model) handleRefresh() (tea.Model, tea.Cmd) {
	m.state = StateLoading
	m.err = nil

	cmds := []tea.Cmd{
		loadPRs(m.ghClient, m.filters),
		m.spinner.Init(),
	}

	return m, tea.Batch(cmds...)
}

// handleAutoRefreshTick handles the auto-refresh timer tick.
func (m Model) handleAutoRefreshTick() (tea.Model, tea.Cmd) {
	// Only auto-refresh if we're in ready state and auto-refresh is enabled
	if !m.autoRefresh || m.state != StateReady {
		// Reschedule the tick even if we skip refresh
		if m.autoRefresh {
			return m, tickAutoRefresh(autoRefreshInterval)
		}
		return m, nil
	}

	// Refresh without changing state to loading (silent refresh)
	cmd := loadPRs(m.ghClient, m.filters)
	return m, tea.Batch(cmd, tickAutoRefresh(autoRefreshInterval))
}

// handleKeyMsg processes keyboard input.
func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle state-specific key bindings first
	switch m.state {
	case StateFiltering:
		return m.handleFilteringKeys(msg)
	case StateError:
		return m.handleErrorKeys(msg)
	case StateLoading:
		return m.handleLoadingKeys(msg)
	case StateReady:
		return m.handleReadyKeys(msg)
	}

	return m, nil
}

// handleFilteringKeys handles keys when in filtering mode.
func (m Model) handleFilteringKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Cancel):
		// Exit filtering mode
		m.state = StateReady
		m.list.ResetFilter()
		return m, nil

	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit
	}

	// Pass to list for filter handling
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// handleErrorKeys handles keys when in error state.
func (m Model) handleErrorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keymap.Refresh):
		return m.handleRefresh()
	}

	return m, nil
}

// handleLoadingKeys handles keys when loading.
func (m Model) handleLoadingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit
	}

	return m, nil
}

// handleReadyKeys handles keys when ready for interaction.
func (m Model) handleReadyKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch {
	// Quit
	case key.Matches(msg, m.keymap.Quit):
		return m, tea.Quit

	// Refresh
	case key.Matches(msg, m.keymap.Refresh):
		return m.handleRefresh()

	// Search/Filter
	case key.Matches(msg, m.keymap.Search):
		m.state = StateFiltering
		// The list handles filtering internally
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	// Select/Enter - open PR detail
	case key.Matches(msg, m.keymap.Select), msg.Type == tea.KeyEnter:
		if pr := m.SelectedPR(); pr != nil {
			return m, openPRDetail(*pr)
		}
		return m, nil

	// Create new PR
	case key.Matches(msg, m.keymap.CreatePR):
		return m, createPR()

	// Open in browser
	case key.Matches(msg, m.keymap.OpenInBrowser):
		if pr := m.SelectedPR(); pr != nil {
			// Return command to open URL (handled by parent)
			return m, openInBrowser(pr.URL)
		}
		return m, nil

	// View diff
	case key.Matches(msg, m.keymap.ViewDiff):
		if pr := m.SelectedPR(); pr != nil {
			return m, viewDiff(pr.Number)
		}
		return m, nil
	}

	// Pass navigation keys to the list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// =============================================================================
// Additional Commands
// =============================================================================

// OpenInBrowserMsg signals to open a URL in the browser.
type OpenInBrowserMsg struct {
	URL string
}

// openInBrowser returns a command that signals to open a URL.
func openInBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		return OpenInBrowserMsg{URL: url}
	}
}

// ViewDiffMsg signals to view a PR's diff.
type ViewDiffMsg struct {
	PRNumber int
}

// viewDiff returns a command that signals to view a PR's diff.
func viewDiff(prNumber int) tea.Cmd {
	return func() tea.Msg {
		return ViewDiffMsg{PRNumber: prNumber}
	}
}
