package gh

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ListPRs retrieves a list of pull requests based on the provided filters.
func (c *Client) ListPRs(filters PRFilters) ([]PR, error) {
	args := []string{"pr", "list"}
	args = append(args, filters.buildArgs()...)
	args = append(args, "--json", strings.Join(prFields, ","))

	output, err := c.executor.Run(ghCommand, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list PRs: %w", err)
	}

	var prs []PR
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, fmt.Errorf("failed to parse PR list response: %w", err)
	}

	return prs, nil
}

// GetPR retrieves a single pull request by number.
func (c *Client) GetPR(number int) (*PR, error) {
	args := []string{
		"pr", "view",
		strconv.Itoa(number),
		"--json", strings.Join(prFields, ","),
	}

	output, err := c.executor.Run(ghCommand, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR #%d: %w", number, err)
	}

	var pr PR
	if err := json.Unmarshal(output, &pr); err != nil {
		return nil, fmt.Errorf("failed to parse PR response: %w", err)
	}

	return &pr, nil
}

// GetPRDiff retrieves the diff for a pull request.
func (c *Client) GetPRDiff(number int) (string, error) {
	args := []string{
		"pr", "diff",
		strconv.Itoa(number),
	}

	output, err := c.executor.Run(ghCommand, args...)
	if err != nil {
		return "", fmt.Errorf("failed to get diff for PR #%d: %w", number, err)
	}

	return string(output), nil
}

// CreatePR creates a new pull request.
func (c *Client) CreatePR(base, head, title, body string, isDraft bool) (*PR, error) {
	args := []string{
		"pr", "create",
		"--base", base,
		"--head", head,
		"--title", title,
		"--body", body,
	}

	if isDraft {
		args = append(args, "--draft")
	}

	// Request JSON output to get the created PR details
	args = append(args, "--json", strings.Join(prFields, ","))

	output, err := c.executor.Run(ghCommand, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR: %w", err)
	}

	var pr PR
	if err := json.Unmarshal(output, &pr); err != nil {
		return nil, fmt.Errorf("failed to parse created PR response: %w", err)
	}

	return &pr, nil
}

// MergePR merges a pull request using the specified merge method.
func (c *Client) MergePR(number int, method MergeMethod) error {
	args := []string{
		"pr", "merge",
		strconv.Itoa(number),
		method.Flag(),
	}

	_, err := c.executor.Run(ghCommand, args...)
	if err != nil {
		return fmt.Errorf("failed to merge PR #%d: %w", number, err)
	}

	return nil
}

// ClosePR closes a pull request without merging.
func (c *Client) ClosePR(number int) error {
	args := []string{
		"pr", "close",
		strconv.Itoa(number),
	}

	_, err := c.executor.Run(ghCommand, args...)
	if err != nil {
		return fmt.Errorf("failed to close PR #%d: %w", number, err)
	}

	return nil
}

// buildArgs constructs command-line arguments from PRFilters.
func (f PRFilters) buildArgs() []string {
	var args []string

	if f.State != "" {
		args = append(args, "--state", f.State)
	}

	if f.Author != "" {
		args = append(args, "--author", f.Author)
	}

	if f.Assignee != "" {
		args = append(args, "--assignee", f.Assignee)
	}

	if f.Base != "" {
		args = append(args, "--base", f.Base)
	}

	if f.Head != "" {
		args = append(args, "--head", f.Head)
	}

	for _, label := range f.Labels {
		args = append(args, "--label", label)
	}

	if f.Limit > 0 {
		args = append(args, "--limit", strconv.Itoa(f.Limit))
	}

	return args
}
