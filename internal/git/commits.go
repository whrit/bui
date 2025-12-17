package git

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Commit represents a Git commit.
type Commit struct {
	Hash        string    // Full commit hash
	ShortHash   string    // Short (7 char) commit hash
	Author      string    // Author name
	AuthorEmail string    // Author email
	Date        time.Time // Commit date
	Subject     string    // First line of commit message
	Body        string    // Rest of commit message (may be empty)
}

// commitLogFormat is the custom format string for git log output.
// Uses ASCII delimiters (0x1E record separator, 0x1F unit separator) to avoid
// conflicts with commit message content.
const commitLogFormat = "%H%x1f%h%x1f%an%x1f%ae%x1f%at%x1f%s%x1f%b%x1e"

// fieldSeparator is the unit separator character (0x1F).
const fieldSeparator = "\x1f"

// recordSeparator is the record separator character (0x1E).
const recordSeparator = "\x1e"

// GetCurrentBranch returns the name of the current branch.
// Returns "HEAD" if in detached HEAD state.
func (r *Repo) GetCurrentBranch() (string, error) {
	output, err := r.executor.Run(gitCommand, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(string(output))
	return branch, nil
}

// GetCommitLog returns commits between base and head references.
// The commits are returned in reverse chronological order (newest first).
// If limit is 0 or negative, all commits are returned.
func (r *Repo) GetCommitLog(base, head string, limit int) ([]Commit, error) {
	args := []string{
		"log",
		"--format=" + commitLogFormat,
	}

	if limit > 0 {
		args = append(args, "-n", strconv.Itoa(limit))
	}

	// Add revision range
	args = append(args, base+".."+head)

	output, err := r.executor.Run(gitCommand, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}

	return parseCommitLog(string(output))
}

// parseCommitLog parses the output of git log with our custom format.
func parseCommitLog(output string) ([]Commit, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return []Commit{}, nil
	}

	// Split by record separator
	records := strings.Split(output, recordSeparator)
	commits := make([]Commit, 0, len(records))

	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}

		commit, err := parseCommitRecord(record)
		if err != nil {
			return nil, fmt.Errorf("failed to parse commit record: %w", err)
		}
		commits = append(commits, commit)
	}

	return commits, nil
}

// parseCommitRecord parses a single commit record.
func parseCommitRecord(record string) (Commit, error) {
	fields := strings.Split(record, fieldSeparator)
	if len(fields) < 6 {
		return Commit{}, fmt.Errorf("invalid commit record: expected at least 6 fields, got %d", len(fields))
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(fields[4], 10, 64)
	if err != nil {
		return Commit{}, fmt.Errorf("invalid timestamp %q: %w", fields[4], err)
	}

	// Body may be empty or contain the 7th field
	body := ""
	if len(fields) > 6 {
		body = strings.TrimSpace(fields[6])
	}

	return Commit{
		Hash:        fields[0],
		ShortHash:   fields[1],
		Author:      fields[2],
		AuthorEmail: fields[3],
		Date:        time.Unix(timestamp, 0),
		Subject:     fields[5],
		Body:        body,
	}, nil
}

// GetCommitMessages returns just the commit subjects (first lines) between base and head.
// This is useful for generating PR descriptions.
func (r *Repo) GetCommitMessages(base, head string) ([]string, error) {
	args := []string{
		"log",
		"--format=%s",
		base + ".." + head,
	}

	output, err := r.executor.Run(gitCommand, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit messages: %w", err)
	}

	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return []string{}, nil
	}

	return strings.Split(outputStr, "\n"), nil
}

// GetCommit returns details for a single commit reference.
func (r *Repo) GetCommit(ref string) (*Commit, error) {
	args := []string{
		"log",
		"-1",
		"--format=" + commitLogFormat,
		ref,
	}

	output, err := r.executor.Run(gitCommand, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit %q: %w", ref, err)
	}

	commits, err := parseCommitLog(string(output))
	if err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("commit %q not found", ref)
	}

	return &commits[0], nil
}

// GetCommitCount returns the number of commits between base and head.
func (r *Repo) GetCommitCount(base, head string) (int, error) {
	args := []string{
		"rev-list",
		"--count",
		base + ".." + head,
	}

	output, err := r.executor.Run(gitCommand, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count commits: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("invalid commit count: %w", err)
	}

	return count, nil
}

// GetMergeBase returns the best common ancestor between two commits.
func (r *Repo) GetMergeBase(ref1, ref2 string) (string, error) {
	output, err := r.executor.Run(gitCommand, "merge-base", ref1, ref2)
	if err != nil {
		return "", fmt.Errorf("failed to find merge base between %q and %q: %w", ref1, ref2, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetHeadCommit returns the HEAD commit hash.
func (r *Repo) GetHeadCommit() (string, error) {
	output, err := r.executor.Run(gitCommand, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// ResolveRef resolves a reference (branch name, tag, etc.) to a commit hash.
func (r *Repo) ResolveRef(ref string) (string, error) {
	output, err := r.executor.Run(gitCommand, "rev-parse", ref)
	if err != nil {
		return "", fmt.Errorf("failed to resolve ref %q: %w", ref, err)
	}

	return strings.TrimSpace(string(output)), nil
}
