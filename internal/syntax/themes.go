package syntax

import (
	"strings"

	"github.com/alecthomas/chroma/v2/styles"
)

func ResolveTheme(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return "github-dark"
	}
	if styles.Get(n) != nil {
		return n
	}
	// friendly aliases
	switch n {
	case "github", "gh":
		return "github"
	case "github-dark", "gh-dark":
		return "github-dark"
	case "dracula":
		return "dracula"
	}
	return "github-dark"
}
