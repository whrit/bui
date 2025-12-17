// Package dashboard implements the main PR listing screen for the bui application.
// It displays a navigable list of pull requests with search/filter capabilities.
package dashboard

import (
	"time"

	"bui/internal/gh"

	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Messages
// =============================================================================

// PRsLoadedMsg is sent when pull requests have been successfully loaded.
type PRsLoadedMsg struct {
	PRs []gh.PR
}

// PRsErrorMsg is sent when loading pull requests fails.
type PRsErrorMsg struct {
	Err error
}

// RefreshMsg triggers a PR list refresh.
type RefreshMsg struct{}

// OpenPRDetailMsg signals that the user wants to open the PR detail view.
type OpenPRDetailMsg struct {
	PR gh.PR
}

// CreatePRMsg signals that the user wants to create a new PR.
type CreatePRMsg struct{}

// AutoRefreshTickMsg is sent periodically when auto-refresh is enabled.
type AutoRefreshTickMsg struct {
	Time time.Time
}

// =============================================================================
// Commands
// =============================================================================

// loadPRs creates a command that loads PRs from GitHub.
func loadPRs(client *gh.Client, filters gh.PRFilters) tea.Cmd {
	return func() tea.Msg {
		prs, err := client.ListPRs(filters)
		if err != nil {
			return PRsErrorMsg{Err: err}
		}
		return PRsLoadedMsg{PRs: prs}
	}
}

// tickAutoRefresh returns a command that sends an AutoRefreshTickMsg after the given duration.
func tickAutoRefresh(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return AutoRefreshTickMsg{Time: t}
	})
}

// openPRDetail returns a command that signals to open the PR detail view.
func openPRDetail(pr gh.PR) tea.Cmd {
	return func() tea.Msg {
		return OpenPRDetailMsg{PR: pr}
	}
}

// createPR returns a command that signals to open the create PR wizard.
func createPR() tea.Cmd {
	return func() tea.Msg {
		return CreatePRMsg{}
	}
}
