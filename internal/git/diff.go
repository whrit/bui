package git

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DiffStat represents overall diff statistics.
type DiffStat struct {
	FilesChanged int // Number of files changed
	Additions    int // Total lines added
	Deletions    int // Total lines deleted
}

// FileDiff represents diff statistics for a single file.
type FileDiff struct {
	Path      string // Current file path
	OldPath   string // Previous path (for renames)
	Status    string // "A" (added), "M" (modified), "D" (deleted), "R" (renamed), "C" (copied)
	Additions int    // Lines added
	Deletions int    // Lines deleted
}

// diffStatRegex matches the summary line of git diff --stat output.
// Example: "3 files changed, 10 insertions(+), 5 deletions(-)"
var diffStatRegex = regexp.MustCompile(`(\d+) files? changed(?:, (\d+) insertions?\(\+\))?(?:, (\d+) deletions?\(-\))?`)

// GetDiff returns the raw diff between base and head references.
func (r *Repo) GetDiff(base, head string) (string, error) {
	output, err := r.executor.Run(gitCommand, "diff", base+".."+head)
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}

	return string(output), nil
}

// GetDiffStat returns overall statistics for changes between base and head.
func (r *Repo) GetDiffStat(base, head string) (*DiffStat, error) {
	// Use --stat to get summary statistics
	output, err := r.executor.Run(gitCommand, "diff", "--stat", base+".."+head)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff stat: %w", err)
	}

	return parseDiffStat(string(output))
}

// parseDiffStat parses the output of git diff --stat.
func parseDiffStat(output string) (*DiffStat, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || output == "" {
		return &DiffStat{}, nil
	}

	// The summary is on the last line
	summary := lines[len(lines)-1]

	matches := diffStatRegex.FindStringSubmatch(summary)
	if matches == nil {
		// No changes or unexpected format
		return &DiffStat{}, nil
	}

	stat := &DiffStat{}

	// Files changed
	if matches[1] != "" {
		stat.FilesChanged, _ = strconv.Atoi(matches[1])
	}

	// Insertions
	if matches[2] != "" {
		stat.Additions, _ = strconv.Atoi(matches[2])
	}

	// Deletions
	if matches[3] != "" {
		stat.Deletions, _ = strconv.Atoi(matches[3])
	}

	return stat, nil
}

// GetFileDiffs returns per-file diff statistics between base and head.
func (r *Repo) GetFileDiffs(base, head string) ([]FileDiff, error) {
	// Use --numstat for machine-readable per-file statistics
	// and --name-status for file status
	numstatOutput, err := r.executor.Run(gitCommand, "diff", "--numstat", base+".."+head)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff numstat: %w", err)
	}

	statusOutput, err := r.executor.Run(gitCommand, "diff", "--name-status", base+".."+head)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff name-status: %w", err)
	}

	return parseFileDiffs(string(numstatOutput), string(statusOutput))
}

// parseFileDiffs combines numstat and name-status output into FileDiff structs.
func parseFileDiffs(numstat, status string) ([]FileDiff, error) {
	// Parse numstat (additions, deletions, path)
	numstatMap := make(map[string]struct{ add, del int })
	for _, line := range strings.Split(strings.TrimSpace(numstat), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		add, _ := strconv.Atoi(parts[0])
		del, _ := strconv.Atoi(parts[1])
		// For renames, the path might include the old and new names
		path := parts[len(parts)-1]
		numstatMap[path] = struct{ add, del int }{add, del}
	}

	// Parse name-status and combine with numstat
	var diffs []FileDiff
	for _, line := range strings.Split(strings.TrimSpace(status), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		diff := FileDiff{
			Status: parts[0][:1], // First character is the status
		}

		// Handle renames (Rxxx format) and copies (Cxxx format)
		if strings.HasPrefix(parts[0], "R") || strings.HasPrefix(parts[0], "C") {
			if len(parts) >= 3 {
				diff.OldPath = parts[1]
				diff.Path = parts[2]
			}
		} else {
			diff.Path = parts[1]
		}

		// Get stats from numstat
		if stats, ok := numstatMap[diff.Path]; ok {
			diff.Additions = stats.add
			diff.Deletions = stats.del
		}

		diffs = append(diffs, diff)
	}

	return diffs, nil
}

// GetStagedDiff returns the diff of staged (indexed) changes.
func (r *Repo) GetStagedDiff() (string, error) {
	output, err := r.executor.Run(gitCommand, "diff", "--staged")
	if err != nil {
		return "", fmt.Errorf("failed to get staged diff: %w", err)
	}

	return string(output), nil
}

// GetUnstagedDiff returns the diff of unstaged working tree changes.
func (r *Repo) GetUnstagedDiff() (string, error) {
	output, err := r.executor.Run(gitCommand, "diff")
	if err != nil {
		return "", fmt.Errorf("failed to get unstaged diff: %w", err)
	}

	return string(output), nil
}

// GetStagedDiffStat returns statistics for staged changes.
func (r *Repo) GetStagedDiffStat() (*DiffStat, error) {
	output, err := r.executor.Run(gitCommand, "diff", "--staged", "--stat")
	if err != nil {
		return nil, fmt.Errorf("failed to get staged diff stat: %w", err)
	}

	return parseDiffStat(string(output))
}

// GetUnstagedDiffStat returns statistics for unstaged changes.
func (r *Repo) GetUnstagedDiffStat() (*DiffStat, error) {
	output, err := r.executor.Run(gitCommand, "diff", "--stat")
	if err != nil {
		return nil, fmt.Errorf("failed to get unstaged diff stat: %w", err)
	}

	return parseDiffStat(string(output))
}

// GetDiffForFile returns the diff for a specific file between base and head.
func (r *Repo) GetDiffForFile(base, head, filePath string) (string, error) {
	output, err := r.executor.Run(gitCommand, "diff", base+".."+head, "--", filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get diff for file %q: %w", filePath, err)
	}

	return string(output), nil
}

// GetChangedFiles returns a list of files changed between base and head.
func (r *Repo) GetChangedFiles(base, head string) ([]string, error) {
	output, err := r.executor.Run(gitCommand, "diff", "--name-only", base+".."+head)
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return []string{}, nil
	}

	return strings.Split(outputStr, "\n"), nil
}

// HasChanges checks if there are any changes between base and head.
func (r *Repo) HasChanges(base, head string) (bool, error) {
	// Use --quiet to suppress output and rely on exit code
	output, err := r.executor.Run(gitCommand, "diff", "--quiet", base+".."+head)
	if err != nil {
		// Exit code 1 means there are differences
		return true, nil
	}

	// Exit code 0 with empty output means no differences
	return len(output) > 0, nil
}

// HasStagedChanges checks if there are any staged changes.
func (r *Repo) HasStagedChanges() (bool, error) {
	output, err := r.executor.Run(gitCommand, "diff", "--staged", "--quiet")
	if err != nil {
		return true, nil
	}
	return len(output) > 0, nil
}

// HasUnstagedChanges checks if there are any unstaged changes.
func (r *Repo) HasUnstagedChanges() (bool, error) {
	output, err := r.executor.Run(gitCommand, "diff", "--quiet")
	if err != nil {
		return true, nil
	}
	return len(output) > 0, nil
}
