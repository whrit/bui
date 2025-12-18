// Package prdetail implements the PR detail screen for the bui application.
// It displays detailed information about a single pull request including
// description, checks, reviews, and provides actions like merge, close, and LLM summary.
package prdetail

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"bui/internal/gh"

	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Data Loading Messages
// =============================================================================

// PRLoadedMsg is sent when the PR data has been successfully loaded.
type PRLoadedMsg struct {
	PR *gh.PR
}

// ChecksLoadedMsg is sent when CI checks have been loaded.
type ChecksLoadedMsg struct {
	Checks []gh.Check
}

// ReviewsLoadedMsg is sent when reviews have been loaded.
type ReviewsLoadedMsg struct {
	Reviews []gh.Review
}

// DiffLoadedMsg is sent when the PR diff has been loaded.
type DiffLoadedMsg struct {
	Diff string
}

// LoadErrorMsg is sent when loading any data fails.
type LoadErrorMsg struct {
	Err error
}

// =============================================================================
// LLM Messages
// =============================================================================

// SummaryStartMsg signals that LLM summary generation has started.
type SummaryStartMsg struct{}

// SummaryTokenMsg contains a single token from the LLM stream.
type SummaryTokenMsg struct {
	Token        string
	GenerationID int64 // Tracks which generation this token belongs to
}

// SummaryDoneMsg signals that LLM summary generation is complete.
type SummaryDoneMsg struct{}

// SummaryErrorMsg is sent when LLM summary generation fails.
type SummaryErrorMsg struct {
	Err error
}

// =============================================================================
// Action Messages
// =============================================================================

// MergeStartMsg signals that a merge operation has started.
type MergeStartMsg struct{}

// MergeDoneMsg signals that the merge operation completed successfully.
type MergeDoneMsg struct{}

// MergeErrorMsg is sent when the merge operation fails.
type MergeErrorMsg struct {
	Err error
}

// CloseStartMsg signals that a close operation has started.
type CloseStartMsg struct{}

// CloseDoneMsg signals that the close operation completed successfully.
type CloseDoneMsg struct{}

// CloseErrorMsg is sent when the close operation fails.
type CloseErrorMsg struct {
	Err error
}

// =============================================================================
// Navigation Messages
// =============================================================================

// BackToDashboardMsg signals to navigate back to the dashboard.
type BackToDashboardMsg struct{}

// ViewDiffMsg signals to navigate to the diff view.
type ViewDiffMsg struct {
	PRNumber int
}

// StartReviewMsg signals to navigate to the review screen.
type StartReviewMsg struct {
	PRNumber int
	PRTitle  string
	Diff     string
}

// OpenInBrowserMsg signals to open a URL in the browser.
type OpenInBrowserMsg struct {
	URL string
}

// =============================================================================
// Commands
// =============================================================================

// loadPR creates a command that loads PR data from GitHub.
func loadPR(client *gh.Client, number int) tea.Cmd {
	return func() tea.Msg {
		pr, err := client.GetPR(number)
		if err != nil {
			return LoadErrorMsg{Err: err}
		}
		return PRLoadedMsg{PR: pr}
	}
}

// loadChecks creates a command that loads CI checks for a PR.
func loadChecks(client *gh.Client, number int) tea.Cmd {
	return func() tea.Msg {
		checks, err := client.GetChecks(number)
		if err != nil {
			return LoadErrorMsg{Err: err}
		}
		return ChecksLoadedMsg{Checks: checks}
	}
}

// loadReviews creates a command that loads reviews for a PR.
func loadReviews(client *gh.Client, number int) tea.Cmd {
	return func() tea.Msg {
		reviews, err := client.GetReviews(number)
		if err != nil {
			return LoadErrorMsg{Err: err}
		}
		return ReviewsLoadedMsg{Reviews: reviews}
	}
}

// loadDiff creates a command that loads the diff for a PR.
func loadDiff(client *gh.Client, number int) tea.Cmd {
	return func() tea.Msg {
		diff, err := client.GetPRDiff(number)
		if err != nil {
			return LoadErrorMsg{Err: err}
		}
		return DiffLoadedMsg{Diff: diff}
	}
}

// Note: LLM streaming is handled directly in update.go via readNextToken()
// The streaming commands are created inline to maintain proper channel references.

// mergePR creates a command that merges a PR.
func mergePR(client *gh.Client, number int, method gh.MergeMethod) tea.Cmd {
	return func() tea.Msg {
		err := client.MergePR(number, method)
		if err != nil {
			return MergeErrorMsg{Err: err}
		}
		return MergeDoneMsg{}
	}
}

// closePR creates a command that closes a PR.
func closePR(client *gh.Client, number int) tea.Cmd {
	return func() tea.Msg {
		err := client.ClosePR(number)
		if err != nil {
			return CloseErrorMsg{Err: err}
		}
		return CloseDoneMsg{}
	}
}

// BrowserErrorMsg is sent when opening the browser fails.
type BrowserErrorMsg struct {
	Err error
}

// validateURL checks that the URL is safe to open in a browser.
// Only allows HTTP and HTTPS schemes to prevent security issues.
func validateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported URL scheme %q: only http and https are allowed", parsed.Scheme)
	}

	if parsed.Host == "" {
		return fmt.Errorf("invalid URL: missing host")
	}

	return nil
}

// openBrowser creates a command that opens a URL in the system browser.
// Only HTTP and HTTPS URLs are allowed for security.
func openBrowser(urlStr string) tea.Cmd {
	return func() tea.Msg {
		// Validate URL before opening
		if err := validateURL(urlStr); err != nil {
			return BrowserErrorMsg{Err: err}
		}

		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", urlStr)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", urlStr)
		default: // linux, bsd, etc.
			cmd = exec.Command("xdg-open", urlStr)
		}
		_ = cmd.Start()
		return nil
	}
}

// backToDashboard creates a command that signals to go back to the dashboard.
func backToDashboard() tea.Cmd {
	return func() tea.Msg {
		return BackToDashboardMsg{}
	}
}

// viewDiff creates a command that signals to view the PR diff.
func viewDiff(prNumber int) tea.Cmd {
	return func() tea.Msg {
		return ViewDiffMsg{PRNumber: prNumber}
	}
}

// startReview creates a command that signals to start a review.
func startReview(prNumber int, prTitle string, diff string) tea.Cmd {
	return func() tea.Msg {
		return StartReviewMsg{PRNumber: prNumber, PRTitle: prTitle, Diff: diff}
	}
}
