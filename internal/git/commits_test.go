package git

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestGetCurrentBranch(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		err        error
		wantBranch string
		wantErr    bool
	}{
		{
			name:       "regular branch",
			output:     "main\n",
			wantBranch: "main",
			wantErr:    false,
		},
		{
			name:       "feature branch",
			output:     "feature/add-login\n",
			wantBranch: "feature/add-login",
			wantErr:    false,
		},
		{
			name:       "detached HEAD",
			output:     "HEAD\n",
			wantBranch: "HEAD",
			wantErr:    false,
		},
		{
			name:    "not a git repo",
			err:     errors.New("fatal: not a git repository"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("rev-parse --abbrev-ref HEAD", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			branch, err := repo.GetCurrentBranch()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && branch != tt.wantBranch {
				t.Errorf("GetCurrentBranch() = %q, want %q", branch, tt.wantBranch)
			}
		})
	}
}

func TestGetCommitLog(t *testing.T) {
	// Example commit log output with our custom format
	// Format: %H%x1f%h%x1f%an%x1f%ae%x1f%at%x1f%s%x1f%b%x1e
	singleCommit := "abc123def456789012345678901234567890abcd\x1f" +
		"abc123d\x1f" +
		"John Doe\x1f" +
		"john@example.com\x1f" +
		"1704067200\x1f" +
		"Add new feature\x1f" +
		"This is the commit body.\x1e"

	multipleCommits := "abc123def456789012345678901234567890abcd\x1f" +
		"abc123d\x1f" +
		"John Doe\x1f" +
		"john@example.com\x1f" +
		"1704067200\x1f" +
		"Add new feature\x1f" +
		"Body for first commit\x1e" +
		"def456789012345678901234567890abcdef12\x1f" +
		"def4567\x1f" +
		"Jane Smith\x1f" +
		"jane@example.com\x1f" +
		"1704063600\x1f" +
		"Fix bug in login\x1f" +
		"\x1e"

	tests := []struct {
		name        string
		base        string
		head        string
		limit       int
		output      string
		err         error
		wantCount   int
		wantErr     bool
		wantSubject string
	}{
		{
			name:        "single commit",
			base:        "main",
			head:        "feature",
			limit:       0,
			output:      singleCommit,
			wantCount:   1,
			wantErr:     false,
			wantSubject: "Add new feature",
		},
		{
			name:        "multiple commits",
			base:        "main",
			head:        "feature",
			limit:       0,
			output:      multipleCommits,
			wantCount:   2,
			wantErr:     false,
			wantSubject: "Add new feature",
		},
		{
			name:      "no commits (empty range)",
			base:      "main",
			head:      "main",
			limit:     0,
			output:    "",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:        "with limit",
			base:        "main",
			head:        "feature",
			limit:       5,
			output:      singleCommit,
			wantCount:   1,
			wantErr:     false,
			wantSubject: "Add new feature",
		},
		{
			name:    "invalid base ref",
			base:    "nonexistent",
			head:    "main",
			limit:   0,
			err:     errors.New("fatal: bad revision 'nonexistent..main'"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("git log", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			commits, err := repo.GetCommitLog(tt.base, tt.head, tt.limit)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommitLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(commits) != tt.wantCount {
					t.Errorf("GetCommitLog() returned %d commits, want %d", len(commits), tt.wantCount)
					return
				}

				if tt.wantCount > 0 && commits[0].Subject != tt.wantSubject {
					t.Errorf("GetCommitLog() first commit subject = %q, want %q", commits[0].Subject, tt.wantSubject)
				}

				// Verify limit argument was passed
				if tt.limit > 0 {
					mock.AssertCalled(t, "git", "-n")
				}
			}
		})
	}
}

func TestParseCommitLog(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []Commit
		wantErr bool
	}{
		{
			name:  "empty input",
			input: "",
			want:  []Commit{},
		},
		{
			name:  "whitespace only",
			input: "  \n\t  ",
			want:  []Commit{},
		},
		{
			name: "single commit with body",
			input: "abc123def456789012345678901234567890abcd\x1f" +
				"abc123d\x1f" +
				"John Doe\x1f" +
				"john@example.com\x1f" +
				"1704067200\x1f" +
				"Add feature\x1f" +
				"This is the body\x1e",
			want: []Commit{
				{
					Hash:        "abc123def456789012345678901234567890abcd",
					ShortHash:   "abc123d",
					Author:      "John Doe",
					AuthorEmail: "john@example.com",
					Date:        time.Unix(1704067200, 0),
					Subject:     "Add feature",
					Body:        "This is the body",
				},
			},
		},
		{
			name: "commit without body",
			input: "abc123def456789012345678901234567890abcd\x1f" +
				"abc123d\x1f" +
				"John Doe\x1f" +
				"john@example.com\x1f" +
				"1704067200\x1f" +
				"Add feature\x1f" +
				"\x1e",
			want: []Commit{
				{
					Hash:        "abc123def456789012345678901234567890abcd",
					ShortHash:   "abc123d",
					Author:      "John Doe",
					AuthorEmail: "john@example.com",
					Date:        time.Unix(1704067200, 0),
					Subject:     "Add feature",
					Body:        "",
				},
			},
		},
		{
			name:    "malformed record - too few fields",
			input:   "abc123\x1fshort\x1fauthor\x1e",
			wantErr: true,
		},
		{
			name: "malformed record - invalid timestamp",
			input: "abc123def456789012345678901234567890abcd\x1f" +
				"abc123d\x1f" +
				"John Doe\x1f" +
				"john@example.com\x1f" +
				"not-a-number\x1f" +
				"Subject\x1f" +
				"\x1e",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commits, err := parseCommitLog(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseCommitLog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(commits) != len(tt.want) {
					t.Errorf("parseCommitLog() returned %d commits, want %d", len(commits), len(tt.want))
					return
				}

				for i, want := range tt.want {
					got := commits[i]
					if got.Hash != want.Hash {
						t.Errorf("commits[%d].Hash = %q, want %q", i, got.Hash, want.Hash)
					}
					if got.ShortHash != want.ShortHash {
						t.Errorf("commits[%d].ShortHash = %q, want %q", i, got.ShortHash, want.ShortHash)
					}
					if got.Author != want.Author {
						t.Errorf("commits[%d].Author = %q, want %q", i, got.Author, want.Author)
					}
					if got.AuthorEmail != want.AuthorEmail {
						t.Errorf("commits[%d].AuthorEmail = %q, want %q", i, got.AuthorEmail, want.AuthorEmail)
					}
					if !got.Date.Equal(want.Date) {
						t.Errorf("commits[%d].Date = %v, want %v", i, got.Date, want.Date)
					}
					if got.Subject != want.Subject {
						t.Errorf("commits[%d].Subject = %q, want %q", i, got.Subject, want.Subject)
					}
					if got.Body != want.Body {
						t.Errorf("commits[%d].Body = %q, want %q", i, got.Body, want.Body)
					}
				}
			}
		})
	}
}

func TestGetCommitMessages(t *testing.T) {
	tests := []struct {
		name         string
		base         string
		head         string
		output       string
		err          error
		wantMessages []string
		wantErr      bool
	}{
		{
			name:         "single message",
			base:         "main",
			head:         "feature",
			output:       "Add new feature\n",
			wantMessages: []string{"Add new feature"},
			wantErr:      false,
		},
		{
			name:         "multiple messages",
			base:         "main",
			head:         "feature",
			output:       "Add feature A\nFix bug B\nUpdate docs\n",
			wantMessages: []string{"Add feature A", "Fix bug B", "Update docs"},
			wantErr:      false,
		},
		{
			name:         "no commits",
			base:         "main",
			head:         "main",
			output:       "",
			wantMessages: []string{},
			wantErr:      false,
		},
		{
			name:    "error case",
			base:    "invalid",
			head:    "main",
			err:     errors.New("fatal: bad revision"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("git log", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			messages, err := repo.GetCommitMessages(tt.base, tt.head)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommitMessages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(messages) != len(tt.wantMessages) {
					t.Errorf("GetCommitMessages() returned %d messages, want %d", len(messages), len(tt.wantMessages))
					return
				}

				for i, want := range tt.wantMessages {
					if messages[i] != want {
						t.Errorf("messages[%d] = %q, want %q", i, messages[i], want)
					}
				}
			}
		})
	}
}

func TestGetCommit(t *testing.T) {
	commitOutput := "abc123def456789012345678901234567890abcd\x1f" +
		"abc123d\x1f" +
		"John Doe\x1f" +
		"john@example.com\x1f" +
		"1704067200\x1f" +
		"Add new feature\x1f" +
		"Detailed description\x1e"

	tests := []struct {
		name        string
		ref         string
		output      string
		err         error
		wantSubject string
		wantErr     bool
	}{
		{
			name:        "by commit hash",
			ref:         "abc123d",
			output:      commitOutput,
			wantSubject: "Add new feature",
			wantErr:     false,
		},
		{
			name:        "by branch name",
			ref:         "main",
			output:      commitOutput,
			wantSubject: "Add new feature",
			wantErr:     false,
		},
		{
			name:        "by tag",
			ref:         "v1.0.0",
			output:      commitOutput,
			wantSubject: "Add new feature",
			wantErr:     false,
		},
		{
			name:    "invalid ref",
			ref:     "nonexistent",
			err:     errors.New("fatal: bad object nonexistent"),
			wantErr: true,
		},
		{
			name:    "empty output",
			ref:     "empty",
			output:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("git log", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			commit, err := repo.GetCommit(tt.ref)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if commit.Subject != tt.wantSubject {
					t.Errorf("GetCommit() subject = %q, want %q", commit.Subject, tt.wantSubject)
				}
			}
		})
	}
}

func TestGetCommitCount(t *testing.T) {
	tests := []struct {
		name      string
		base      string
		head      string
		output    string
		err       error
		wantCount int
		wantErr   bool
	}{
		{
			name:      "multiple commits",
			base:      "main",
			head:      "feature",
			output:    "15\n",
			wantCount: 15,
			wantErr:   false,
		},
		{
			name:      "no commits",
			base:      "main",
			head:      "main",
			output:    "0\n",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:    "invalid refs",
			base:    "invalid",
			head:    "refs",
			err:     errors.New("fatal: bad revision"),
			wantErr: true,
		},
		{
			name:    "invalid output",
			base:    "main",
			head:    "feature",
			output:  "not-a-number\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("rev-list --count", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			count, err := repo.GetCommitCount(tt.base, tt.head)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCommitCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && count != tt.wantCount {
				t.Errorf("GetCommitCount() = %d, want %d", count, tt.wantCount)
			}
		})
	}
}

func TestGetMergeBase(t *testing.T) {
	tests := []struct {
		name     string
		ref1     string
		ref2     string
		output   string
		err      error
		wantHash string
		wantErr  bool
	}{
		{
			name:     "valid merge base",
			ref1:     "main",
			ref2:     "feature",
			output:   "abc123def456789012345678901234567890abcd\n",
			wantHash: "abc123def456789012345678901234567890abcd",
			wantErr:  false,
		},
		{
			name:    "no common ancestor",
			ref1:    "main",
			ref2:    "unrelated",
			err:     errors.New("fatal: Not a valid object name"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("merge-base", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			hash, err := repo.GetMergeBase(tt.ref1, tt.ref2)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetMergeBase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && hash != tt.wantHash {
				t.Errorf("GetMergeBase() = %q, want %q", hash, tt.wantHash)
			}
		})
	}
}

func TestGetHeadCommit(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		err      error
		wantHash string
		wantErr  bool
	}{
		{
			name:     "success",
			output:   "abc123def456789012345678901234567890abcd\n",
			wantHash: "abc123def456789012345678901234567890abcd",
			wantErr:  false,
		},
		{
			name:    "not a git repo",
			err:     errors.New("fatal: not a git repository"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("rev-parse HEAD", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			hash, err := repo.GetHeadCommit()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetHeadCommit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && hash != tt.wantHash {
				t.Errorf("GetHeadCommit() = %q, want %q", hash, tt.wantHash)
			}
		})
	}
}

func TestResolveRef(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		output   string
		err      error
		wantHash string
		wantErr  bool
	}{
		{
			name:     "branch name",
			ref:      "main",
			output:   "abc123def456789012345678901234567890abcd\n",
			wantHash: "abc123def456789012345678901234567890abcd",
			wantErr:  false,
		},
		{
			name:     "tag",
			ref:      "v1.0.0",
			output:   "def456789012345678901234567890abcdef12\n",
			wantHash: "def456789012345678901234567890abcdef12",
			wantErr:  false,
		},
		{
			name:     "HEAD~1",
			ref:      "HEAD~1",
			output:   "789012345678901234567890abcdef1234567890\n",
			wantHash: "789012345678901234567890abcdef1234567890",
			wantErr:  false,
		},
		{
			name:    "invalid ref",
			ref:     "nonexistent",
			err:     errors.New("fatal: ambiguous argument 'nonexistent'"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("rev-parse "+tt.ref, []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			hash, err := repo.ResolveRef(tt.ref)

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveRef() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && hash != tt.wantHash {
				t.Errorf("ResolveRef() = %q, want %q", hash, tt.wantHash)
			}
		})
	}
}

func TestCommit_Fields(t *testing.T) {
	// Test that Commit struct holds all expected fields correctly
	commit := Commit{
		Hash:        "abc123def456789012345678901234567890abcd",
		ShortHash:   "abc123d",
		Author:      "John Doe",
		AuthorEmail: "john@example.com",
		Date:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Subject:     "Add feature",
		Body:        "Detailed description",
	}

	if commit.Hash != "abc123def456789012345678901234567890abcd" {
		t.Error("Hash field not set correctly")
	}
	if commit.ShortHash != "abc123d" {
		t.Error("ShortHash field not set correctly")
	}
	if commit.Author != "John Doe" {
		t.Error("Author field not set correctly")
	}
	if commit.AuthorEmail != "john@example.com" {
		t.Error("AuthorEmail field not set correctly")
	}
	if commit.Subject != "Add feature" {
		t.Error("Subject field not set correctly")
	}
	if commit.Body != "Detailed description" {
		t.Error("Body field not set correctly")
	}
}

// Test argument construction
func TestGetCommitLog_Arguments(t *testing.T) {
	mock := newMockExecutor()
	mock.On("git log", []byte(""), nil)

	repo := NewWithExecutor(mock)
	_, _ = repo.GetCommitLog("main", "feature", 10)

	// Verify the command was called with correct arguments
	calls := mock.GetCalls()
	if len(calls) == 0 {
		t.Fatal("No calls recorded")
	}

	args := strings.Join(calls[0].args, " ")
	if !strings.Contains(args, "-n") {
		t.Error("Expected -n flag for limit")
	}
	if !strings.Contains(args, "10") {
		t.Error("Expected limit value 10")
	}
	if !strings.Contains(args, "main..feature") {
		t.Error("Expected revision range main..feature")
	}
}
