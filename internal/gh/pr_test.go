package gh

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

// mockExecutor implements CommandExecutor for testing.
// It is thread-safe for concurrent access.
type mockExecutor struct {
	mu sync.RWMutex
	// responses maps command patterns to responses
	responses map[string]mockResponse
	// calls records all commands executed
	calls []mockCall
}

type mockResponse struct {
	output []byte
	err    error
}

type mockCall struct {
	name string
	args []string
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{
		responses: make(map[string]mockResponse),
		calls:     make([]mockCall, 0),
	}
}

func (m *mockExecutor) On(pattern string, output []byte, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[pattern] = mockResponse{output: output, err: err}
}

func (m *mockExecutor) Run(name string, args ...string) ([]byte, error) {
	m.mu.Lock()
	m.calls = append(m.calls, mockCall{name: name, args: args})
	m.mu.Unlock()

	// Build the full command for pattern matching
	fullCmd := name + " " + strings.Join(args, " ")

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Try exact match first
	if resp, ok := m.responses[fullCmd]; ok {
		return resp.output, resp.err
	}

	// Try pattern matching (check if command contains pattern)
	for pattern, resp := range m.responses {
		if strings.Contains(fullCmd, pattern) {
			return resp.output, resp.err
		}
	}

	return nil, errors.New("no mock response configured for: " + fullCmd)
}

func (m *mockExecutor) AssertCalled(t *testing.T, name string, argsContain ...string) {
	t.Helper()
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, call := range m.calls {
		if call.name == name {
			fullArgs := strings.Join(call.args, " ")
			allFound := true
			for _, arg := range argsContain {
				if !strings.Contains(fullArgs, arg) {
					allFound = false
					break
				}
			}
			if allFound {
				return
			}
		}
	}
	t.Errorf("expected call to %q with args containing %v, got calls: %+v", name, argsContain, m.calls)
}

func TestListPRs(t *testing.T) {
	tests := []struct {
		name       string
		filters    PRFilters
		mockOutput string
		mockErr    error
		wantCount  int
		wantErr    bool
		wantArgs   []string
	}{
		{
			name:    "list open PRs with default filters",
			filters: PRFilters{State: "open", Limit: 10},
			mockOutput: `[
				{
					"number": 1,
					"title": "Test PR",
					"body": "Test body",
					"state": "OPEN",
					"author": {"login": "testuser"},
					"headRefName": "feature-branch",
					"baseRefName": "main",
					"url": "https://github.com/owner/repo/pull/1",
					"createdAt": "2024-01-15T10:00:00Z",
					"updatedAt": "2024-01-15T12:00:00Z",
					"isDraft": false,
					"mergeable": "MERGEABLE",
					"additions": 100,
					"deletions": 50,
					"labels": [{"name": "bug", "color": "ff0000"}],
					"reviewRequests": []
				}
			]`,
			wantCount: 1,
			wantErr:   false,
			wantArgs:  []string{"pr", "list", "--state", "open", "--limit", "10", "--json"},
		},
		{
			name:    "list PRs with author filter",
			filters: PRFilters{State: "all", Author: "octocat", Limit: 5},
			mockOutput: `[
				{
					"number": 2,
					"title": "Another PR",
					"body": "",
					"state": "MERGED",
					"author": {"login": "octocat"},
					"headRefName": "fix-bug",
					"baseRefName": "main",
					"url": "https://github.com/owner/repo/pull/2",
					"createdAt": "2024-01-10T08:00:00Z",
					"updatedAt": "2024-01-12T09:00:00Z",
					"isDraft": false,
					"mergeable": "UNKNOWN",
					"additions": 20,
					"deletions": 5,
					"labels": [],
					"reviewRequests": []
				}
			]`,
			wantCount: 1,
			wantErr:   false,
			wantArgs:  []string{"pr", "list", "--author", "octocat"},
		},
		{
			name:    "list PRs with labels filter",
			filters: PRFilters{State: "open", Labels: []string{"bug", "urgent"}, Limit: 10},
			mockOutput: `[
				{
					"number": 3,
					"title": "Bug fix",
					"body": "Fixes critical bug",
					"state": "OPEN",
					"author": {"login": "developer"},
					"headRefName": "bugfix",
					"baseRefName": "main",
					"url": "https://github.com/owner/repo/pull/3",
					"createdAt": "2024-01-14T10:00:00Z",
					"updatedAt": "2024-01-14T11:00:00Z",
					"isDraft": false,
					"mergeable": "MERGEABLE",
					"additions": 10,
					"deletions": 2,
					"labels": [{"name": "bug", "color": "ff0000"}, {"name": "urgent", "color": "ff00ff"}],
					"reviewRequests": []
				}
			]`,
			wantCount: 1,
			wantErr:   false,
			wantArgs:  []string{"pr", "list", "--label", "bug", "--label", "urgent"},
		},
		{
			name:       "handle gh CLI error",
			filters:    PRFilters{State: "open", Limit: 10},
			mockOutput: "",
			mockErr:    errors.New("gh: not logged in"),
			wantCount:  0,
			wantErr:    true,
		},
		{
			name:       "handle invalid JSON response",
			filters:    PRFilters{State: "open", Limit: 10},
			mockOutput: "invalid json",
			mockErr:    nil,
			wantCount:  0,
			wantErr:    true,
		},
		{
			name:       "handle empty list",
			filters:    PRFilters{State: "open", Limit: 10},
			mockOutput: "[]",
			wantCount:  0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr list", []byte(tt.mockOutput), tt.mockErr)

			client := NewWithExecutor(mock)
			prs, err := client.ListPRs(tt.filters)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListPRs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(prs) != tt.wantCount {
					t.Errorf("ListPRs() returned %d PRs, want %d", len(prs), tt.wantCount)
				}

				// Verify command arguments
				if len(tt.wantArgs) > 0 {
					mock.AssertCalled(t, "gh", tt.wantArgs...)
				}
			}
		})
	}
}

func TestGetPR(t *testing.T) {
	tests := []struct {
		name       string
		prNumber   int
		mockOutput string
		mockErr    error
		wantTitle  string
		wantErr    bool
	}{
		{
			name:     "get existing PR",
			prNumber: 42,
			mockOutput: `{
				"number": 42,
				"title": "Add new feature",
				"body": "This PR adds a new feature",
				"state": "OPEN",
				"author": {"login": "developer"},
				"headRefName": "feature/new-thing",
				"baseRefName": "main",
				"url": "https://github.com/owner/repo/pull/42",
				"createdAt": "2024-01-15T10:00:00Z",
				"updatedAt": "2024-01-15T14:00:00Z",
				"isDraft": true,
				"mergeable": "MERGEABLE",
				"additions": 250,
				"deletions": 30,
				"labels": [{"name": "feature", "color": "00ff00"}],
				"reviewRequests": [{"login": "reviewer1", "state": "PENDING"}]
			}`,
			wantTitle: "Add new feature",
			wantErr:   false,
		},
		{
			name:       "PR not found",
			prNumber:   999,
			mockOutput: "",
			mockErr:    errors.New("GraphQL: Could not resolve to a PullRequest"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr view", []byte(tt.mockOutput), tt.mockErr)

			client := NewWithExecutor(mock)
			pr, err := client.GetPR(tt.prNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if pr.Title != tt.wantTitle {
					t.Errorf("GetPR() title = %q, want %q", pr.Title, tt.wantTitle)
				}
				if pr.Number != tt.prNumber {
					t.Errorf("GetPR() number = %d, want %d", pr.Number, tt.prNumber)
				}
			}
		})
	}
}

func TestGetPRDiff(t *testing.T) {
	tests := []struct {
		name       string
		prNumber   int
		mockOutput string
		mockErr    error
		wantDiff   string
		wantErr    bool
	}{
		{
			name:     "get diff successfully",
			prNumber: 42,
			mockOutput: `diff --git a/file.go b/file.go
index 1234567..abcdefg 100644
--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 package main
+
+import "fmt"
`,
			wantDiff: "diff --git a/file.go b/file.go",
			wantErr:  false,
		},
		{
			name:       "PR not found",
			prNumber:   999,
			mockOutput: "",
			mockErr:    errors.New("could not find PR"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr diff", []byte(tt.mockOutput), tt.mockErr)

			client := NewWithExecutor(mock)
			diff, err := client.GetPRDiff(tt.prNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPRDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !strings.Contains(diff, tt.wantDiff) {
				t.Errorf("GetPRDiff() diff does not contain %q", tt.wantDiff)
			}
		})
	}
}

func TestCreatePR(t *testing.T) {
	tests := []struct {
		name       string
		base       string
		head       string
		title      string
		body       string
		isDraft    bool
		mockOutput string
		mockErr    error
		wantNumber int
		wantErr    bool
		wantArgs   []string
	}{
		{
			name:    "create regular PR",
			base:    "main",
			head:    "feature-branch",
			title:   "New feature",
			body:    "This adds a new feature",
			isDraft: false,
			mockOutput: `{
				"number": 100,
				"title": "New feature",
				"body": "This adds a new feature",
				"state": "OPEN",
				"author": {"login": "developer"},
				"headRefName": "feature-branch",
				"baseRefName": "main",
				"url": "https://github.com/owner/repo/pull/100",
				"createdAt": "2024-01-15T10:00:00Z",
				"updatedAt": "2024-01-15T10:00:00Z",
				"isDraft": false,
				"mergeable": "UNKNOWN",
				"additions": 0,
				"deletions": 0,
				"labels": [],
				"reviewRequests": []
			}`,
			wantNumber: 100,
			wantErr:    false,
			wantArgs:   []string{"pr", "create", "--base", "main", "--head", "feature-branch", "--title", "New feature"},
		},
		{
			name:    "create draft PR",
			base:    "main",
			head:    "wip-branch",
			title:   "WIP: Draft feature",
			body:    "Work in progress",
			isDraft: true,
			mockOutput: `{
				"number": 101,
				"title": "WIP: Draft feature",
				"body": "Work in progress",
				"state": "OPEN",
				"author": {"login": "developer"},
				"headRefName": "wip-branch",
				"baseRefName": "main",
				"url": "https://github.com/owner/repo/pull/101",
				"createdAt": "2024-01-15T10:00:00Z",
				"updatedAt": "2024-01-15T10:00:00Z",
				"isDraft": true,
				"mergeable": "UNKNOWN",
				"additions": 0,
				"deletions": 0,
				"labels": [],
				"reviewRequests": []
			}`,
			wantNumber: 101,
			wantErr:    false,
			wantArgs:   []string{"pr", "create", "--draft"},
		},
		{
			name:       "create fails - no commits",
			base:       "main",
			head:       "empty-branch",
			title:      "Empty PR",
			body:       "No changes",
			isDraft:    false,
			mockOutput: "",
			mockErr:    errors.New("pull request create failed: No commits between main and empty-branch"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr create", []byte(tt.mockOutput), tt.mockErr)

			client := NewWithExecutor(mock)
			pr, err := client.CreatePR(tt.base, tt.head, tt.title, tt.body, tt.isDraft)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if pr.Number != tt.wantNumber {
					t.Errorf("CreatePR() number = %d, want %d", pr.Number, tt.wantNumber)
				}

				if len(tt.wantArgs) > 0 {
					mock.AssertCalled(t, "gh", tt.wantArgs...)
				}
			}
		})
	}
}

func TestMergePR(t *testing.T) {
	tests := []struct {
		name     string
		prNumber int
		method   MergeMethod
		mockErr  error
		wantErr  bool
		wantArgs []string
	}{
		{
			name:     "merge with merge commit",
			prNumber: 42,
			method:   MergeMethodMerge,
			mockErr:  nil,
			wantErr:  false,
			wantArgs: []string{"pr", "merge", "42", "--merge"},
		},
		{
			name:     "merge with squash",
			prNumber: 43,
			method:   MergeMethodSquash,
			mockErr:  nil,
			wantErr:  false,
			wantArgs: []string{"pr", "merge", "43", "--squash"},
		},
		{
			name:     "merge with rebase",
			prNumber: 44,
			method:   MergeMethodRebase,
			mockErr:  nil,
			wantErr:  false,
			wantArgs: []string{"pr", "merge", "44", "--rebase"},
		},
		{
			name:     "merge fails - not mergeable",
			prNumber: 45,
			method:   MergeMethodMerge,
			mockErr:  errors.New("Pull request #45 is not mergeable: the base branch policy prohibits the merge"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr merge", []byte(""), tt.mockErr)

			client := NewWithExecutor(mock)
			err := client.MergePR(tt.prNumber, tt.method)

			if (err != nil) != tt.wantErr {
				t.Errorf("MergePR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(tt.wantArgs) > 0 {
				mock.AssertCalled(t, "gh", tt.wantArgs...)
			}
		})
	}
}

func TestClosePR(t *testing.T) {
	tests := []struct {
		name     string
		prNumber int
		mockErr  error
		wantErr  bool
	}{
		{
			name:     "close PR successfully",
			prNumber: 42,
			mockErr:  nil,
			wantErr:  false,
		},
		{
			name:     "close fails - already closed",
			prNumber: 43,
			mockErr:  errors.New("Pull request #43 is already closed"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("pr close", []byte(""), tt.mockErr)

			client := NewWithExecutor(mock)
			err := client.ClosePR(tt.prNumber)

			if (err != nil) != tt.wantErr {
				t.Errorf("ClosePR() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPRFilters_BuildArgs(t *testing.T) {
	tests := []struct {
		name     string
		filters  PRFilters
		wantArgs []string
	}{
		{
			name:     "empty filters",
			filters:  PRFilters{},
			wantArgs: []string{},
		},
		{
			name:     "state filter only",
			filters:  PRFilters{State: "open"},
			wantArgs: []string{"--state", "open"},
		},
		{
			name:     "all filters",
			filters:  PRFilters{State: "all", Author: "user", Base: "main", Head: "feature", Labels: []string{"bug"}, Limit: 10},
			wantArgs: []string{"--state", "all", "--author", "user", "--base", "main", "--head", "feature", "--label", "bug", "--limit", "10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.filters.buildArgs()

			// Check that all expected args are present
			for _, want := range tt.wantArgs {
				found := false
				for _, got := range args {
					if got == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("buildArgs() missing arg %q, got %v", want, args)
				}
			}
		})
	}
}
