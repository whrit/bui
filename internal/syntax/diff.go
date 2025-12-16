package syntax

import (
	"bytes"
	"strings"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// HighlightDiff is a pragmatic v1 highlighter:
// - Adds/removals are colored via diff prefixes
// - Code lines can be syntax-highlighted if a file language is known
//
// For v1 scaffolding, we do *not* fully parse unified diff into hunks/files yet.
// Instead, this function accepts an optional lexer name; if empty, it only applies diff coloring.
func HighlightDiff(raw string, themeName string, lexerName string) (string, error) {
	style := styles.Get(ResolveTheme(themeName))
	if style == nil {
		style = styles.Fallback
	}

	// Terminal formatter with ANSI sequences.
	f := formatters.Get("terminal16m")
	if f == nil {
		f = formatters.Fallback
	}

	lexer := lexers.Get(lexerName)
	// If we don't know the language, we'll only do lightweight diff prefix coloring.
	if lexer == nil {
		return colorDiffPrefixes(raw), nil
	}

	// We apply highlighting only on "content lines" (not headers like @@, diff --git, +++, ---).
	// Keep this simple for scaffolding; refine later with real diff parsing.
	var out bytes.Buffer
	lines := strings.Split(raw, "\n")
	for i, line := range lines {
		if isDiffHeader(line) {
			out.WriteString(colorDiffPrefixes(line))
		} else if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, " ") {
			// Strip the prefix for lexing, then prepend the prefix back.
			prefix := line[:1]
			code := line[1:]
			it, err := lexer.Tokenise(nil, code)
			if err != nil {
				out.WriteString(colorDiffPrefixes(line))
			} else {
				var tmp bytes.Buffer
				_ = f.Format(&tmp, style, it)
				out.WriteString(colorPrefix(prefix))
				out.WriteString(strings.TrimRight(tmp.String(), "\n"))
			}
		} else {
			out.WriteString(line)
		}
		if i < len(lines)-1 {
			out.WriteString("\n")
		}
	}
	return out.String(), nil
}

func isDiffHeader(line string) bool {
	return strings.HasPrefix(line, "diff --git ") ||
		strings.HasPrefix(line, "index ") ||
		strings.HasPrefix(line, "--- ") ||
		strings.HasPrefix(line, "+++ ") ||
		strings.HasPrefix(line, "@@")
}

func colorDiffPrefixes(s string) string {
	// Minimal coloring using ANSI:
	// green for additions, red for removals, dim for headers.
	if strings.HasPrefix(s, "+") && !strings.HasPrefix(s, "+++ ") {
		return "\x1b[38;2;63;185;80m" + s + "\x1b[0m"
	}
	if strings.HasPrefix(s, "-") && !strings.HasPrefix(s, "--- ") {
		return "\x1b[38;2;248;81;73m" + s + "\x1b[0m"
	}
	if isDiffHeader(s) || strings.HasPrefix(s, "diff ") {
		return "\x1b[38;2;154;164;175m" + s + "\x1b[0m"
	}
	return s
}

func colorPrefix(prefix string) string {
	switch prefix {
	case "+":
		return "\x1b[38;2;63;185;80m+\x1b[0m"
	case "-":
		return "\x1b[38;2;248;81;73m-\x1b[0m"
	case " ":
		return " "
	default:
		return prefix
	}
}
