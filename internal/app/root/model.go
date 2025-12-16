package root

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"bui/internal/config"
	"bui/internal/ui"
)

type Model struct {
	cfg     config.Config
	cfgPath string

	w, h int

	palette ui.Palette
	styles  ui.Styles
}

func New(cfg config.Config, cfgPath string) Model {
	pal := ui.NewPalette(cfg)
	styles := ui.NewStyles(pal)

	return Model{
		cfg:     cfg,
		cfgPath: cfgPath,
		palette: pal,
		styles:  styles,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
	case tea.KeyMsg:
		if msg.String() == m.cfg.Keys.Quit || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	header := lipgloss.JoinHorizontal(lipgloss.Left,
		m.styles.Badge.Render("bui"),
		" ",
		m.styles.Header.Render("PR manager + LLM copilot"),
	)

	body := m.styles.Bordered.Render(lipgloss.JoinVertical(lipgloss.Left,
		m.styles.Subtle.Render("Starter scaffold is running."),
		"",
		fmt.Sprintf("Config: %s", m.cfgPath),
		fmt.Sprintf("Theme: %s  Accent: %s  Syntax: %s", m.cfg.UI.Theme, m.cfg.UI.Accent, m.cfg.UI.SyntaxTheme),
		"",
		m.styles.Subtle.Render("Next steps: implement Dashboard → PR Detail → Diff → Composer → Create PR."),
		m.styles.Subtle.Render("This scaffold already includes config + LLM providers + diff syntax highlighting utilities."),
	))

	help := m.styles.Subtle.Render(
		m.styles.HelpKey.Render(m.cfg.Keys.Quit) + " quit  " +
			m.styles.HelpKey.Render(m.cfg.Keys.Help) + " help (todo)",
	)

	view := lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		body,
		"",
		help,
	)

	return m.styles.App.Width(m.w).Height(m.h).Padding(1, 2).Render(view)
}
