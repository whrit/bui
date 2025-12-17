package git

import (
	"errors"
	"strings"
	"testing"
)

func TestGetDiff(t *testing.T) {
	exampleDiff := `diff --git a/file.go b/file.go
index 1234567..abcdefg 100644
--- a/file.go
+++ b/file.go
@@ -1,3 +1,4 @@
 package main
+
+import "fmt"
`

	tests := []struct {
		name     string
		base     string
		head     string
		output   string
		err      error
		wantDiff string
		wantErr  bool
	}{
		{
			name:     "successful diff",
			base:     "main",
			head:     "feature",
			output:   exampleDiff,
			wantDiff: "diff --git a/file.go b/file.go",
			wantErr:  false,
		},
		{
			name:     "empty diff (no changes)",
			base:     "main",
			head:     "main",
			output:   "",
			wantDiff: "",
			wantErr:  false,
		},
		{
			name:    "invalid reference",
			base:    "nonexistent",
			head:    "main",
			err:     errors.New("fatal: bad revision 'nonexistent..main'"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("diff "+tt.base+".."+tt.head, []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			diff, err := repo.GetDiff(tt.base, tt.head)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !strings.Contains(diff, tt.wantDiff) {
				t.Errorf("GetDiff() = %q, want to contain %q", diff, tt.wantDiff)
			}
		})
	}
}

func TestGetDiffStat(t *testing.T) {
	tests := []struct {
		name             string
		output           string
		err              error
		wantFilesChanged int
		wantAdditions    int
		wantDeletions    int
		wantErr          bool
	}{
		{
			name: "multiple files with additions and deletions",
			output: ` file1.go | 10 +++++++---
 file2.go | 5 +++++
 file3.go | 3 ---
 3 files changed, 12 insertions(+), 6 deletions(-)
`,
			wantFilesChanged: 3,
			wantAdditions:    12,
			wantDeletions:    6,
			wantErr:          false,
		},
		{
			name: "single file with only additions",
			output: ` newfile.go | 50 +++++++++++++++++++++++
 1 file changed, 50 insertions(+)
`,
			wantFilesChanged: 1,
			wantAdditions:    50,
			wantDeletions:    0,
			wantErr:          false,
		},
		{
			name: "single file with only deletions",
			output: ` oldfile.go | 25 -------------------------
 1 file changed, 25 deletions(-)
`,
			wantFilesChanged: 1,
			wantAdditions:    0,
			wantDeletions:    25,
			wantErr:          false,
		},
		{
			name:             "no changes",
			output:           "",
			wantFilesChanged: 0,
			wantAdditions:    0,
			wantDeletions:    0,
			wantErr:          false,
		},
		{
			name:    "git error",
			err:     errors.New("fatal: bad revision"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("diff --stat", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			stat, err := repo.GetDiffStat("main", "feature")

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDiffStat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if stat.FilesChanged != tt.wantFilesChanged {
					t.Errorf("GetDiffStat() FilesChanged = %d, want %d", stat.FilesChanged, tt.wantFilesChanged)
				}
				if stat.Additions != tt.wantAdditions {
					t.Errorf("GetDiffStat() Additions = %d, want %d", stat.Additions, tt.wantAdditions)
				}
				if stat.Deletions != tt.wantDeletions {
					t.Errorf("GetDiffStat() Deletions = %d, want %d", stat.Deletions, tt.wantDeletions)
				}
			}
		})
	}
}

func TestParseDiffStat(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		wantFilesChanged int
		wantAdditions    int
		wantDeletions    int
	}{
		{
			name:             "empty input",
			input:            "",
			wantFilesChanged: 0,
			wantAdditions:    0,
			wantDeletions:    0,
		},
		{
			name:             "full stats",
			input:            "5 files changed, 100 insertions(+), 50 deletions(-)",
			wantFilesChanged: 5,
			wantAdditions:    100,
			wantDeletions:    50,
		},
		{
			name:             "singular file",
			input:            "1 file changed, 10 insertions(+), 2 deletions(-)",
			wantFilesChanged: 1,
			wantAdditions:    10,
			wantDeletions:    2,
		},
		{
			name:             "only additions",
			input:            "2 files changed, 30 insertions(+)",
			wantFilesChanged: 2,
			wantAdditions:    30,
			wantDeletions:    0,
		},
		{
			name:             "only deletions",
			input:            "1 file changed, 15 deletions(-)",
			wantFilesChanged: 1,
			wantAdditions:    0,
			wantDeletions:    15,
		},
		{
			name:             "singular insertion",
			input:            "1 file changed, 1 insertion(+)",
			wantFilesChanged: 1,
			wantAdditions:    1,
			wantDeletions:    0,
		},
		{
			name:             "singular deletion",
			input:            "1 file changed, 1 deletion(-)",
			wantFilesChanged: 1,
			wantAdditions:    0,
			wantDeletions:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stat, err := parseDiffStat(tt.input)
			if err != nil {
				t.Fatalf("parseDiffStat() unexpected error: %v", err)
			}

			if stat.FilesChanged != tt.wantFilesChanged {
				t.Errorf("parseDiffStat() FilesChanged = %d, want %d", stat.FilesChanged, tt.wantFilesChanged)
			}
			if stat.Additions != tt.wantAdditions {
				t.Errorf("parseDiffStat() Additions = %d, want %d", stat.Additions, tt.wantAdditions)
			}
			if stat.Deletions != tt.wantDeletions {
				t.Errorf("parseDiffStat() Deletions = %d, want %d", stat.Deletions, tt.wantDeletions)
			}
		})
	}
}

func TestGetFileDiffs(t *testing.T) {
	tests := []struct {
		name       string
		numstat    string
		status     string
		numstatErr error
		statusErr  error
		wantCount  int
		wantFirst  FileDiff
		wantErr    bool
	}{
		{
			name:      "multiple files with different statuses",
			numstat:   "10\t5\tmodified.go\n50\t0\tadded.go\n0\t25\tdeleted.go\n",
			status:    "M\tmodified.go\nA\tadded.go\nD\tdeleted.go\n",
			wantCount: 3,
			wantFirst: FileDiff{
				Path:      "modified.go",
				Status:    "M",
				Additions: 10,
				Deletions: 5,
			},
			wantErr: false,
		},
		{
			name:      "renamed file",
			numstat:   "5\t2\tnew_name.go\n",
			status:    "R100\told_name.go\tnew_name.go\n",
			wantCount: 1,
			wantFirst: FileDiff{
				Path:      "new_name.go",
				OldPath:   "old_name.go",
				Status:    "R",
				Additions: 5,
				Deletions: 2,
			},
			wantErr: false,
		},
		{
			name:      "no changes",
			numstat:   "",
			status:    "",
			wantCount: 0,
			wantErr:   false,
		},
		{
			name:       "numstat error",
			numstatErr: errors.New("fatal: bad revision"),
			wantErr:    true,
		},
		{
			name:      "status error",
			numstat:   "10\t5\tfile.go\n",
			statusErr: errors.New("fatal: bad revision"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("diff --numstat", []byte(tt.numstat), tt.numstatErr)
			mock.On("diff --name-status", []byte(tt.status), tt.statusErr)

			repo := NewWithExecutor(mock)
			diffs, err := repo.GetFileDiffs("main", "feature")

			if (err != nil) != tt.wantErr {
				t.Errorf("GetFileDiffs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(diffs) != tt.wantCount {
					t.Errorf("GetFileDiffs() returned %d files, want %d", len(diffs), tt.wantCount)
					return
				}

				if tt.wantCount > 0 {
					first := diffs[0]
					if first.Path != tt.wantFirst.Path {
						t.Errorf("GetFileDiffs()[0].Path = %q, want %q", first.Path, tt.wantFirst.Path)
					}
					if first.Status != tt.wantFirst.Status {
						t.Errorf("GetFileDiffs()[0].Status = %q, want %q", first.Status, tt.wantFirst.Status)
					}
					if first.Additions != tt.wantFirst.Additions {
						t.Errorf("GetFileDiffs()[0].Additions = %d, want %d", first.Additions, tt.wantFirst.Additions)
					}
					if first.Deletions != tt.wantFirst.Deletions {
						t.Errorf("GetFileDiffs()[0].Deletions = %d, want %d", first.Deletions, tt.wantFirst.Deletions)
					}
					if first.OldPath != tt.wantFirst.OldPath {
						t.Errorf("GetFileDiffs()[0].OldPath = %q, want %q", first.OldPath, tt.wantFirst.OldPath)
					}
				}
			}
		})
	}
}

func TestParseFileDiffs(t *testing.T) {
	tests := []struct {
		name      string
		numstat   string
		status    string
		wantCount int
	}{
		{
			name:      "empty inputs",
			numstat:   "",
			status:    "",
			wantCount: 0,
		},
		{
			name:      "single added file",
			numstat:   "100\t0\tnew_file.go\n",
			status:    "A\tnew_file.go\n",
			wantCount: 1,
		},
		{
			name:      "multiple files",
			numstat:   "10\t5\ta.go\n20\t10\tb.go\n",
			status:    "M\ta.go\nM\tb.go\n",
			wantCount: 2,
		},
		{
			name:      "copied file",
			numstat:   "0\t0\tcopy.go\n",
			status:    "C100\toriginal.go\tcopy.go\n",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs, err := parseFileDiffs(tt.numstat, tt.status)
			if err != nil {
				t.Fatalf("parseFileDiffs() unexpected error: %v", err)
			}

			if len(diffs) != tt.wantCount {
				t.Errorf("parseFileDiffs() returned %d files, want %d", len(diffs), tt.wantCount)
			}
		})
	}
}

func TestGetStagedDiff(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		err     error
		wantErr bool
	}{
		{
			name:    "has staged changes",
			output:  "diff --git a/staged.go b/staged.go\n+added line\n",
			wantErr: false,
		},
		{
			name:    "no staged changes",
			output:  "",
			wantErr: false,
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
			mock.On("diff --staged", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			diff, err := repo.GetStagedDiff()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetStagedDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && diff != tt.output {
				t.Errorf("GetStagedDiff() = %q, want %q", diff, tt.output)
			}
		})
	}
}

func TestGetUnstagedDiff(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		err     error
		wantErr bool
	}{
		{
			name:    "has unstaged changes",
			output:  "diff --git a/working.go b/working.go\n-removed line\n",
			wantErr: false,
		},
		{
			name:    "no unstaged changes",
			output:  "",
			wantErr: false,
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
			// Match just "diff" without --staged
			mock.On("git diff", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			diff, err := repo.GetUnstagedDiff()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUnstagedDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && diff != tt.output {
				t.Errorf("GetUnstagedDiff() = %q, want %q", diff, tt.output)
			}
		})
	}
}

func TestGetDiffForFile(t *testing.T) {
	fileDiff := `diff --git a/specific.go b/specific.go
index 1234567..abcdefg 100644
--- a/specific.go
+++ b/specific.go
@@ -1,3 +1,4 @@
 package main
+
+func newFunc() {}
`

	tests := []struct {
		name     string
		base     string
		head     string
		filePath string
		output   string
		err      error
		wantErr  bool
	}{
		{
			name:     "get diff for specific file",
			base:     "main",
			head:     "feature",
			filePath: "specific.go",
			output:   fileDiff,
			wantErr:  false,
		},
		{
			name:     "file not changed",
			base:     "main",
			head:     "feature",
			filePath: "unchanged.go",
			output:   "",
			wantErr:  false,
		},
		{
			name:     "error",
			base:     "main",
			head:     "feature",
			filePath: "file.go",
			err:      errors.New("fatal: bad revision"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("diff "+tt.base+".."+tt.head, []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			diff, err := repo.GetDiffForFile(tt.base, tt.head, tt.filePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetDiffForFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && diff != tt.output {
				t.Errorf("GetDiffForFile() = %q, want %q", diff, tt.output)
			}
		})
	}
}

func TestGetChangedFiles(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		err       error
		wantFiles []string
		wantErr   bool
	}{
		{
			name:      "multiple changed files",
			output:    "file1.go\nfile2.go\ndir/file3.go\n",
			wantFiles: []string{"file1.go", "file2.go", "dir/file3.go"},
			wantErr:   false,
		},
		{
			name:      "single file",
			output:    "only_file.go\n",
			wantFiles: []string{"only_file.go"},
			wantErr:   false,
		},
		{
			name:      "no changed files",
			output:    "",
			wantFiles: []string{},
			wantErr:   false,
		},
		{
			name:    "error",
			err:     errors.New("fatal: bad revision"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("diff --name-only", []byte(tt.output), tt.err)

			repo := NewWithExecutor(mock)
			files, err := repo.GetChangedFiles("main", "feature")

			if (err != nil) != tt.wantErr {
				t.Errorf("GetChangedFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(files) != len(tt.wantFiles) {
					t.Errorf("GetChangedFiles() returned %d files, want %d", len(files), len(tt.wantFiles))
					return
				}

				for i, want := range tt.wantFiles {
					if files[i] != want {
						t.Errorf("GetChangedFiles()[%d] = %q, want %q", i, files[i], want)
					}
				}
			}
		})
	}
}

func TestHasChanges(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantResult bool
	}{
		{
			name:       "has changes (diff --quiet returns error)",
			err:        errors.New("exit status 1"),
			wantResult: true,
		},
		{
			name:       "no changes (diff --quiet succeeds)",
			err:        nil,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("diff --quiet", []byte(""), tt.err)

			repo := NewWithExecutor(mock)
			result, _ := repo.HasChanges("main", "feature")

			if result != tt.wantResult {
				t.Errorf("HasChanges() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestHasStagedChanges(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantResult bool
	}{
		{
			name:       "has staged changes",
			err:        errors.New("exit status 1"),
			wantResult: true,
		},
		{
			name:       "no staged changes",
			err:        nil,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("diff --staged --quiet", []byte(""), tt.err)

			repo := NewWithExecutor(mock)
			result, _ := repo.HasStagedChanges()

			if result != tt.wantResult {
				t.Errorf("HasStagedChanges() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestHasUnstagedChanges(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantResult bool
	}{
		{
			name:       "has unstaged changes",
			err:        errors.New("exit status 1"),
			wantResult: true,
		},
		{
			name:       "no unstaged changes",
			err:        nil,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockExecutor()
			mock.On("diff --quiet", []byte(""), tt.err)

			repo := NewWithExecutor(mock)
			result, _ := repo.HasUnstagedChanges()

			if result != tt.wantResult {
				t.Errorf("HasUnstagedChanges() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestDiffStat_Fields(t *testing.T) {
	stat := DiffStat{
		FilesChanged: 5,
		Additions:    100,
		Deletions:    50,
	}

	if stat.FilesChanged != 5 {
		t.Error("FilesChanged field not set correctly")
	}
	if stat.Additions != 100 {
		t.Error("Additions field not set correctly")
	}
	if stat.Deletions != 50 {
		t.Error("Deletions field not set correctly")
	}
}

func TestFileDiff_Fields(t *testing.T) {
	diff := FileDiff{
		Path:      "new_name.go",
		OldPath:   "old_name.go",
		Status:    "R",
		Additions: 10,
		Deletions: 5,
	}

	if diff.Path != "new_name.go" {
		t.Error("Path field not set correctly")
	}
	if diff.OldPath != "old_name.go" {
		t.Error("OldPath field not set correctly")
	}
	if diff.Status != "R" {
		t.Error("Status field not set correctly")
	}
	if diff.Additions != 10 {
		t.Error("Additions field not set correctly")
	}
	if diff.Deletions != 5 {
		t.Error("Deletions field not set correctly")
	}
}
