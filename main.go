package main

import (
	"fmt"
	"os"
	"ruffnut/ui/screens"
	"ruffnut/ui/theme"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ruffnut: %v\n", err)
		os.Exit(1)
	}
}

type model struct {
	quitting bool
	maxH int
	maxW int
	appH int
	appW int
	termH int
	termW int
	styles theme.Styles
}

func initialModel() model {
	return model{
		maxW: 100,
		maxH: 30,
		appW: 100,
		appH: 30,
		styles: theme.NewStyles(theme.Default(), 100, 30),

	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.termW = msg.Width
		m.termH = msg.Height
		m.appW = min(m.termW, m.maxW)
		m.appH = min(m.termH, m.maxH)
		m.styles = theme.NewStyles(theme.Default(), m.appH, m.appW)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	if (m.quitting) {
		return tea.NewView("bye bye")
	}
	box := m.styles.Box.Render(screens.HomeView(m.styles))
	v := tea.NewView(lipgloss.Place(m.termW, m.termH, lipgloss.Center, lipgloss.Center, box))
	v.AltScreen = true
	return v
}
