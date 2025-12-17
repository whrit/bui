package git

import (
	"fmt"
	"strings"
)

// Branch represents a Git branch.
type Branch struct {
	Name       string // Branch name (e.g., "main", "feature/login")
	IsRemote   bool   // True if this is a remote tracking branch
	IsCurrent  bool   // True if this is the currently checked out branch
	Upstream   string // Upstream branch name (e.g., "origin/main")
	LastCommit string // Short hash of the last commit on this branch
}

// branchFormat is the custom format for git branch output.
// %(HEAD) - "*" if current branch
// %(refname:short) - branch name
// %(upstream:short) - upstream branch name
// %(objectname:short) - short commit hash
const branchFormat = "%(HEAD)%x1f%(refname:short)%x1f%(upstream:short)%x1f%(objectname:short)"

// ListBranches returns all local and remote branches.
func (r *Repo) ListBranches() ([]Branch, error) {
	// Get all branches (local and remote) with custom format
	output, err := r.executor.Run(gitCommand, "branch", "-a", "--format="+branchFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	return parseBranches(string(output))
}

// ListLocalBranches returns only local branches.
func (r *Repo) ListLocalBranches() ([]Branch, error) {
	output, err := r.executor.Run(gitCommand, "branch", "--format="+branchFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to list local branches: %w", err)
	}

	return parseBranches(string(output))
}

// ListRemoteBranches returns only remote tracking branches.
func (r *Repo) ListRemoteBranches() ([]Branch, error) {
	output, err := r.executor.Run(gitCommand, "branch", "-r", "--format="+branchFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to list remote branches: %w", err)
	}

	branches, err := parseBranches(string(output))
	if err != nil {
		return nil, err
	}

	// Mark all as remote
	for i := range branches {
		branches[i].IsRemote = true
	}

	return branches, nil
}

// parseBranches parses the output of git branch with our custom format.
func parseBranches(output string) ([]Branch, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	branches := make([]Branch, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		branch, err := parseBranchLine(line)
		if err != nil {
			continue // Skip malformed lines
		}

		branches = append(branches, branch)
	}

	return branches, nil
}

// parseBranchLine parses a single branch line.
func parseBranchLine(line string) (Branch, error) {
	fields := strings.Split(line, "\x1f")
	if len(fields) < 4 {
		return Branch{}, fmt.Errorf("invalid branch line: expected 4 fields, got %d", len(fields))
	}

	branch := Branch{
		IsCurrent:  fields[0] == "*",
		Name:       fields[1],
		Upstream:   fields[2],
		LastCommit: fields[3],
	}

	// Check if it's a remote branch (starts with "remotes/")
	if strings.HasPrefix(branch.Name, "remotes/") {
		branch.IsRemote = true
		// Remove "remotes/" prefix for cleaner display
		branch.Name = strings.TrimPrefix(branch.Name, "remotes/")
	}

	return branch, nil
}

// GetBranch returns details for a specific branch.
func (r *Repo) GetBranch(name string) (*Branch, error) {
	// First check if the branch exists
	exists, err := r.BranchExists(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("branch %q not found", name)
	}

	// Get branch details
	output, err := r.executor.Run(gitCommand, "branch", "--list", "--format="+branchFormat, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch %q: %w", name, err)
	}

	branches, err := parseBranches(string(output))
	if err != nil {
		return nil, err
	}

	if len(branches) == 0 {
		return nil, fmt.Errorf("branch %q not found", name)
	}

	return &branches[0], nil
}

// BranchExists checks if a branch exists (local or remote).
func (r *Repo) BranchExists(name string) (bool, error) {
	// Use git show-ref to check if the branch exists
	_, err := r.executor.Run(gitCommand, "show-ref", "--verify", "--quiet", "refs/heads/"+name)
	if err == nil {
		return true, nil
	}

	// Also check remote branches
	_, err = r.executor.Run(gitCommand, "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+name)
	if err == nil {
		return true, nil
	}

	return false, nil
}

// GetUpstreamBranch returns the upstream tracking branch for a local branch.
func (r *Repo) GetUpstreamBranch(name string) (string, error) {
	output, err := r.executor.Run(gitCommand, "rev-parse", "--abbrev-ref", name+"@{upstream}")
	if err != nil {
		return "", fmt.Errorf("failed to get upstream for branch %q: %w", name, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetDefaultBranch returns the default branch name (usually "main" or "master").
// It first tries to get from git config, then falls back to checking origin/HEAD.
func (r *Repo) GetDefaultBranch() (string, error) {
	// Try to get the default branch from git config
	output, err := r.executor.Run(gitCommand, "config", "--get", "init.defaultBranch")
	if err == nil {
		branch := strings.TrimSpace(string(output))
		if branch != "" {
			return branch, nil
		}
	}

	// Try to get from origin/HEAD
	output, err = r.executor.Run(gitCommand, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		ref := strings.TrimSpace(string(output))
		// Extract branch name from "refs/remotes/origin/main"
		if strings.HasPrefix(ref, "refs/remotes/origin/") {
			return strings.TrimPrefix(ref, "refs/remotes/origin/"), nil
		}
	}

	// Fall back to common defaults
	// Check if "main" exists
	if exists, _ := r.BranchExists("main"); exists {
		return "main", nil
	}

	// Check if "master" exists
	if exists, _ := r.BranchExists("master"); exists {
		return "master", nil
	}

	return "", fmt.Errorf("could not determine default branch")
}

// GetTrackingBranches returns all local branches that track a remote branch.
func (r *Repo) GetTrackingBranches() ([]Branch, error) {
	branches, err := r.ListLocalBranches()
	if err != nil {
		return nil, err
	}

	// Filter to only branches with upstream
	tracking := make([]Branch, 0)
	for _, branch := range branches {
		if branch.Upstream != "" {
			tracking = append(tracking, branch)
		}
	}

	return tracking, nil
}

// SetUpstream sets the upstream tracking branch for a local branch.
func (r *Repo) SetUpstream(local, remote string) error {
	_, err := r.executor.Run(gitCommand, "branch", "--set-upstream-to="+remote, local)
	if err != nil {
		return fmt.Errorf("failed to set upstream %q for branch %q: %w", remote, local, err)
	}
	return nil
}

// CreateBranch creates a new branch from a starting point.
func (r *Repo) CreateBranch(name, startPoint string) error {
	args := []string{"branch", name}
	if startPoint != "" {
		args = append(args, startPoint)
	}

	_, err := r.executor.Run(gitCommand, args...)
	if err != nil {
		return fmt.Errorf("failed to create branch %q: %w", name, err)
	}
	return nil
}

// DeleteBranch deletes a local branch.
// If force is true, uses -D (force delete), otherwise uses -d (safe delete).
func (r *Repo) DeleteBranch(name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}

	_, err := r.executor.Run(gitCommand, "branch", flag, name)
	if err != nil {
		return fmt.Errorf("failed to delete branch %q: %w", name, err)
	}
	return nil
}

// RenameBranch renames a branch.
func (r *Repo) RenameBranch(oldName, newName string) error {
	_, err := r.executor.Run(gitCommand, "branch", "-m", oldName, newName)
	if err != nil {
		return fmt.Errorf("failed to rename branch %q to %q: %w", oldName, newName, err)
	}
	return nil
}

// CheckoutBranch switches to a different branch.
func (r *Repo) CheckoutBranch(name string) error {
	_, err := r.executor.Run(gitCommand, "checkout", name)
	if err != nil {
		return fmt.Errorf("failed to checkout branch %q: %w", name, err)
	}
	return nil
}

// GetBranchCommitsBehindAhead returns the number of commits behind and ahead
// of the upstream branch.
func (r *Repo) GetBranchCommitsBehindAhead(branch string) (behind, ahead int, err error) {
	upstream, err := r.GetUpstreamBranch(branch)
	if err != nil {
		return 0, 0, err
	}

	// Use rev-list to count commits
	output, err := r.executor.Run(gitCommand, "rev-list", "--left-right", "--count", upstream+"..."+branch)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get commit counts: %w", err)
	}

	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list output: %q", string(output))
	}

	_, _ = fmt.Sscanf(parts[0], "%d", &behind)
	_, _ = fmt.Sscanf(parts[1], "%d", &ahead)

	return behind, ahead, nil
}
