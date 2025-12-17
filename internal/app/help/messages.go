// Package help implements the help screen for the bui application.
// It displays keybindings and usage information in a scrollable viewport,
// providing context-sensitive help based on the current screen.
package help

import tea "github.com/charmbracelet/bubbletea"

// =============================================================================
// Navigation Messages
// =============================================================================

// CloseHelpMsg signals that the help screen should be closed
// and the application should return to the previous screen.
type CloseHelpMsg struct{}

// BackMsg is an alias for CloseHelpMsg for consistency with other screens.
type BackMsg struct{}

// =============================================================================
// Commands
// =============================================================================

// closeHelp creates a command that signals to close the help screen.
func closeHelp() tea.Cmd {
	return func() tea.Msg {
		return CloseHelpMsg{}
	}
}
