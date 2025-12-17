package git

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
	dir  string
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

	return m.findResponse(name, args)
}

func (m *mockExecutor) RunInDir(dir, name string, args ...string) ([]byte, error) {
	m.mu.Lock()
	m.calls = append(m.calls, mockCall{name: name, args: args, dir: dir})
	m.mu.Unlock()

	return m.findResponse(name, args)
}

func (m *mockExecutor) findResponse(name string, args []string) ([]byte, error) {
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

func (m *mockExecutor) GetCalls() []mockCall {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]mockCall, len(m.calls))
	copy(result, m.calls)
	return result
}

// Test constructors
func TestNew(t *testing.T) {
	repo := New()
	if repo == nil {
		t.Fatal("New() returned nil")
	}
	if repo.executor == nil {
		t.Error("New() repo has nil executor")
	}
}

func TestNewWithExecutor(t *testing.T) {
	mock := newMockExecutor()
	repo := NewWithExecutor(mock)

	if repo == nil {
		t.Fatal("NewWithExecutor() returned nil")
	}
	if repo.executor != mock {
		t.Error("NewWithExecutor() did not set provided executor")
	}
}

// Test parseRemoteURL
func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantName  string
		wantErr   bool
	}{
		{
			name:      "HTTPS URL with .git suffix",
			url:       "https://github.com/owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS URL without .git suffix",
			url:       "https://github.com/owner/repo",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS URL with www prefix",
			url:       "https://www.github.com/owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "SSH URL with .git suffix",
			url:       "git@github.com:owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "SSH URL without .git suffix",
			url:       "git@github.com:owner/repo",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "organization owner",
			url:       "https://github.com/my-org/my-repo.git",
			wantOwner: "my-org",
			wantName:  "my-repo",
			wantErr:   false,
		},
		{
			name:      "repo with numbers",
			url:       "git@github.com:user123/project456.git",
			wantOwner: "user123",
			wantName:  "project456",
			wantErr:   false,
		},
		{
			name:    "invalid URL - not GitHub",
			url:     "https://gitlab.com/owner/repo.git",
			wantErr: true,
		},
		{
			name:    "invalid URL - malformed",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "invalid URL - empty",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, name, err := parseRemoteURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRemoteURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if owner != tt.wantOwner {
					t.Errorf("parseRemoteURL() owner = %q, want %q", owner, tt.wantOwner)
				}
				if name != tt.wantName {
					t.Errorf("parseRemoteURL() name = %q, want %q", name, tt.wantName)
				}
			}
		})
	}
}

// Test GetRepoInfo
func TestGetRepoInfo(t *testing.T) {
	tests := []struct {
		name          string
		remoteURL     string
		remoteErr     error
		rootPath      string
		rootErr       error
		wantOwner     string
		wantName      string
		wantErr       bool
		wantErrSubstr string
	}{
		{
			name:      "successful with HTTPS URL",
			remoteURL: "https://github.com/octocat/hello-world.git",
			rootPath:  "/home/user/projects/hello-world",
			wantOwner: "octocat",
			wantName:  "hello-world",
			wantErr:   false,
		},
		{
			name:      "successful with SSH URL",
			remoteURL: "git@github.com:myorg/myproject.git",
			rootPath:  "/Users/dev/code/myproject",
			wantOwner: "myorg",
			wantName:  "myproject",
			wantErr:   false,
		},
		{
			name:          "remote URL fetch fails",
			remoteErr:     errors.New("not a git repository"),
			wantErr:       true,
			wantErrSubstr: "failed to get remote URL",
		},
		{
			name:          "empty remote URL",
			remoteURL:     "",
			wantErr:       true,
			wantErrSubstr: "no remote URL configured",
		},
		{
			name:          "invalid remote URL format",
			remoteURL:     "https://bitbucket.org/owner/repo.git",
			wantErr:       true,
			wantErrSubstr: "failed to parse remote URL",
		},
		{
			name:          "root path fetch fails",
			remoteURL:     "https://github.com/owner/repo.git",
			rootErr:       errors.New("fatal: not a git repository"),
			wantErr:       true,
			wantErrSubstr: "failed to get repository root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()

			// Setup mock for remote URL
			if tt.remoteErr != nil {
				mock.On("remote get-url origin", nil, tt.remoteErr)
			} else {
				mock.On("remote get-url origin", []byte(tt.remoteURL+"\n"), nil)
			}

			// Setup mock for root path
			if tt.rootErr != nil {
				mock.On("rev-parse --show-toplevel", nil, tt.rootErr)
			} else if tt.rootPath != "" {
				mock.On("rev-parse --show-toplevel", []byte(tt.rootPath+"\n"), nil)
			}

			repo := NewWithExecutor(mock)
			info, err := repo.GetRepoInfo()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRepoInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.wantErrSubstr != "" && !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Errorf("GetRepoInfo() error = %q, want error containing %q", err, tt.wantErrSubstr)
				}
				return
			}

			if info.Owner != tt.wantOwner {
				t.Errorf("GetRepoInfo() owner = %q, want %q", info.Owner, tt.wantOwner)
			}
			if info.Name != tt.wantName {
				t.Errorf("GetRepoInfo() name = %q, want %q", info.Name, tt.wantName)
			}
			if info.RootPath != tt.rootPath {
				t.Errorf("GetRepoInfo() rootPath = %q, want %q", info.RootPath, tt.rootPath)
			}
		})
	}
}

// Test GetRootPath
func TestGetRootPath(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		err      error
		wantPath string
		wantErr  bool
	}{
		{
			name:     "successful",
			output:   "/home/user/projects/myrepo\n",
			wantPath: "/home/user/projects/myrepo",
			wantErr:  false,
		},
		{
			name:     "path with trailing whitespace",
			output:   "/path/to/repo  \n\n",
			wantPath: "/path/to/repo",
			wantErr:  false,
		},
		{
			name:    "not a git repository",
			err:     errors.New("fatal: not a git repository"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("rev-parse --show-toplevel", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			path, err := repo.GetRootPath()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRootPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && path != tt.wantPath {
				t.Errorf("GetRootPath() = %q, want %q", path, tt.wantPath)
			}
		})
	}
}

// Test IsInsideWorkTree
func TestIsInsideWorkTree(t *testing.T) {
	tests := []struct {
		name   string
		output string
		err    error
		want   bool
	}{
		{
			name:   "inside work tree",
			output: "true\n",
			want:   true,
		},
		{
			name:   "not inside work tree",
			output: "false\n",
			want:   false,
		},
		{
			name: "command fails - not a git repo",
			err:  errors.New("fatal: not a git repository"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("rev-parse --is-inside-work-tree", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			result, _ := repo.IsInsideWorkTree()

			if result != tt.want {
				t.Errorf("IsInsideWorkTree() = %v, want %v", result, tt.want)
			}
		})
	}
}

// Test GetRemoteURL
func TestGetRemoteURL(t *testing.T) {
	tests := []struct {
		name    string
		remote  string
		output  string
		err     error
		wantURL string
		wantErr bool
	}{
		{
			name:    "origin remote",
			remote:  "origin",
			output:  "https://github.com/owner/repo.git\n",
			wantURL: "https://github.com/owner/repo.git",
			wantErr: false,
		},
		{
			name:    "upstream remote",
			remote:  "upstream",
			output:  "git@github.com:upstream/repo.git\n",
			wantURL: "git@github.com:upstream/repo.git",
			wantErr: false,
		},
		{
			name:    "remote not found",
			remote:  "nonexistent",
			err:     errors.New("fatal: No such remote 'nonexistent'"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("remote get-url "+tt.remote, []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			url, err := repo.GetRemoteURL(tt.remote)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRemoteURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && url != tt.wantURL {
				t.Errorf("GetRemoteURL() = %q, want %q", url, tt.wantURL)
			}
		})
	}
}

// Test ListRemotes
func TestListRemotes(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		err         error
		wantRemotes []string
		wantErr     bool
	}{
		{
			name:        "single remote",
			output:      "origin\n",
			wantRemotes: []string{"origin"},
			wantErr:     false,
		},
		{
			name:        "multiple remotes",
			output:      "origin\nupstream\nfork\n",
			wantRemotes: []string{"origin", "upstream", "fork"},
			wantErr:     false,
		},
		{
			name:        "no remotes",
			output:      "",
			wantRemotes: []string{},
			wantErr:     false,
		},
		{
			name:    "command fails",
			err:     errors.New("fatal: not a git repository"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("git remote", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			remotes, err := repo.ListRemotes()

			if (err != nil) != tt.wantErr {
				t.Errorf("ListRemotes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(remotes) != len(tt.wantRemotes) {
					t.Errorf("ListRemotes() returned %d remotes, want %d", len(remotes), len(tt.wantRemotes))
					return
				}
				for i, want := range tt.wantRemotes {
					if remotes[i] != want {
						t.Errorf("ListRemotes()[%d] = %q, want %q", i, remotes[i], want)
					}
				}
			}
		})
	}
}

// Test mock executor functionality
func TestMockExecutor_RecordsCalls(t *testing.T) {
	mock := newMockExecutor()
	mock.On("test", []byte("output"), nil)

	_, _ = mock.Run("git", "test", "arg1", "arg2")

	calls := mock.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}

	call := calls[0]
	if call.name != "git" {
		t.Errorf("Call name = %q, want %q", call.name, "git")
	}
	if len(call.args) != 3 {
		t.Errorf("Call args count = %d, want 3", len(call.args))
	}
}

func TestMockExecutor_RunInDir(t *testing.T) {
	mock := newMockExecutor()
	mock.On("test", []byte("output"), nil)

	output, err := mock.RunInDir("/some/dir", "git", "test")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if string(output) != "output" {
		t.Errorf("Output = %q, want %q", string(output), "output")
	}

	calls := mock.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}
	if calls[0].dir != "/some/dir" {
		t.Errorf("Call dir = %q, want %q", calls[0].dir, "/some/dir")
	}
}

// Test concurrent access to Repo
func TestRepo_ConcurrentAccess(t *testing.T) {
	mock := newMockExecutor()
	mock.On("rev-parse --show-toplevel", []byte("/path/to/repo\n"), nil)
	mock.On("rev-parse --is-inside-work-tree", []byte("true\n"), nil)
	mock.On("remote get-url origin", []byte("https://github.com/owner/repo.git\n"), nil)
	mock.On("git remote", []byte("origin\n"), nil)

	repo := NewWithExecutor(mock)
	done := make(chan bool)

	// Run multiple goroutines concurrently
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = repo.GetRootPath()
			done <- true
		}()
		go func() {
			_, _ = repo.IsInsideWorkTree()
			done <- true
		}()
		go func() {
			_, _ = repo.GetRemoteURL("origin")
			done <- true
		}()
		go func() {
			_, _ = repo.ListRemotes()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 40; i++ {
		<-done
	}
}
