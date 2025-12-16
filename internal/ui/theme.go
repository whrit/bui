package ui

import (
	"strings"

	"bui/internal/config"

	"github.com/charmbracelet/lipgloss"
)

type Palette struct {
	Bg     lipgloss.Color
	Panel  lipgloss.Color
	Text   lipgloss.Color
	Muted  lipgloss.Color
	Border lipgloss.Color
	Accent lipgloss.Color
	Good   lipgloss.Color
	Warn   lipgloss.Color
	Bad    lipgloss.Color
}

func NewPalette(cfg config.Config) Palette {
	// Modern, calm palette tuned for terminals (works well in iTerm2/macOS Terminal).
	// Dark is default; light mode can be added by swapping these tokens.
	dark := Palette{
		Bg:     lipgloss.Color("#0B0F14"),
		Panel:  lipgloss.Color("#0F1620"),
		Text:   lipgloss.Color("#E6EDF3"),
		Muted:  lipgloss.Color("#9AA4AF"),
		Border: lipgloss.Color("#223041"),
		Good:   lipgloss.Color("#3FB950"),
		Warn:   lipgloss.Color("#D29922"),
		Bad:    lipgloss.Color("#F85149"),
	}

	accent := strings.ToLower(strings.TrimSpace(cfg.UI.Accent))
	switch accent {
	case "blue":
		dark.Accent = lipgloss.Color("#2F81F7")
	case "cyan":
		dark.Accent = lipgloss.Color("#39C5CF")
	case "teal":
		dark.Accent = lipgloss.Color("#2DD4BF")
	case "purple":
		dark.Accent = lipgloss.Color("#A371F7")
	case "pink":
		dark.Accent = lipgloss.Color("#FF7EB6")
	case "amber":
		dark.Accent = lipgloss.Color("#F59E0B")
	case "green":
		dark.Accent = lipgloss.Color("#3FB950")
	default: // indigo
		dark.Accent = lipgloss.Color("#7C5CFC")
	}

	if strings.ToLower(cfg.UI.Theme) == "light" {
		// Light palette variant (kept intentionally simple for the starter).
		return Palette{
			Bg:     lipgloss.Color("#FFFFFF"),
			Panel:  lipgloss.Color("#F6F8FA"),
			Text:   lipgloss.Color("#24292F"),
			Muted:  lipgloss.Color("#57606A"),
			Border: lipgloss.Color("#D0D7DE"),
			Accent: dark.Accent,
			Good:   lipgloss.Color("#1A7F37"),
			Warn:   lipgloss.Color("#9A6700"),
			Bad:    lipgloss.Color("#D1242F"),
		}
	}

	return dark
}

type Styles struct {
	App      lipgloss.Style
	Header   lipgloss.Style
	Subtle   lipgloss.Style
	Panel    lipgloss.Style
	Bordered lipgloss.Style
	Badge    lipgloss.Style
	HelpKey  lipgloss.Style
}

func NewStyles(p Palette) Styles {
	return Styles{
		App: lipgloss.NewStyle().
			Background(p.Bg).
			Foreground(p.Text),

		Header: lipgloss.NewStyle().
			Foreground(p.Text).
			Bold(true),

		Subtle: lipgloss.NewStyle().
			Foreground(p.Muted),

		Panel: lipgloss.NewStyle().
			Background(p.Panel).
			Foreground(p.Text).
			Padding(1, 2),

		Bordered: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(p.Border).
			Background(p.Panel).
			Foreground(p.Text).
			Padding(1, 2),

		Badge: lipgloss.NewStyle().
			Foreground(p.Bg).
			Background(p.Accent).
			Padding(0, 1).
			Bold(true),

		HelpKey: lipgloss.NewStyle().
			Foreground(p.Accent).
			Bold(true),
	}
}
