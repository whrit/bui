package composer

import tea "github.com/charmbracelet/bubbletea"

// =============================================================================
// LLM Generation Messages
// =============================================================================

// GenerateStartMsg signals that LLM generation should start.
type GenerateStartMsg struct{}

// GenerateTokenMsg contains a single token from the LLM stream.
type GenerateTokenMsg struct {
	Token string
}

// GenerateDoneMsg signals that LLM generation is complete.
type GenerateDoneMsg struct{}

// GenerateErrorMsg is sent when LLM generation fails.
type GenerateErrorMsg struct {
	Err error
}

// =============================================================================
// Content Messages
// =============================================================================

// ContentAcceptedMsg signals that the user accepted the generated content.
type ContentAcceptedMsg struct {
	Content string
}

// ContentCancelledMsg signals that the user cancelled the composer.
type ContentCancelledMsg struct{}

// =============================================================================
// Navigation Messages
// =============================================================================

// BackMsg signals to navigate back from the composer.
type BackMsg struct{}

// =============================================================================
// Commands
// =============================================================================

// acceptContent creates a command that signals content was accepted.
func acceptContent(content string) tea.Cmd {
	return func() tea.Msg {
		return ContentAcceptedMsg{Content: content}
	}
}

// cancelComposer creates a command that signals the composer was cancelled.
func cancelComposer() tea.Cmd {
	return func() tea.Msg {
		return ContentCancelledMsg{}
	}
}
