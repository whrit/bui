// Package diffview implements the diff viewer screen for the bui application.
// It displays a split-panel view with a file tree and syntax-highlighted diff content.
package diffview

import (
	"regexp"
	"strconv"
	"strings"
)

// =============================================================================
// Types
// =============================================================================

// LineType represents the type of a diff line.
type LineType int

const (
	// LineContext represents an unchanged line.
	LineContext LineType = iota
	// LineAdded represents an added line.
	LineAdded
	// LineRemoved represents a removed line.
	LineRemoved
	// LineHeader represents a hunk header or diff header line.
	LineHeader
)

// String returns a string representation of the line type.
func (lt LineType) String() string {
	switch lt {
	case LineContext:
		return "context"
	case LineAdded:
		return "added"
	case LineRemoved:
		return "removed"
	case LineHeader:
		return "header"
	default:
		return "unknown"
	}
}

// DiffLine represents a single line in a diff.
type DiffLine struct {
	Type    LineType // Type of the line (Added, Removed, Context, Header)
	Content string   // The line content including prefix
	OldNum  int      // Line number in old file (0 if not applicable)
	NewNum  int      // Line number in new file (0 if not applicable)
}

// Hunk represents a diff hunk.
type Hunk struct {
	OldStart int        // Starting line number in old file
	OldCount int        // Number of lines from old file
	NewStart int        // Starting line number in new file
	NewCount int        // Number of lines in new file
	Header   string     // The @@ header line
	Lines    []DiffLine // Lines in this hunk
}

// FileDiff represents a single file's diff.
type FileDiff struct {
	OldPath   string // Path in the old version (may be /dev/null for new files)
	NewPath   string // Path in the new version (may be /dev/null for deleted files)
	Status    string // "added", "deleted", "modified", "renamed"
	Hunks     []Hunk // Parsed hunks
	Additions int    // Total lines added
	Deletions int    // Total lines deleted
}

// DisplayPath returns the most appropriate path to display for this file.
func (f FileDiff) DisplayPath() string {
	switch f.Status {
	case "added":
		return f.NewPath
	case "deleted":
		return f.OldPath
	default:
		return f.NewPath
	}
}

// =============================================================================
// Parsing
// =============================================================================

// Regular expressions for parsing unified diffs.
var (
	// diffHeaderRE matches "diff --git a/path b/path" lines.
	diffHeaderRE = regexp.MustCompile(`^diff --git a/(.*?) b/(.*)$`)

	// hunkHeaderRE matches "@@ -old,count +new,count @@" lines.
	// Supports optional count (defaults to 1 if omitted).
	hunkHeaderRE = regexp.MustCompile(`^@@\s+-(\d+)(?:,(\d+))?\s+\+(\d+)(?:,(\d+))?\s+@@(.*)$`)

	// oldFileRE matches "--- a/path" or "--- /dev/null".
	oldFileRE = regexp.MustCompile(`^---\s+(?:a/)?(.*)$`)

	// newFileRE matches "+++ b/path" or "+++ /dev/null".
	newFileRE = regexp.MustCompile(`^[+][+][+]\s+(?:b/)?(.*)$`)
)

// ParseDiff parses a unified diff string into structured FileDiffs.
func ParseDiff(raw string) []FileDiff {
	if raw == "" {
		return nil
	}

	lines := strings.Split(raw, "\n")
	var files []FileDiff
	var currentFile *FileDiff
	var currentHunk *Hunk
	var oldLineNum, newLineNum int

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Check for diff header (start of new file)
		if matches := diffHeaderRE.FindStringSubmatch(line); matches != nil {
			// Save previous file if exists
			if currentFile != nil {
				if currentHunk != nil {
					currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
				}
				files = append(files, *currentFile)
			}

			currentFile = &FileDiff{
				OldPath: matches[1],
				NewPath: matches[2],
				Status:  "modified", // Default, will be updated based on paths
			}
			currentHunk = nil
			continue
		}

		// Check for old file path
		if matches := oldFileRE.FindStringSubmatch(line); matches != nil && currentFile != nil {
			path := matches[1]
			if path == "/dev/null" {
				currentFile.OldPath = "/dev/null"
				currentFile.Status = "added"
			} else if currentFile.OldPath == "" {
				currentFile.OldPath = path
			}
			continue
		}

		// Check for new file path
		if matches := newFileRE.FindStringSubmatch(line); matches != nil && currentFile != nil {
			path := matches[1]
			if path == "/dev/null" {
				currentFile.NewPath = "/dev/null"
				currentFile.Status = "deleted"
			} else if currentFile.NewPath == "" {
				currentFile.NewPath = path
			}
			continue
		}

		// Check for hunk header
		if matches := hunkHeaderRE.FindStringSubmatch(line); matches != nil && currentFile != nil {
			// Save previous hunk if exists
			if currentHunk != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			}

			oldStart := parseIntOrDefault(matches[1], 1)
			oldCount := parseIntOrDefault(matches[2], 1)
			newStart := parseIntOrDefault(matches[3], 1)
			newCount := parseIntOrDefault(matches[4], 1)

			currentHunk = &Hunk{
				OldStart: oldStart,
				OldCount: oldCount,
				NewStart: newStart,
				NewCount: newCount,
				Header:   line,
			}

			// Initialize line numbers
			oldLineNum = oldStart
			newLineNum = newStart

			// Add the header as a line
			currentHunk.Lines = append(currentHunk.Lines, DiffLine{
				Type:    LineHeader,
				Content: line,
			})
			continue
		}

		// Parse content lines within a hunk
		if currentHunk != nil && len(line) > 0 {
			diffLine := DiffLine{Content: line}

			switch line[0] {
			case '+':
				diffLine.Type = LineAdded
				diffLine.NewNum = newLineNum
				newLineNum++
				currentFile.Additions++
			case '-':
				diffLine.Type = LineRemoved
				diffLine.OldNum = oldLineNum
				oldLineNum++
				currentFile.Deletions++
			case ' ':
				diffLine.Type = LineContext
				diffLine.OldNum = oldLineNum
				diffLine.NewNum = newLineNum
				oldLineNum++
				newLineNum++
			default:
				// Could be a "\ No newline at end of file" or other metadata
				diffLine.Type = LineHeader
			}

			currentHunk.Lines = append(currentHunk.Lines, diffLine)
		} else if currentHunk != nil && len(line) == 0 {
			// Empty line in diff (context line that is blank)
			diffLine := DiffLine{
				Type:    LineContext,
				Content: line,
				OldNum:  oldLineNum,
				NewNum:  newLineNum,
			}
			oldLineNum++
			newLineNum++
			currentHunk.Lines = append(currentHunk.Lines, diffLine)
		}
	}

	// Save the last file and hunk
	if currentFile != nil {
		if currentHunk != nil {
			currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
		}
		files = append(files, *currentFile)
	}

	// Detect renamed files
	for i := range files {
		if files[i].OldPath != files[i].NewPath &&
			files[i].Status == "modified" &&
			files[i].OldPath != "/dev/null" &&
			files[i].NewPath != "/dev/null" {
			files[i].Status = "renamed"
		}
	}

	return files
}

// ParseHunks parses hunks from a file diff section.
// This is a convenience function for parsing just the hunk portion of a diff.
func ParseHunks(content string) []Hunk {
	files := ParseDiff(content)
	if len(files) == 0 {
		return nil
	}
	return files[0].Hunks
}

// parseIntOrDefault parses a string to int, returning defaultVal if empty or invalid.
func parseIntOrDefault(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

// =============================================================================
// Helper Functions
// =============================================================================

// FileStatusSymbol returns a single character symbol for the file status.
func FileStatusSymbol(status string) string {
	switch status {
	case "added":
		return "A"
	case "deleted":
		return "D"
	case "modified":
		return "M"
	case "renamed":
		return "R"
	default:
		return "?"
	}
}

// CountStats returns total additions and deletions across all files.
func CountStats(files []FileDiff) (additions, deletions int) {
	for _, f := range files {
		additions += f.Additions
		deletions += f.Deletions
	}
	return
}

// GetFilePath extracts the file path from a diff, handling /dev/null cases.
func GetFilePath(f FileDiff) string {
	if f.NewPath == "/dev/null" {
		return f.OldPath
	}
	return f.NewPath
}
