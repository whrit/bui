package syntax

import (
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2/lexers"
)

func LexerForFile(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return ""
	}
	l := lexers.Match(path)
	if l == nil {
		// Try by extension without dot
		l = lexers.Get(strings.TrimPrefix(ext, "."))
	}
	if l == nil {
		return ""
	}
	return l.Config().Name
}
