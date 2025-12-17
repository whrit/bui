package diffview

import (
	"reflect"
	"testing"
)

// =============================================================================
// Test Data
// =============================================================================

const sampleDiff = `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,5 +1,7 @@
 package main

 func main() {
-    fmt.Println("Hello")
+    fmt.Println("Hello, World!")
+    fmt.Println("Goodbye")
 }
`

const newFileDiff = `diff --git a/newfile.go b/newfile.go
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/newfile.go
@@ -0,0 +1,5 @@
+package main
+
+func newFunc() {
+    return
+}
`

const deletedFileDiff = `diff --git a/oldfile.go b/oldfile.go
deleted file mode 100644
index 1234567..0000000
--- a/oldfile.go
+++ /dev/null
@@ -1,3 +0,0 @@
-package main
-
-func oldFunc() {}
`

const renamedFileDiff = `diff --git a/old_name.go b/new_name.go
similarity index 95%
rename from old_name.go
rename to new_name.go
index 1234567..abcdefg 100644
--- a/old_name.go
+++ b/new_name.go
@@ -1,3 +1,3 @@
 package main

-func oldName() {}
+func newName() {}
`

const multiFileDiff = `diff --git a/file1.go b/file1.go
index 1234567..abcdefg 100644
--- a/file1.go
+++ b/file1.go
@@ -1,3 +1,4 @@
 package main

+// Comment added
 func func1() {}
diff --git a/file2.go b/file2.go
index 1234567..abcdefg 100644
--- a/file2.go
+++ b/file2.go
@@ -1,3 +1,2 @@
 package main
-
 func func2() {}
`

const multiHunkDiff = `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,5 +1,5 @@
 package main

-import "fmt"
+import "log"

 func main() {
@@ -10,5 +10,6 @@
 }

 func helper() {
+    // New comment
     return
 }
`

const binaryFileDiff = `diff --git a/image.png b/image.png
new file mode 100644
index 0000000..abcdef1
Binary files /dev/null and b/image.png differ
`

const binaryFileDiffModified = `diff --git a/icon.ico b/icon.ico
index 1234567..abcdefg 100644
Binary files a/icon.ico and b/icon.ico differ
`

const gitBinaryPatchDiff = `diff --git a/data.bin b/data.bin
index 1234567..abcdefg 100644
GIT binary patch
literal 1234
somebase64encodeddata==
`

const mixedBinaryTextDiff = `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+// comment
 func main() {}
diff --git a/logo.png b/logo.png
new file mode 100644
index 0000000..abcdef1
Binary files /dev/null and b/logo.png differ
diff --git a/util.go b/util.go
index 1234567..abcdefg 100644
--- a/util.go
+++ b/util.go
@@ -1,2 +1,3 @@
 package main
+func helper() {}
`

// =============================================================================
// ParseDiff Tests
// =============================================================================

func TestParseDiff_Empty(t *testing.T) {
	files := ParseDiff("")
	if len(files) != 0 {
		t.Errorf("expected 0 files for empty input, got %d", len(files))
	}
}

func TestParseDiff_SingleFile(t *testing.T) {
	files := ParseDiff(sampleDiff)

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if f.OldPath != "main.go" {
		t.Errorf("expected OldPath 'main.go', got %q", f.OldPath)
	}
	if f.NewPath != "main.go" {
		t.Errorf("expected NewPath 'main.go', got %q", f.NewPath)
	}
	if f.Status != "modified" {
		t.Errorf("expected Status 'modified', got %q", f.Status)
	}
	if f.Additions != 2 {
		t.Errorf("expected 2 additions, got %d", f.Additions)
	}
	if f.Deletions != 1 {
		t.Errorf("expected 1 deletion, got %d", f.Deletions)
	}
}

func TestParseDiff_NewFile(t *testing.T) {
	files := ParseDiff(newFileDiff)

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if f.Status != "added" {
		t.Errorf("expected Status 'added', got %q", f.Status)
	}
	if f.OldPath != "/dev/null" {
		t.Errorf("expected OldPath '/dev/null', got %q", f.OldPath)
	}
	if f.NewPath != "newfile.go" {
		t.Errorf("expected NewPath 'newfile.go', got %q", f.NewPath)
	}
	if f.Additions != 5 {
		t.Errorf("expected 5 additions, got %d", f.Additions)
	}
	if f.Deletions != 0 {
		t.Errorf("expected 0 deletions, got %d", f.Deletions)
	}
}

func TestParseDiff_DeletedFile(t *testing.T) {
	files := ParseDiff(deletedFileDiff)

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if f.Status != "deleted" {
		t.Errorf("expected Status 'deleted', got %q", f.Status)
	}
	if f.OldPath != "oldfile.go" {
		t.Errorf("expected OldPath 'oldfile.go', got %q", f.OldPath)
	}
	if f.NewPath != "/dev/null" {
		t.Errorf("expected NewPath '/dev/null', got %q", f.NewPath)
	}
	if f.Additions != 0 {
		t.Errorf("expected 0 additions, got %d", f.Additions)
	}
	if f.Deletions != 3 {
		t.Errorf("expected 3 deletions, got %d", f.Deletions)
	}
}

func TestParseDiff_RenamedFile(t *testing.T) {
	files := ParseDiff(renamedFileDiff)

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if f.Status != "renamed" {
		t.Errorf("expected Status 'renamed', got %q", f.Status)
	}
	if f.OldPath != "old_name.go" {
		t.Errorf("expected OldPath 'old_name.go', got %q", f.OldPath)
	}
	if f.NewPath != "new_name.go" {
		t.Errorf("expected NewPath 'new_name.go', got %q", f.NewPath)
	}
}

func TestParseDiff_MultipleFiles(t *testing.T) {
	files := ParseDiff(multiFileDiff)

	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}

	// First file
	if files[0].NewPath != "file1.go" {
		t.Errorf("expected first file 'file1.go', got %q", files[0].NewPath)
	}
	if files[0].Additions != 1 {
		t.Errorf("expected 1 addition in file1, got %d", files[0].Additions)
	}

	// Second file
	if files[1].NewPath != "file2.go" {
		t.Errorf("expected second file 'file2.go', got %q", files[1].NewPath)
	}
	if files[1].Deletions != 1 {
		t.Errorf("expected 1 deletion in file2, got %d", files[1].Deletions)
	}
}

func TestParseDiff_MultipleHunks(t *testing.T) {
	files := ParseDiff(multiHunkDiff)

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if len(f.Hunks) != 2 {
		t.Fatalf("expected 2 hunks, got %d", len(f.Hunks))
	}

	// First hunk
	h1 := f.Hunks[0]
	if h1.OldStart != 1 || h1.OldCount != 5 {
		t.Errorf("hunk 1 old: expected 1,5 got %d,%d", h1.OldStart, h1.OldCount)
	}
	if h1.NewStart != 1 || h1.NewCount != 5 {
		t.Errorf("hunk 1 new: expected 1,5 got %d,%d", h1.NewStart, h1.NewCount)
	}

	// Second hunk
	h2 := f.Hunks[1]
	if h2.OldStart != 10 || h2.OldCount != 5 {
		t.Errorf("hunk 2 old: expected 10,5 got %d,%d", h2.OldStart, h2.OldCount)
	}
	if h2.NewStart != 10 || h2.NewCount != 6 {
		t.Errorf("hunk 2 new: expected 10,6 got %d,%d", h2.NewStart, h2.NewCount)
	}
}

// =============================================================================
// Hunk Parsing Tests
// =============================================================================

func TestParseHunks(t *testing.T) {
	hunks := ParseHunks(sampleDiff)

	if len(hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(hunks))
	}

	h := hunks[0]
	if h.OldStart != 1 {
		t.Errorf("expected OldStart 1, got %d", h.OldStart)
	}
	if h.OldCount != 5 {
		t.Errorf("expected OldCount 5, got %d", h.OldCount)
	}
	if h.NewStart != 1 {
		t.Errorf("expected NewStart 1, got %d", h.NewStart)
	}
	if h.NewCount != 7 {
		t.Errorf("expected NewCount 7, got %d", h.NewCount)
	}
}

func TestHunkLines(t *testing.T) {
	files := ParseDiff(sampleDiff)
	if len(files) == 0 || len(files[0].Hunks) == 0 {
		t.Fatal("expected at least one file with one hunk")
	}

	hunk := files[0].Hunks[0]

	// Count line types
	var added, removed, context, header int
	for _, line := range hunk.Lines {
		switch line.Type {
		case LineAdded:
			added++
		case LineRemoved:
			removed++
		case LineContext:
			context++
		case LineHeader:
			header++
		}
	}

	if added != 2 {
		t.Errorf("expected 2 added lines, got %d", added)
	}
	if removed != 1 {
		t.Errorf("expected 1 removed line, got %d", removed)
	}
	// Context lines include empty line at end and the space-prefixed lines
	// The sample diff has 5 context lines: " package main", "", " func main() {", " }", and the trailing empty
	if context != 5 {
		t.Errorf("expected 5 context lines, got %d", context)
	}
	if header != 1 {
		t.Errorf("expected 1 header line, got %d", header)
	}
}

func TestLineNumbers(t *testing.T) {
	files := ParseDiff(sampleDiff)
	if len(files) == 0 || len(files[0].Hunks) == 0 {
		t.Fatal("expected at least one file with one hunk")
	}

	hunk := files[0].Hunks[0]

	// Find specific lines and verify their line numbers
	for _, line := range hunk.Lines {
		switch line.Type {
		case LineContext:
			// Context lines should have both old and new line numbers
			if line.Content == " package main" {
				if line.OldNum != 1 || line.NewNum != 1 {
					t.Errorf("'package main' expected old=1, new=1, got old=%d, new=%d",
						line.OldNum, line.NewNum)
				}
			}
		case LineRemoved:
			// Removed lines should only have old line number
			if line.OldNum == 0 {
				t.Errorf("removed line should have old line number: %q", line.Content)
			}
			if line.NewNum != 0 {
				t.Errorf("removed line should not have new line number: %q", line.Content)
			}
		case LineAdded:
			// Added lines should only have new line number
			if line.NewNum == 0 {
				t.Errorf("added line should have new line number: %q", line.Content)
			}
			if line.OldNum != 0 {
				t.Errorf("added line should not have old line number: %q", line.Content)
			}
		}
	}
}

// =============================================================================
// FileDiff Tests
// =============================================================================

func TestFileDiff_DisplayPath(t *testing.T) {
	tests := []struct {
		name     string
		file     FileDiff
		expected string
	}{
		{
			name:     "modified file",
			file:     FileDiff{OldPath: "main.go", NewPath: "main.go", Status: "modified"},
			expected: "main.go",
		},
		{
			name:     "added file",
			file:     FileDiff{OldPath: "/dev/null", NewPath: "new.go", Status: "added"},
			expected: "new.go",
		},
		{
			name:     "deleted file",
			file:     FileDiff{OldPath: "old.go", NewPath: "/dev/null", Status: "deleted"},
			expected: "old.go",
		},
		{
			name:     "renamed file",
			file:     FileDiff{OldPath: "old.go", NewPath: "new.go", Status: "renamed"},
			expected: "new.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.file.DisplayPath()
			if got != tt.expected {
				t.Errorf("DisplayPath() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestFileStatusSymbol(t *testing.T) {
	tests := []struct {
		status   string
		expected string
	}{
		{"added", "A"},
		{"deleted", "D"},
		{"modified", "M"},
		{"renamed", "R"},
		{"unknown", "?"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := FileStatusSymbol(tt.status)
			if got != tt.expected {
				t.Errorf("FileStatusSymbol(%q) = %q, want %q", tt.status, got, tt.expected)
			}
		})
	}
}

func TestCountStats(t *testing.T) {
	files := []FileDiff{
		{Additions: 10, Deletions: 5},
		{Additions: 3, Deletions: 2},
		{Additions: 0, Deletions: 8},
	}

	additions, deletions := CountStats(files)

	if additions != 13 {
		t.Errorf("expected 13 additions, got %d", additions)
	}
	if deletions != 15 {
		t.Errorf("expected 15 deletions, got %d", deletions)
	}
}

func TestCountStats_Empty(t *testing.T) {
	additions, deletions := CountStats(nil)

	if additions != 0 || deletions != 0 {
		t.Errorf("expected 0,0 for nil input, got %d,%d", additions, deletions)
	}
}

func TestGetFilePath(t *testing.T) {
	tests := []struct {
		name     string
		file     FileDiff
		expected string
	}{
		{
			name:     "normal file",
			file:     FileDiff{OldPath: "old.go", NewPath: "new.go"},
			expected: "new.go",
		},
		{
			name:     "deleted file",
			file:     FileDiff{OldPath: "deleted.go", NewPath: "/dev/null"},
			expected: "deleted.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFilePath(tt.file)
			if got != tt.expected {
				t.Errorf("GetFilePath() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestLineType_String(t *testing.T) {
	tests := []struct {
		lt       LineType
		expected string
	}{
		{LineContext, "context"},
		{LineAdded, "added"},
		{LineRemoved, "removed"},
		{LineHeader, "header"},
		{LineType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.lt.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestParseDiff_NoCount(t *testing.T) {
	// Some diffs omit the count when it's 1
	diff := `diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1 +1 @@
-old
+new
`
	files := ParseDiff(diff)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if len(files[0].Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(files[0].Hunks))
	}

	h := files[0].Hunks[0]
	if h.OldCount != 1 || h.NewCount != 1 {
		t.Errorf("expected counts of 1,1 got %d,%d", h.OldCount, h.NewCount)
	}
}

func TestParseDiff_PathsWithSpaces(t *testing.T) {
	diff := `diff --git a/path with spaces/file.go b/path with spaces/file.go
--- a/path with spaces/file.go
+++ b/path with spaces/file.go
@@ -1,3 +1,3 @@
 package main
-// old
+// new
`
	files := ParseDiff(diff)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	if files[0].NewPath != "path with spaces/file.go" {
		t.Errorf("expected path with spaces, got %q", files[0].NewPath)
	}
}

// =============================================================================
// Binary File Tests
// =============================================================================

func TestParseDiff_BinaryFile(t *testing.T) {
	files := ParseDiff(binaryFileDiff)

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if !f.IsBinary {
		t.Error("expected file to be marked as binary")
	}
	if f.NewPath != "image.png" {
		t.Errorf("expected NewPath 'image.png', got %q", f.NewPath)
	}
	if len(f.Hunks) != 0 {
		t.Errorf("expected 0 hunks for binary file, got %d", len(f.Hunks))
	}
}

func TestParseDiff_BinaryFileModified(t *testing.T) {
	files := ParseDiff(binaryFileDiffModified)

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if !f.IsBinary {
		t.Error("expected file to be marked as binary")
	}
	if f.NewPath != "icon.ico" {
		t.Errorf("expected NewPath 'icon.ico', got %q", f.NewPath)
	}
}

func TestParseDiff_GitBinaryPatch(t *testing.T) {
	files := ParseDiff(gitBinaryPatchDiff)

	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}

	f := files[0]
	if !f.IsBinary {
		t.Error("expected file to be marked as binary (GIT binary patch)")
	}
}

func TestParseDiff_MixedBinaryAndText(t *testing.T) {
	files := ParseDiff(mixedBinaryTextDiff)

	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}

	// First file: main.go (text)
	if files[0].IsBinary {
		t.Error("main.go should not be binary")
	}
	if files[0].NewPath != "main.go" {
		t.Errorf("expected first file main.go, got %q", files[0].NewPath)
	}
	if len(files[0].Hunks) == 0 {
		t.Error("main.go should have hunks")
	}

	// Second file: logo.png (binary)
	if !files[1].IsBinary {
		t.Error("logo.png should be binary")
	}
	if files[1].NewPath != "logo.png" {
		t.Errorf("expected second file logo.png, got %q", files[1].NewPath)
	}

	// Third file: util.go (text)
	if files[2].IsBinary {
		t.Error("util.go should not be binary")
	}
	if files[2].NewPath != "util.go" {
		t.Errorf("expected third file util.go, got %q", files[2].NewPath)
	}
}

func TestParseDiff_BinaryFileTableDriven(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantBinary bool
		wantPath   string
	}{
		{
			name:       "new binary file",
			input:      binaryFileDiff,
			wantBinary: true,
			wantPath:   "image.png",
		},
		{
			name:       "modified binary file",
			input:      binaryFileDiffModified,
			wantBinary: true,
			wantPath:   "icon.ico",
		},
		{
			name:       "GIT binary patch format",
			input:      gitBinaryPatchDiff,
			wantBinary: true,
			wantPath:   "data.bin",
		},
		{
			name:       "normal text file",
			input:      sampleDiff,
			wantBinary: false,
			wantPath:   "main.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := ParseDiff(tt.input)
			if len(files) != 1 {
				t.Fatalf("expected 1 file, got %d", len(files))
			}

			f := files[0]
			if f.IsBinary != tt.wantBinary {
				t.Errorf("IsBinary = %v, want %v", f.IsBinary, tt.wantBinary)
			}
			if f.NewPath != tt.wantPath {
				t.Errorf("NewPath = %q, want %q", f.NewPath, tt.wantPath)
			}
		})
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkParseDiff_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseDiff(sampleDiff)
	}
}

func BenchmarkParseDiff_MultiFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseDiff(multiFileDiff)
	}
}

func TestParseIntOrDefault(t *testing.T) {
	tests := []struct {
		input    string
		def      int
		expected int
	}{
		{"", 1, 1},
		{"5", 1, 5},
		{"invalid", 1, 1},
		{"0", 1, 0},
		{"100", 1, 100},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseIntOrDefault(tt.input, tt.def)
			if got != tt.expected {
				t.Errorf("parseIntOrDefault(%q, %d) = %d, want %d",
					tt.input, tt.def, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Table-Driven Integration Test
// =============================================================================

func TestParseDiff_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedFiles int
		checkFunc     func(t *testing.T, files []FileDiff)
	}{
		{
			name:          "empty diff",
			input:         "",
			expectedFiles: 0,
			checkFunc:     nil,
		},
		{
			name:          "single modified file",
			input:         sampleDiff,
			expectedFiles: 1,
			checkFunc: func(t *testing.T, files []FileDiff) {
				if files[0].Status != "modified" {
					t.Errorf("expected modified status")
				}
			},
		},
		{
			name:          "new file",
			input:         newFileDiff,
			expectedFiles: 1,
			checkFunc: func(t *testing.T, files []FileDiff) {
				if files[0].Status != "added" {
					t.Errorf("expected added status")
				}
			},
		},
		{
			name:          "deleted file",
			input:         deletedFileDiff,
			expectedFiles: 1,
			checkFunc: func(t *testing.T, files []FileDiff) {
				if files[0].Status != "deleted" {
					t.Errorf("expected deleted status")
				}
			},
		},
		{
			name:          "renamed file",
			input:         renamedFileDiff,
			expectedFiles: 1,
			checkFunc: func(t *testing.T, files []FileDiff) {
				if files[0].Status != "renamed" {
					t.Errorf("expected renamed status")
				}
			},
		},
		{
			name:          "multiple files",
			input:         multiFileDiff,
			expectedFiles: 2,
			checkFunc: func(t *testing.T, files []FileDiff) {
				paths := []string{files[0].NewPath, files[1].NewPath}
				expected := []string{"file1.go", "file2.go"}
				if !reflect.DeepEqual(paths, expected) {
					t.Errorf("expected paths %v, got %v", expected, paths)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := ParseDiff(tt.input)
			if len(files) != tt.expectedFiles {
				t.Fatalf("expected %d files, got %d", tt.expectedFiles, len(files))
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, files)
			}
		})
	}
}
