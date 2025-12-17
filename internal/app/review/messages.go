package review

import tea "github.com/charmbracelet/bubbletea"

// =============================================================================
// Review Type Selection Messages
// =============================================================================

// ReviewTypeSelectedMsg is sent when a review type is selected.
type ReviewTypeSelectedMsg struct {
	Type ReviewType
}

// =============================================================================
// LLM Generation Messages
// =============================================================================

// GenerateCommentStartMsg signals that LLM generation should start.
type GenerateCommentStartMsg struct{}

// GenerateCommentTokenMsg contains a single token from the LLM stream.
type GenerateCommentTokenMsg struct {
	Token string
}

// GenerateCommentDoneMsg signals that LLM generation is complete.
type GenerateCommentDoneMsg struct{}

// GenerateCommentErrorMsg is sent when LLM generation fails.
type GenerateCommentErrorMsg struct {
	Err error
}

// =============================================================================
// Comment Messages
// =============================================================================

// CommentChangedMsg is sent when the comment content changes.
type CommentChangedMsg struct {
	Comment string
}

// =============================================================================
// Submission Messages
// =============================================================================

// SubmitReviewMsg signals that the review should be submitted.
type SubmitReviewMsg struct{}

// ReviewSubmittedMsg signals that the review was successfully submitted.
type ReviewSubmittedMsg struct{}

// ReviewSubmitErrorMsg is sent when review submission fails.
type ReviewSubmitErrorMsg struct {
	Err error
}

// =============================================================================
// Navigation Messages
// =============================================================================

// BackToPRDetailMsg signals to navigate back to PR detail screen.
type BackToPRDetailMsg struct {
	PRNumber int
}

// BackMsg signals to navigate back (generic).
type BackMsg struct{}

// =============================================================================
// Command Factories
// =============================================================================

// backToPRDetail creates a command that signals to go back to PR detail.
func backToPRDetail(prNumber int) tea.Cmd {
	return func() tea.Msg {
		return BackToPRDetailMsg{PRNumber: prNumber}
	}
}
