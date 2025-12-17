package gh

import (
	"errors"
	"fmt"
)

// Sentinel errors for common conditions
var (
	// ErrPRNotFound indicates the requested PR doesn't exist
	ErrPRNotFound = errors.New("pull request not found")

	// ErrInvalidPRNumber indicates an invalid PR number was provided
	ErrInvalidPRNumber = errors.New("invalid PR number")

	// ErrNotMergeable indicates the PR cannot be merged
	ErrNotMergeable = errors.New("pull request is not mergeable")

	// ErrAuthRequired indicates GitHub authentication is needed
	ErrAuthRequired = errors.New("GitHub authentication required")

	// ErrRateLimited indicates API rate limit exceeded
	ErrRateLimited = errors.New("GitHub API rate limit exceeded")

	// ErrCommandNotAllowed indicates a disallowed command was attempted
	ErrCommandNotAllowed = errors.New("command not allowed")
)

// PRError wraps errors with PR context
type PRError struct {
	Number int
	Op     string // Operation that failed (get, merge, close, etc.)
	Err    error
}

func (e *PRError) Error() string {
	return fmt.Sprintf("PR #%d %s: %v", e.Number, e.Op, e.Err)
}

func (e *PRError) Unwrap() error {
	return e.Err
}

// NewPRError creates a new PRError
func NewPRError(number int, op string, err error) *PRError {
	return &PRError{Number: number, Op: op, Err: err}
}

// validatePRNumber validates that the PR number is positive
func validatePRNumber(n int) error {
	if n <= 0 {
		return fmt.Errorf("%w: got %d", ErrInvalidPRNumber, n)
	}
	return nil
}
