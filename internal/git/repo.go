package git

import (
	"fmt"
	"regexp"
	"strings"
)

// RepoInfo contains information about a Git repository.
type RepoInfo struct {
	Owner     string // Repository owner (user or organization)
	Name      string // Repository name
	RemoteURL string // Full remote URL
	RootPath  string // Absolute path to repository root
}

// Repo provides Git repository operations.
type Repo struct {
	executor CommandExecutor
}

// New creates a new Repo using the real command executor.
func New() *Repo {
	return &Repo{
		executor: NewExecutor(),
	}
}

// NewWithExecutor creates a new Repo with a custom executor for testing.
func NewWithExecutor(exec CommandExecutor) *Repo {
	return &Repo{
		executor: exec,
	}
}

// gitCommand is the base command for Git CLI.
const gitCommand = "git"

// Regular expressions for parsing GitHub remote URLs.
var (
	// Matches HTTPS URLs: https://github.com/owner/repo.git or https://github.com/owner/repo
	httpsRemoteRegex = regexp.MustCompile(`^https?://(?:www\.)?github\.com/([^/]+)/([^/\.]+?)(?:\.git)?$`)

	// Matches SSH URLs: git@github.com:owner/repo.git or git@github.com:owner/repo
	sshRemoteRegex = regexp.MustCompile(`^git@github\.com:([^/]+)/([^/\.]+?)(?:\.git)?$`)
)

// GetRepoInfo retrieves repository information from git remote.
// It parses the origin remote URL to extract owner and repository name.
func (r *Repo) GetRepoInfo() (*RepoInfo, error) {
	// Get the remote URL for origin
	output, err := r.executor.Run(gitCommand, "remote", "get-url", "origin")
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))
	if remoteURL == "" {
		return nil, fmt.Errorf("no remote URL configured for origin")
	}

	// Parse the remote URL
	owner, name, err := parseRemoteURL(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse remote URL %q: %w", remoteURL, err)
	}

	// Get the repository root path
	rootPath, err := r.GetRootPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get repository root: %w", err)
	}

	return &RepoInfo{
		Owner:     owner,
		Name:      name,
		RemoteURL: remoteURL,
		RootPath:  rootPath,
	}, nil
}

// parseRemoteURL extracts owner and repository name from a GitHub remote URL.
// Supports both HTTPS and SSH URL formats.
func parseRemoteURL(url string) (owner, name string, err error) {
	// Try HTTPS format first
	if matches := httpsRemoteRegex.FindStringSubmatch(url); len(matches) == 3 {
		return matches[1], matches[2], nil
	}

	// Try SSH format
	if matches := sshRemoteRegex.FindStringSubmatch(url); len(matches) == 3 {
		return matches[1], matches[2], nil
	}

	return "", "", fmt.Errorf("unsupported remote URL format")
}

// GetRootPath returns the absolute path to the repository root.
func (r *Repo) GetRootPath() (string, error) {
	output, err := r.executor.Run(gitCommand, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("failed to get repository root: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// IsInsideWorkTree checks if the current directory is inside a Git work tree.
func (r *Repo) IsInsideWorkTree() (bool, error) {
	output, err := r.executor.Run(gitCommand, "rev-parse", "--is-inside-work-tree")
	if err != nil {
		// If the command fails, we're not in a git repo
		return false, nil
	}

	return strings.TrimSpace(string(output)) == "true", nil
}

// GetRemoteURL returns the URL of a specific remote.
func (r *Repo) GetRemoteURL(remote string) (string, error) {
	output, err := r.executor.Run(gitCommand, "remote", "get-url", remote)
	if err != nil {
		return "", fmt.Errorf("failed to get URL for remote %q: %w", remote, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// ListRemotes returns a list of configured remotes.
func (r *Repo) ListRemotes() ([]string, error) {
	output, err := r.executor.Run(gitCommand, "remote")
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}

	lines := strings.TrimSpace(string(output))
	if lines == "" {
		return []string{}, nil
	}

	return strings.Split(lines, "\n"), nil
}
