package git

import "errors"

// Sentinel errors for common git conditions
var (
	// ErrNotARepository indicates the directory is not a git repository
	ErrNotARepository = errors.New("not a git repository")

	// ErrBranchNotFound indicates the requested branch doesn't exist
	ErrBranchNotFound = errors.New("branch not found")

	// ErrCommandNotAllowed indicates a disallowed command was attempted
	ErrCommandNotAllowed = errors.New("command not allowed")
)
