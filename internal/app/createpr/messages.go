// Package createpr implements the Create PR wizard screen for the bui application.
// It provides a multi-step wizard for creating GitHub pull requests with optional
// LLM-assisted draft generation.
package createpr

import (
	"bui/internal/gh"
	"bui/internal/git"

	tea "github.com/charmbracelet/bubbletea"
)

// =============================================================================
// Branch Loading Messages
// =============================================================================

// BranchesLoadedMsg is sent when branches have been successfully loaded.
type BranchesLoadedMsg struct {
	Branches      []string
	CurrentBranch string
	DefaultBranch string
}

// BranchesLoadErrorMsg is sent when loading branches fails.
type BranchesLoadErrorMsg struct {
	Err error
}

// BranchSelectedMsg signals that a branch has been selected.
type BranchSelectedMsg struct {
	Branch string
	IsBase bool // true if base branch, false if head branch
}

// =============================================================================
// Commit Loading Messages
// =============================================================================

// CommitsLoadedMsg is sent when commits between branches have been loaded.
type CommitsLoadedMsg struct {
	Commits []git.Commit
}

// CommitsLoadErrorMsg is sent when loading commits fails.
type CommitsLoadErrorMsg struct {
	Err error
}

// =============================================================================
// LLM Draft Generation Messages
// =============================================================================

// DraftStartMsg signals that LLM draft generation should start.
type DraftStartMsg struct{}

// DraftTokenMsg contains a single token from the LLM stream.
type DraftTokenMsg struct {
	Token string
}

// DraftDoneMsg signals that LLM draft generation is complete.
type DraftDoneMsg struct{}

// DraftErrorMsg is sent when LLM draft generation fails.
type DraftErrorMsg struct {
	Err error
}

// SkipDraftMsg signals that the user wants to skip LLM draft generation.
type SkipDraftMsg struct{}

// =============================================================================
// Content Messages
// =============================================================================

// TitleChangedMsg signals that the PR title has been changed.
type TitleChangedMsg struct {
	Title string
}

// BodyChangedMsg signals that the PR body has been changed.
type BodyChangedMsg struct {
	Body string
}

// =============================================================================
// PR Creation Messages
// =============================================================================

// SubmitPRMsg signals that the user wants to submit the PR.
type SubmitPRMsg struct{}

// PRCreatedMsg is sent when a PR has been successfully created.
type PRCreatedMsg struct {
	PR *gh.PR
}

// PRCreateErrorMsg is sent when creating a PR fails.
type PRCreateErrorMsg struct {
	Err error
}

// =============================================================================
// Navigation Messages
// =============================================================================

// BackToDashboardMsg signals to navigate back to the dashboard.
type BackToDashboardMsg struct{}

// NextStepMsg signals to advance to the next wizard step.
type NextStepMsg struct{}

// PrevStepMsg signals to go back to the previous wizard step.
type PrevStepMsg struct{}

// =============================================================================
// Commands - Branch Loading
// =============================================================================

// loadBranches creates a command that loads branches from the git repository.
func loadBranches(repo *git.Repo) tea.Cmd {
	return func() tea.Msg {
		if repo == nil {
			return BranchesLoadErrorMsg{Err: &createprError{"git repository is not available"}}
		}

		// Get local branches
		branches, err := repo.ListLocalBranches()
		if err != nil {
			return BranchesLoadErrorMsg{Err: err}
		}

		// Extract branch names
		branchNames := make([]string, len(branches))
		for i, b := range branches {
			branchNames[i] = b.Name
		}

		// Get current branch
		currentBranch, err := repo.GetCurrentBranch()
		if err != nil {
			return BranchesLoadErrorMsg{Err: err}
		}

		// Get default branch
		defaultBranch, err := repo.GetDefaultBranch()
		if err != nil {
			// Fall back to "main" if we can't determine default
			defaultBranch = "main"
		}

		return BranchesLoadedMsg{
			Branches:      branchNames,
			CurrentBranch: currentBranch,
			DefaultBranch: defaultBranch,
		}
	}
}

// =============================================================================
// Commands - Commit Loading
// =============================================================================

// loadCommits creates a command that loads commits between base and head branches.
func loadCommits(repo *git.Repo, base, head string) tea.Cmd {
	return func() tea.Msg {
		if repo == nil {
			return CommitsLoadErrorMsg{Err: &createprError{"git repository is not available"}}
		}

		commits, err := repo.GetCommitLog(base, head, 50) // Limit to 50 commits
		if err != nil {
			return CommitsLoadErrorMsg{Err: err}
		}

		return CommitsLoadedMsg{Commits: commits}
	}
}

// =============================================================================
// Commands - PR Creation
// =============================================================================

// createPR creates a command that creates a new PR via the GitHub CLI.
func createPR(client *gh.Client, base, head, title, body string, isDraft bool) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return PRCreateErrorMsg{Err: &createprError{"GitHub client is not available"}}
		}

		pr, err := client.CreatePR(base, head, title, body, isDraft)
		if err != nil {
			return PRCreateErrorMsg{Err: err}
		}

		return PRCreatedMsg{PR: pr}
	}
}

// =============================================================================
// Commands - Navigation
// =============================================================================

// backToDashboard creates a command that signals to return to the dashboard.
func backToDashboard() tea.Cmd {
	return func() tea.Msg {
		return BackToDashboardMsg{}
	}
}

// =============================================================================
// Commands - LLM Draft
// =============================================================================

// startDraft creates a command that signals to start LLM draft generation.
func startDraft() tea.Cmd {
	return func() tea.Msg {
		return DraftStartMsg{}
	}
}

// skipDraft creates a command that signals to skip LLM draft generation.
func skipDraft() tea.Cmd {
	return func() tea.Msg {
		return SkipDraftMsg{}
	}
}

// submitPR creates a command that signals to submit the PR.
func submitPR() tea.Cmd {
	return func() tea.Msg {
		return SubmitPRMsg{}
	}
}

// =============================================================================
// Error Types
// =============================================================================

// createprError is a simple error type for the createpr package.
type createprError struct {
	msg string
}

func (e *createprError) Error() string {
	return e.msg
}
