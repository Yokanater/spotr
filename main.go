package main

import (
	"fmt"
	"os"
	"ruffnut/commands"
	"ruffnut/store"
	"ruffnut/ui/screens"
	"ruffnut/ui/theme"
	"ruffnut/ui/utils"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func main() {
	st, err := store.NewSQLite("ruffnut.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ruffnut: open db: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	if _, err := tea.NewProgram(initialModel(st)).Run(); err != nil {
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
	theme    theme.Theme
	styles   theme.Styles
	input    textinput.Model
	store    *store.Store
	status string
}

func initialModel(st *store.Store) model {
	ti := textinput.New()
	ti.Placeholder = "Type something..."
	t := theme.Default()
	ti.SetWidth(t.InputMax)
	ti.CharLimit = 128
	ti.Focus()
	return model{
		maxW:   utils.DefaultStruct.MaxW,
		maxH:   utils.DefaultStruct.MaxH,
		appW:   utils.DefaultStruct.W,
		appH:   utils.DefaultStruct.H,
		theme:  t,
		styles: theme.NewStyles(t, utils.DefaultStruct.MaxW, utils.DefaultStruct.MaxH),
		input:  ti,
		store:  st,
		status: "",
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
		m.appW = min(m.termW, m.maxW)
		m.appH = min(m.termH, m.maxH)
		m.styles = theme.NewStyles(m.theme, m.appW, m.appH)
		m.input.SetWidth(min(m.theme.InputMax, m.appW-m.theme.PadX))
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter": {
			line := m.input.Value()
			m.input.SetValue("")
			if line == "" {
				break
			}
			command, ok := commands.Parse(line)
			if (!ok) {
				m.status = "error parsing command"
				return m, cmd
			}
			resolved, status := commands.Resolve(command)
			
			if (!status){
				m.status = fmt.Sprintf("Command not defined: %v", resolved)
				return m, cmd
			}

			if resolved == "quit" {
				return m, tea.Quit
			}
		}
		case "ctrl+c", "esc":
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
	join := lipgloss.JoinVertical(lipgloss.Center, screens.HomeView(m.styles), boxed)
	box := m.styles.Box.Render(join)
	v := tea.NewView(
		utils.CenterPlace(m.termW, m.termH, box),
	)
	v.BackgroundColor = m.theme.Background
	v.AltScreen = true
	return v
}
