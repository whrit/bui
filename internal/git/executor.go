// Package git provides utilities for interacting with Git repositories.
// It wraps git commands to provide repository information, commit logs,
// diffs, and branch operations needed for PR management.
package git

import (
	"fmt"
	"os/exec"
)

// allowedCommands defines the set of commands that the executor is permitted to run.
// This is a security measure to prevent arbitrary command execution.
var allowedCommands = map[string]bool{
	"git": true,
}

// CommandExecutor defines the interface for executing shell commands.
// This abstraction enables dependency injection for testing.
type CommandExecutor interface {
	// Run executes the named command with the given arguments and returns the output.
	Run(name string, args ...string) ([]byte, error)

	// RunInDir executes the named command in the specified directory.
	RunInDir(dir, name string, args ...string) ([]byte, error)
}

// realExecutor implements CommandExecutor using os/exec.
type realExecutor struct{}

// NewExecutor creates a new CommandExecutor that executes real commands.
func NewExecutor() CommandExecutor {
	return &realExecutor{}
}

// Run executes the command and returns combined stdout/stderr output.
func (e *realExecutor) Run(name string, args ...string) ([]byte, error) {
	if !allowedCommands[name] {
		return nil, fmt.Errorf("%w: %q", ErrCommandNotAllowed, name)
	}

	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("command %q failed: %w\nOutput: %s", name, exitErr, string(output))
		}
		return nil, fmt.Errorf("command %q failed: %w", name, err)
	}
	return output, nil
}

// RunInDir executes the command in the specified directory.
func (e *realExecutor) RunInDir(dir, name string, args ...string) ([]byte, error) {
	if !allowedCommands[name] {
		return nil, fmt.Errorf("%w: %q", ErrCommandNotAllowed, name)
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("command %q in dir %q failed: %w\nOutput: %s", name, dir, exitErr, string(output))
		}
		return nil, fmt.Errorf("command %q in dir %q failed: %w", name, dir, err)
	}
	return output, nil
}
