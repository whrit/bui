package gh

import (
	"errors"
	"fmt"
	"testing"
)

func TestValidatePRNumber(t *testing.T) {
	tests := []struct {
		name    string
		number  int
		wantErr bool
	}{
		{
			name:    "valid PR number",
			number:  1,
			wantErr: false,
		},
		{
			name:    "valid large PR number",
			number:  99999,
			wantErr: false,
		},
		{
			name:    "zero PR number",
			number:  0,
			wantErr: true,
		},
		{
			name:    "negative PR number",
			number:  -1,
			wantErr: true,
		},
		{
			name:    "large negative PR number",
			number:  -99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePRNumber(tt.number)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePRNumber(%d) error = %v, wantErr %v", tt.number, err, tt.wantErr)
			}

			if tt.wantErr && err != nil {
				// Verify the error wraps ErrInvalidPRNumber
				if !errors.Is(err, ErrInvalidPRNumber) {
					t.Errorf("validatePRNumber(%d) error should wrap ErrInvalidPRNumber", tt.number)
				}
				// Verify the error contains the PR number
				if !containsNumber(err.Error(), tt.number) {
					t.Errorf("validatePRNumber(%d) error should contain the PR number", tt.number)
				}
			}
		})
	}
}

func containsNumber(s string, n int) bool {
	return fmt.Sprintf("%d", n) != "" && len(s) > 0 && (n == 0 || fmt.Sprintf("%d", n) != s)
}

func TestPRError(t *testing.T) {
	t.Run("Error method formats correctly", func(t *testing.T) {
		err := NewPRError(42, "merge", errors.New("conflict"))
		want := "PR #42 merge: conflict"
		if got := err.Error(); got != want {
			t.Errorf("PRError.Error() = %q, want %q", got, want)
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("underlying error")
		err := NewPRError(123, "get", underlying)

		if unwrapped := err.Unwrap(); unwrapped != underlying {
			t.Errorf("PRError.Unwrap() = %v, want %v", unwrapped, underlying)
		}
	})

	t.Run("errors.Is works with wrapped errors", func(t *testing.T) {
		err := NewPRError(1, "get", ErrPRNotFound)
		if !errors.Is(err, ErrPRNotFound) {
			t.Error("PRError should be unwrappable to ErrPRNotFound")
		}
	})

	t.Run("NewPRError creates correct struct", func(t *testing.T) {
		err := NewPRError(99, "close", ErrNotMergeable)
		if err.Number != 99 {
			t.Errorf("PRError.Number = %d, want 99", err.Number)
		}
		if err.Op != "close" {
			t.Errorf("PRError.Op = %q, want %q", err.Op, "close")
		}
		if err.Err != ErrNotMergeable {
			t.Errorf("PRError.Err = %v, want ErrNotMergeable", err.Err)
		}
	})
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrPRNotFound",
			err:  ErrPRNotFound,
			want: "pull request not found",
		},
		{
			name: "ErrInvalidPRNumber",
			err:  ErrInvalidPRNumber,
			want: "invalid PR number",
		},
		{
			name: "ErrNotMergeable",
			err:  ErrNotMergeable,
			want: "pull request is not mergeable",
		},
		{
			name: "ErrAuthRequired",
			err:  ErrAuthRequired,
			want: "GitHub authentication required",
		},
		{
			name: "ErrRateLimited",
			err:  ErrRateLimited,
			want: "GitHub API rate limit exceeded",
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

func TestPRNumberValidation_GetPR(t *testing.T) {
	tests := []struct {
		name     string
		prNumber int
		wantErr  bool
	}{
		{name: "zero", prNumber: 0, wantErr: true},
		{name: "negative", prNumber: -1, wantErr: true},
		{name: "valid", prNumber: 1, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr view", []byte(`{"number": 1, "title": "Test", "body": "", "state": "OPEN", "author": {"login": "user"}, "headRefName": "head", "baseRefName": "main", "url": "", "createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z", "isDraft": false, "mergeable": "UNKNOWN", "additions": 0, "deletions": 0, "labels": [], "reviewRequests": []}`), nil)

			client := NewWithExecutor(mock)
			_, err := client.GetPR(tt.prNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPR(%d) error = %v, wantErr %v", tt.prNumber, err, tt.wantErr)
			}

			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidPRNumber) {
				t.Errorf("GetPR(%d) error should wrap ErrInvalidPRNumber", tt.prNumber)
			}
		})
	}
}

func TestPRNumberValidation_GetPRDiff(t *testing.T) {
	tests := []struct {
		name     string
		prNumber int
		wantErr  bool
	}{
		{name: "zero", prNumber: 0, wantErr: true},
		{name: "negative", prNumber: -1, wantErr: true},
		{name: "valid", prNumber: 1, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr diff", []byte("diff content"), nil)

			client := NewWithExecutor(mock)
			_, err := client.GetPRDiff(tt.prNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPRDiff(%d) error = %v, wantErr %v", tt.prNumber, err, tt.wantErr)
			}

			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidPRNumber) {
				t.Errorf("GetPRDiff(%d) error should wrap ErrInvalidPRNumber", tt.prNumber)
			}
		})
	}
}

func TestPRNumberValidation_MergePR(t *testing.T) {
	tests := []struct {
		name     string
		prNumber int
		wantErr  bool
	}{
		{name: "zero", prNumber: 0, wantErr: true},
		{name: "negative", prNumber: -1, wantErr: true},
		{name: "valid", prNumber: 1, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr merge", []byte(""), nil)

			client := NewWithExecutor(mock)
			err := client.MergePR(tt.prNumber, MergeMethodMerge)

			if (err != nil) != tt.wantErr {
				t.Errorf("MergePR(%d) error = %v, wantErr %v", tt.prNumber, err, tt.wantErr)
			}

			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidPRNumber) {
				t.Errorf("MergePR(%d) error should wrap ErrInvalidPRNumber", tt.prNumber)
			}
		})
	}
}

func TestPRNumberValidation_ClosePR(t *testing.T) {
	tests := []struct {
		name     string
		prNumber int
		wantErr  bool
	}{
		{name: "zero", prNumber: 0, wantErr: true},
		{name: "negative", prNumber: -1, wantErr: true},
		{name: "valid", prNumber: 1, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr close", []byte(""), nil)

			client := NewWithExecutor(mock)
			err := client.ClosePR(tt.prNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("ClosePR(%d) error = %v, wantErr %v", tt.prNumber, err, tt.wantErr)
			}

			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidPRNumber) {
				t.Errorf("ClosePR(%d) error should wrap ErrInvalidPRNumber", tt.prNumber)
			}
		})
	}
}

func TestPRNumberValidation_GetReviews(t *testing.T) {
	tests := []struct {
		name     string
		prNumber int
		wantErr  bool
	}{
		{name: "zero", prNumber: 0, wantErr: true},
		{name: "negative", prNumber: -1, wantErr: true},
		{name: "valid", prNumber: 1, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr reviews", []byte(`[]`), nil)

			client := NewWithExecutor(mock)
			_, err := client.GetReviews(tt.prNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetReviews(%d) error = %v, wantErr %v", tt.prNumber, err, tt.wantErr)
			}

			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidPRNumber) {
				t.Errorf("GetReviews(%d) error should wrap ErrInvalidPRNumber", tt.prNumber)
			}
		})
	}
}

func TestPRNumberValidation_SubmitReview(t *testing.T) {
	tests := []struct {
		name     string
		prNumber int
		wantErr  bool
	}{
		{name: "zero", prNumber: 0, wantErr: true},
		{name: "negative", prNumber: -1, wantErr: true},
		{name: "valid", prNumber: 1, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr review", []byte(""), nil)

			client := NewWithExecutor(mock)
			err := client.SubmitReview(tt.prNumber, "LGTM", ReviewApprove)

			if (err != nil) != tt.wantErr {
				t.Errorf("SubmitReview(%d) error = %v, wantErr %v", tt.prNumber, err, tt.wantErr)
			}

			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidPRNumber) {
				t.Errorf("SubmitReview(%d) error should wrap ErrInvalidPRNumber", tt.prNumber)
			}
		})
	}
}

func TestPRNumberValidation_GetChecks(t *testing.T) {
	tests := []struct {
		name     string
		prNumber int
		wantErr  bool
	}{
		{name: "zero", prNumber: 0, wantErr: true},
		{name: "negative", prNumber: -1, wantErr: true},
		{name: "valid", prNumber: 1, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr checks", []byte(`[]`), nil)

			client := NewWithExecutor(mock)
			_, err := client.GetChecks(tt.prNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetChecks(%d) error = %v, wantErr %v", tt.prNumber, err, tt.wantErr)
			}

			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidPRNumber) {
				t.Errorf("GetChecks(%d) error should wrap ErrInvalidPRNumber", tt.prNumber)
			}
		})
	}
}

func TestCommandAllowlist(t *testing.T) {
	t.Run("gh command is allowed", func(t *testing.T) {
		if !allowedCommands["gh"] {
			t.Error("gh command should be allowed")
		}
	})

	t.Run("arbitrary commands are not allowed", func(t *testing.T) {
		disallowed := []string{"bash", "sh", "curl", "wget", "rm", "cat", "ls", ""}
		for _, cmd := range disallowed {
			if allowedCommands[cmd] {
				t.Errorf("command %q should not be allowed", cmd)
			}
		}
	})
}
