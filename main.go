package main

import (
	"fmt"
	"os"
	"ruffnut/ui/screens"
	"ruffnut/ui/theme"
	"ruffnut/ui/utils"

	"charm.land/bubbles/v2/textinput"
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
	maxH     int
	maxW     int
	appH     int
	appW     int
	termH    int
	termW    int
	styles   theme.Styles
	input textinput.Model
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Type something..."
	ti.SetWidth(100 - 2)
	ti.CharLimit = 128
	ti.Focus()
	return model{
		maxW:   1000,
		maxH:   300,
		appW:   100,
		appH:   30,
		styles: theme.NewStyles(theme.Default(), 100, 30),
		input: ti,
		
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termW = msg.Width
		m.termH = msg.Height
		m.appW = min(m.termW - 1, m.maxW)
		m.appH = min(m.termH - 1, m.maxH)
		m.styles = theme.NewStyles(theme.Default(), m.appH, m.appW)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		}
	}
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("bye bye")
	}
	join := lipgloss.JoinVertical(lipgloss.Center,screens.HomeView(m.styles), m.input.View())
	box := m.styles.Box.Render(join)
	
	v := tea.NewView(
		utils.CenterPlace(m.appW, m.appH, box),
	)
	v.AltScreen = true
	return v
} 


