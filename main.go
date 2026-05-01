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
	ti.SetWidth(utils.DefaultStruct.W)
	ti.CharLimit = 128
	ti.Focus()
	s := ti.Styles()
	s.Focused.Placeholder = s.Focused.Placeholder.Foreground(lipgloss.Color("#0000ff"))
	ti.SetStyles(s)
	return model{
		maxW:   utils.DefaultStruct.MaxW,
		maxH:   utils.DefaultStruct.MaxH,
		appW:   utils.DefaultStruct.W,
		appH:   utils.DefaultStruct.H,
		styles: theme.NewStyles(theme.Default(), utils.DefaultStruct.MaxW, utils.DefaultStruct.MaxH),
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
		m.appH = min(m.termH, m.maxH)
		m.styles = theme.NewStyles(theme.Default(), m.appW, m.appH)
		m.input.SetWidth(m.appW)
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
	raw := m.input.View()
	boxed := m.styles.Input.Render(raw)
	join := lipgloss.JoinVertical(lipgloss.Center,screens.HomeView(m.styles), boxed)
	box := m.styles.Box.Render(join)
	v := tea.NewView(
		utils.CenterPlace(m.appW, m.appH, box),
	)
	v.AltScreen = true
	return v
} 


