// Package gh provides a wrapper around the GitHub CLI (gh) for managing pull requests,
// reviews, and CI checks in a terminal-native application.
package gh

import "time"

// PR represents a GitHub Pull Request with all relevant metadata.
type PR struct {
	Number     int        `json:"number"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	State      string     `json:"state"` // "open", "closed", "merged"
	Author     User       `json:"author"`
	HeadBranch string     `json:"headRefName"`
	BaseBranch string     `json:"baseRefName"`
	URL        string     `json:"url"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	IsDraft    bool       `json:"isDraft"`
	Mergeable  string     `json:"mergeable"` // "MERGEABLE", "CONFLICTING", "UNKNOWN"
	Additions  int        `json:"additions"`
	Deletions  int        `json:"deletions"`
	Labels     []Label    `json:"labels"`
	Reviewers  []Reviewer `json:"reviewRequests"`
}

// User represents a GitHub user.
type User struct {
	Login string `json:"login"`
}

// Label represents a GitHub label on a PR.
type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Reviewer represents a reviewer on a PR with their review state.
type Reviewer struct {
	Login string `json:"login"`
	State string `json:"state"` // "APPROVED", "CHANGES_REQUESTED", "COMMENTED", "PENDING"
}

// Review represents a single review on a PR.
type Review struct {
	ID        int       `json:"id"`
	Author    User      `json:"author"`
	State     string    `json:"state"` // "APPROVED", "CHANGES_REQUESTED", "COMMENTED", "PENDING"
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"submittedAt"`
}

// Check represents a CI/status check on a PR.
type Check struct {
	Name       string `json:"name"`
	Status     string `json:"state"`      // "queued", "in_progress", "completed" (gh uses "state")
	Conclusion string `json:"conclusion"` // "success", "failure", "neutral", "cancelled", "skipped", "timed_out"
	URL        string `json:"link"`
}

// PRFilters defines filtering options for listing PRs.
type PRFilters struct {
	State    string   // "open", "closed", "merged", "all"
	Author   string   // Filter by author username
	Assignee string   // Filter by assignee username
	Base     string   // Filter by base branch
	Head     string   // Filter by head branch
	Labels   []string // Filter by labels
	Limit    int      // Maximum number of PRs to return
}

// MergeMethod specifies the strategy for merging a PR.
type MergeMethod string

const (
	// MergeMethodMerge creates a merge commit.
	MergeMethodMerge MergeMethod = "merge"
	// MergeMethodSquash squashes all commits into one.
	MergeMethodSquash MergeMethod = "squash"
	// MergeMethodRebase rebases commits onto the base branch.
	MergeMethodRebase MergeMethod = "rebase"
)

// ReviewEvent specifies the type of review action.
type ReviewEvent string

const (
	// ReviewApprove approves the PR.
	ReviewApprove ReviewEvent = "APPROVE"
	// ReviewRequestChanges requests changes on the PR.
	ReviewRequestChanges ReviewEvent = "REQUEST_CHANGES"
	// ReviewComment leaves a comment without approval or rejection.
	ReviewComment ReviewEvent = "COMMENT"
)

// String returns the lowercase flag name for gh CLI.
func (m MergeMethod) String() string {
	return string(m)
}

// Flag returns the CLI flag for the merge method.
func (m MergeMethod) Flag() string {
	return "--" + string(m)
}

// Flag returns the CLI flag for the review event.
func (e ReviewEvent) Flag() string {
	switch e {
	case ReviewApprove:
		return "--approve"
	case ReviewRequestChanges:
		return "--request-changes"
	case ReviewComment:
		return "--comment"
	default:
		return "--comment"
	}
}
