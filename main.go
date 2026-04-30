package main

import (
	"fmt"
	"os"
	"ruffnut/ui/screens"
	tea "charm.land/bubbletea/v2"
)

func main() {
	if _, err := tea.NewProgram(initialModel()).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ruffnut: %v\n", err)
		os.Exit(1)
	}
}

type model struct {
	quitting bool
}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyPressMsg); ok {
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("bye\n")
	}
	homeView := screens.HomeView()
	return homeView
}
