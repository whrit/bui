package git

import (
	"errors"
	"testing"
)

func TestListBranches(t *testing.T) {
	// Format: %(HEAD)%x1f%(refname:short)%x1f%(upstream:short)%x1f%(objectname:short)
	branchOutput := "*\x1fmain\x1forigin/main\x1fabc123d\n" +
		" \x1ffeature/login\x1forigin/feature/login\x1fdef4567\n" +
		" \x1fremotes/origin/main\x1f\x1fabc123d\n" +
		" \x1fremotes/origin/feature/login\x1f\x1fdef4567\n"

	tests := []struct {
		name      string
		output    string
		err       error
		wantCount int
		wantErr   bool
	}{
		{
			name:      "multiple branches",
			output:    branchOutput,
			wantCount: 4,
			wantErr:   false,
		},
		{
			name:      "single branch",
			output:    "*\x1fmain\x1forigin/main\x1fabc123d\n",
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "no branches",
			output:    "",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:    "git error",
			err:     errors.New("fatal: not a git repository"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("branch -a", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			branches, err := repo.ListBranches()

			if (err != nil) != tt.wantErr {
				t.Errorf("ListBranches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(branches) != tt.wantCount {
				t.Errorf("ListBranches() returned %d branches, want %d", len(branches), tt.wantCount)
			}
		})
	}
}

func TestListLocalBranches(t *testing.T) {
	localOutput := "*\x1fmain\x1forigin/main\x1fabc123d\n" +
		" \x1ffeature/login\x1forigin/feature/login\x1fdef4567\n" +
		" \x1fbugfix\x1f\x1f789abcd\n"

	tests := []struct {
		name      string
		output    string
		err       error
		wantCount int
		wantErr   bool
	}{
		{
			name:      "local branches only",
			output:    localOutput,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:    "error",
			err:     errors.New("fatal: not a git repository"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("git branch --format", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			branches, err := repo.ListLocalBranches()

			if (err != nil) != tt.wantErr {
				t.Errorf("ListLocalBranches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(branches) != tt.wantCount {
				t.Errorf("ListLocalBranches() returned %d branches, want %d", len(branches), tt.wantCount)
			}
		})
	}
}

func TestListRemoteBranches(t *testing.T) {
	remoteOutput := " \x1forigin/main\x1f\x1fabc123d\n" +
		" \x1forigin/feature\x1f\x1fdef4567\n"

	tests := []struct {
		name      string
		output    string
		err       error
		wantCount int
		wantErr   bool
	}{
		{
			name:      "remote branches",
			output:    remoteOutput,
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:    "error",
			err:     errors.New("fatal: not a git repository"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("branch -r", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			branches, err := repo.ListRemoteBranches()

			if (err != nil) != tt.wantErr {
				t.Errorf("ListRemoteBranches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(branches) != tt.wantCount {
					t.Errorf("ListRemoteBranches() returned %d branches, want %d", len(branches), tt.wantCount)
				}
				// Verify all are marked as remote
				for _, b := range branches {
					if !b.IsRemote {
						t.Errorf("Branch %q should be marked as remote", b.Name)
					}
				}
			}
		})
	}
}

func TestParseBranches(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCount int
		wantFirst Branch
	}{
		{
			name:      "empty input",
			input:     "",
			wantCount: 0,
		},
		{
			name:      "single current branch",
			input:     "*\x1fmain\x1forigin/main\x1fabc123d",
			wantCount: 1,
			wantFirst: Branch{
				Name:       "main",
				IsCurrent:  true,
				Upstream:   "origin/main",
				LastCommit: "abc123d",
			},
		},
		{
			name:      "branch without upstream",
			input:     " \x1ffeature\x1f\x1fdef4567",
			wantCount: 1,
			wantFirst: Branch{
				Name:       "feature",
				IsCurrent:  false,
				Upstream:   "",
				LastCommit: "def4567",
			},
		},
		{
			name:      "remote branch",
			input:     " \x1fremotes/origin/main\x1f\x1fabc123d",
			wantCount: 1,
			wantFirst: Branch{
				Name:       "origin/main",
				IsRemote:   true,
				LastCommit: "abc123d",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branches, err := parseBranches(tt.input)
			if err != nil {
				t.Fatalf("parseBranches() unexpected error: %v", err)
			}

			if len(branches) != tt.wantCount {
				t.Errorf("parseBranches() returned %d branches, want %d", len(branches), tt.wantCount)
				return
			}

			if tt.wantCount > 0 {
				got := branches[0]
				if got.Name != tt.wantFirst.Name {
					t.Errorf("branches[0].Name = %q, want %q", got.Name, tt.wantFirst.Name)
				}
				if got.IsCurrent != tt.wantFirst.IsCurrent {
					t.Errorf("branches[0].IsCurrent = %v, want %v", got.IsCurrent, tt.wantFirst.IsCurrent)
				}
				if got.IsRemote != tt.wantFirst.IsRemote {
					t.Errorf("branches[0].IsRemote = %v, want %v", got.IsRemote, tt.wantFirst.IsRemote)
				}
				if got.Upstream != tt.wantFirst.Upstream {
					t.Errorf("branches[0].Upstream = %q, want %q", got.Upstream, tt.wantFirst.Upstream)
				}
				if got.LastCommit != tt.wantFirst.LastCommit {
					t.Errorf("branches[0].LastCommit = %q, want %q", got.LastCommit, tt.wantFirst.LastCommit)
				}
			}
		})
	}
}

func TestGetBranch(t *testing.T) {
	branchOutput := "*\x1fmain\x1forigin/main\x1fabc123d\n"

	tests := []struct {
		name       string
		branchName string
		existsErr  error
		listOutput string
		listErr    error
		wantName   string
		wantErr    bool
	}{
		{
			name:       "existing branch",
			branchName: "main",
			existsErr:  nil,
			listOutput: branchOutput,
			wantName:   "main",
			wantErr:    false,
		},
		{
			name:       "non-existing branch",
			branchName: "nonexistent",
			existsErr:  errors.New("not found"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			// For BranchExists check
			mock.On("show-ref --verify --quiet refs/heads/"+tt.branchName, []byte(""), tt.existsErr)
			// For branch list
			mock.On("branch --list", []byte(tt.listOutput), tt.listErr)

			repo := NewWithExecutor(mock)
			branch, err := repo.GetBranch(tt.branchName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && branch.Name != tt.wantName {
				t.Errorf("GetBranch() name = %q, want %q", branch.Name, tt.wantName)
			}
		})
	}
}

func TestBranchExists(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		localErr   error
		remoteErr  error
		wantExists bool
	}{
		{
			name:       "local branch exists",
			branchName: "main",
			localErr:   nil,
			wantExists: true,
		},
		{
			name:       "remote branch exists",
			branchName: "feature",
			localErr:   errors.New("not found"),
			remoteErr:  nil,
			wantExists: true,
		},
		{
			name:       "branch does not exist",
			branchName: "nonexistent",
			localErr:   errors.New("not found"),
			remoteErr:  errors.New("not found"),
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("show-ref --verify --quiet refs/heads/"+tt.branchName, []byte(""), tt.localErr)
			mock.On("show-ref --verify --quiet refs/remotes/origin/"+tt.branchName, []byte(""), tt.remoteErr)

			repo := NewWithExecutor(mock)
			exists, _ := repo.BranchExists(tt.branchName)

			if exists != tt.wantExists {
				t.Errorf("BranchExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestGetUpstreamBranch(t *testing.T) {
	tests := []struct {
		name         string
		branchName   string
		output       string
		err          error
		wantUpstream string
		wantErr      bool
	}{
		{
			name:         "branch with upstream",
			branchName:   "main",
			output:       "origin/main\n",
			wantUpstream: "origin/main",
			wantErr:      false,
		},
		{
			name:       "branch without upstream",
			branchName: "local-only",
			err:        errors.New("fatal: no upstream configured for branch 'local-only'"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("rev-parse --abbrev-ref "+tt.branchName+"@{upstream}", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			upstream, err := repo.GetUpstreamBranch(tt.branchName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUpstreamBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && upstream != tt.wantUpstream {
				t.Errorf("GetUpstreamBranch() = %q, want %q", upstream, tt.wantUpstream)
			}
		})
	}
}

func TestGetDefaultBranch(t *testing.T) {
	tests := []struct {
		name           string
		configOutput   string
		configErr      error
		symbolicOutput string
		symbolicErr    error
		mainExists     bool
		masterExists   bool
		wantBranch     string
		wantErr        bool
	}{
		{
			name:         "from git config",
			configOutput: "main\n",
			wantBranch:   "main",
			wantErr:      false,
		},
		{
			name:           "from origin/HEAD",
			configErr:      errors.New("no config"),
			symbolicOutput: "refs/remotes/origin/main\n",
			wantBranch:     "main",
			wantErr:        false,
		},
		{
			name:        "fallback to main",
			configErr:   errors.New("no config"),
			symbolicErr: errors.New("no symbolic ref"),
			mainExists:  true,
			wantBranch:  "main",
			wantErr:     false,
		},
		{
			name:         "fallback to master",
			configErr:    errors.New("no config"),
			symbolicErr:  errors.New("no symbolic ref"),
			mainExists:   false,
			masterExists: true,
			wantBranch:   "master",
			wantErr:      false,
		},
		{
			name:         "no default found",
			configErr:    errors.New("no config"),
			symbolicErr:  errors.New("no symbolic ref"),
			mainExists:   false,
			masterExists: false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("config --get init.defaultBranch", []byte(tt.configOutput), tt.configErr)
			mock.On("symbolic-ref refs/remotes/origin/HEAD", []byte(tt.symbolicOutput), tt.symbolicErr)

			// BranchExists checks
			if tt.mainExists {
				mock.On("show-ref --verify --quiet refs/heads/main", []byte(""), nil)
			} else {
				mock.On("show-ref --verify --quiet refs/heads/main", []byte(""), errors.New("not found"))
				mock.On("show-ref --verify --quiet refs/remotes/origin/main", []byte(""), errors.New("not found"))
			}

			if tt.masterExists {
				mock.On("show-ref --verify --quiet refs/heads/master", []byte(""), nil)
			} else {
				mock.On("show-ref --verify --quiet refs/heads/master", []byte(""), errors.New("not found"))
				mock.On("show-ref --verify --quiet refs/remotes/origin/master", []byte(""), errors.New("not found"))
			}

			repo := NewWithExecutor(mock)
			branch, err := repo.GetDefaultBranch()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDefaultBranch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && branch != tt.wantBranch {
				t.Errorf("GetDefaultBranch() = %q, want %q", branch, tt.wantBranch)
			}
		})
	}
}

func TestSetUpstream(t *testing.T) {
	tests := []struct {
		name    string
		local   string
		remote  string
		err     error
		wantErr bool
	}{
		{
			name:    "success",
			local:   "feature",
			remote:  "origin/feature",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "error",
			local:   "feature",
			remote:  "invalid",
			err:     errors.New("fatal: not a valid ref"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("branch --set-upstream-to", []byte(""), tt.err)

			repo := NewWithExecutor(mock)
			err := repo.SetUpstream(tt.local, tt.remote)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetUpstream() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateBranch(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		startPoint string
		err        error
		wantErr    bool
	}{
		{
			name:       "create from HEAD",
			branchName: "new-feature",
			startPoint: "",
			err:        nil,
			wantErr:    false,
		},
		{
			name:       "create from specific commit",
			branchName: "hotfix",
			startPoint: "abc123d",
			err:        nil,
			wantErr:    false,
		},
		{
			name:       "branch already exists",
			branchName: "existing",
			err:        errors.New("fatal: A branch named 'existing' already exists"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("git branch", []byte(""), tt.err)

			repo := NewWithExecutor(mock)
			err := repo.CreateBranch(tt.branchName, tt.startPoint)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteBranch(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		force      bool
		err        error
		wantErr    bool
	}{
		{
			name:       "safe delete",
			branchName: "merged-feature",
			force:      false,
			err:        nil,
			wantErr:    false,
		},
		{
			name:       "force delete",
			branchName: "unmerged-feature",
			force:      true,
			err:        nil,
			wantErr:    false,
		},
		{
			name:       "safe delete fails (unmerged)",
			branchName: "unmerged",
			force:      false,
			err:        errors.New("error: The branch 'unmerged' is not fully merged"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("git branch", []byte(""), tt.err)

			repo := NewWithExecutor(mock)
			err := repo.DeleteBranch(tt.branchName, tt.force)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRenameBranch(t *testing.T) {
	tests := []struct {
		name    string
		oldName string
		newName string
		err     error
		wantErr bool
	}{
		{
			name:    "success",
			oldName: "old-feature",
			newName: "new-feature",
			err:     nil,
			wantErr: false,
		},
		{
			name:    "new name already exists",
			oldName: "feature",
			newName: "main",
			err:     errors.New("fatal: A branch named 'main' already exists"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("branch -m", []byte(""), tt.err)

			repo := NewWithExecutor(mock)
			err := repo.RenameBranch(tt.oldName, tt.newName)

			if (err != nil) != tt.wantErr {
				t.Errorf("RenameBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckoutBranch(t *testing.T) {
	tests := []struct {
		name       string
		branchName string
		err        error
		wantErr    bool
	}{
		{
			name:       "success",
			branchName: "feature",
			err:        nil,
			wantErr:    false,
		},
		{
			name:       "branch not found",
			branchName: "nonexistent",
			err:        errors.New("error: pathspec 'nonexistent' did not match any file(s)"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("checkout "+tt.branchName, []byte(""), tt.err)

			repo := NewWithExecutor(mock)
			err := repo.CheckoutBranch(tt.branchName)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckoutBranch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetBranchCommitsBehindAhead(t *testing.T) {
	tests := []struct {
		name        string
		branch      string
		upstreamOut string
		upstreamErr error
		revlistOut  string
		revlistErr  error
		wantBehind  int
		wantAhead   int
		wantErr     bool
	}{
		{
			name:        "branch is behind and ahead",
			branch:      "feature",
			upstreamOut: "origin/feature\n",
			revlistOut:  "3\t5\n",
			wantBehind:  3,
			wantAhead:   5,
			wantErr:     false,
		},
		{
			name:        "branch is only behind",
			branch:      "feature",
			upstreamOut: "origin/feature\n",
			revlistOut:  "10\t0\n",
			wantBehind:  10,
			wantAhead:   0,
			wantErr:     false,
		},
		{
			name:        "branch is up to date",
			branch:      "main",
			upstreamOut: "origin/main\n",
			revlistOut:  "0\t0\n",
			wantBehind:  0,
			wantAhead:   0,
			wantErr:     false,
		},
		{
			name:        "no upstream",
			branch:      "local-only",
			upstreamErr: errors.New("fatal: no upstream"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("rev-parse --abbrev-ref "+tt.branch+"@{upstream}", []byte(tt.upstreamOut), tt.upstreamErr)
			mock.On("rev-list --left-right --count", []byte(tt.revlistOut), tt.revlistErr)

			repo := NewWithExecutor(mock)
			behind, ahead, err := repo.GetBranchCommitsBehindAhead(tt.branch)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBranchCommitsBehindAhead() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if behind != tt.wantBehind {
					t.Errorf("GetBranchCommitsBehindAhead() behind = %d, want %d", behind, tt.wantBehind)
				}
				if ahead != tt.wantAhead {
					t.Errorf("GetBranchCommitsBehindAhead() ahead = %d, want %d", ahead, tt.wantAhead)
				}
			}
		})
	}
}

func TestBranch_Fields(t *testing.T) {
	branch := Branch{
		Name:       "feature/login",
		IsRemote:   false,
		IsCurrent:  true,
		Upstream:   "origin/feature/login",
		LastCommit: "abc123d",
	}

	if branch.Name != "feature/login" {
		t.Error("Name field not set correctly")
	}
	if branch.IsRemote != false {
		t.Error("IsRemote field not set correctly")
	}
	if branch.IsCurrent != true {
		t.Error("IsCurrent field not set correctly")
	}
	if branch.Upstream != "origin/feature/login" {
		t.Error("Upstream field not set correctly")
	}
	if branch.LastCommit != "abc123d" {
		t.Error("LastCommit field not set correctly")
	}
}

func TestGetTrackingBranches(t *testing.T) {
	branchOutput := "*\x1fmain\x1forigin/main\x1fabc123d\n" +
		" \x1ffeature\x1forigin/feature\x1fdef4567\n" +
		" \x1flocal-only\x1f\x1f789abcd\n"

	tests := []struct {
		name      string
		output    string
		err       error
		wantCount int
		wantErr   bool
	}{
		{
			name:      "mixed branches",
			output:    branchOutput,
			wantCount: 2, // Only main and feature have upstream
			wantErr:   false,
		},
		{
			name:    "error",
			err:     errors.New("fatal: not a git repository"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("git branch --format", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			branches, err := repo.GetTrackingBranches()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTrackingBranches() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(branches) != tt.wantCount {
					t.Errorf("GetTrackingBranches() returned %d branches, want %d", len(branches), tt.wantCount)
				}
				// Verify all returned branches have upstream
				for _, b := range branches {
					if b.Upstream == "" {
						t.Errorf("Branch %q should have upstream", b.Name)
					}
				}
			}
		})
	}
}
