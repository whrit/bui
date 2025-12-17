package git

import (
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrNotARepository",
			err:  ErrNotARepository,
			want: "not a git repository",
		},
		{
			name: "ErrBranchNotFound",
			err:  ErrBranchNotFound,
			want: "branch not found",
		},
		{
			name: "ErrCommandNotAllowed",
			err:  ErrCommandNotAllowed,
			want: "command not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("%s.Error() = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestCommandAllowlist(t *testing.T) {
	t.Run("git command is allowed", func(t *testing.T) {
		if !allowedCommands["git"] {
			t.Error("git command should be allowed")
		}
	})

	t.Run("arbitrary commands are not allowed", func(t *testing.T) {
		disallowed := []string{"bash", "sh", "curl", "wget", "rm", "cat", "ls", "gh", ""}
		for _, cmd := range disallowed {
			if allowedCommands[cmd] {
				t.Errorf("command %q should not be allowed", cmd)
			}
		}
	})
}
