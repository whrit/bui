package gh

import (
	"fmt"
	"os/exec"
)

// CommandExecutor defines the interface for executing shell commands.
// This abstraction enables dependency injection for testing.
type CommandExecutor interface {
	// Run executes the named command with the given arguments and returns the output.
	Run(name string, args ...string) ([]byte, error)
}

// Client wraps the GitHub CLI (gh) for PR operations.
type Client struct {
	executor CommandExecutor
}

// New creates a new Client using the real command executor.
func New() *Client {
	return &Client{
		executor: &realExecutor{},
	}
}

// NewWithExecutor creates a new Client with a custom executor for testing.
func NewWithExecutor(exec CommandExecutor) *Client {
	return &Client{
		executor: exec,
	}
}

// realExecutor implements CommandExecutor using os/exec.
type realExecutor struct{}

// Run executes the command and returns combined stdout/stderr output.
func (e *realExecutor) Run(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Include the output in the error for debugging
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("command %q failed: %w\nOutput: %s", name, exitErr, string(output))
		}
		return nil, fmt.Errorf("command %q failed: %w", name, err)
	}
	return output, nil
}

// ghCommand is the base command for GitHub CLI.
const ghCommand = "gh"

// prFields defines the JSON fields to request from gh pr list/view commands.
var prFields = []string{
	"number",
	"title",
	"body",
	"state",
	"author",
	"headRefName",
	"baseRefName",
	"url",
	"createdAt",
	"updatedAt",
	"isDraft",
	"mergeable",
	"additions",
	"deletions",
	"labels",
	"reviewRequests",
}

// reviewFields defines the JSON fields for gh pr reviews command.
var reviewFields = []string{
	"id",
	"author",
	"state",
	"body",
	"submittedAt",
}

// checkFields defines the JSON fields for gh pr checks command.
var checkFields = []string{
	"name",
	"state",
	"conclusion",
	"link",
}
