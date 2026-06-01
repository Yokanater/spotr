package main

import (
	"fmt"
	"os"
	"ruffnut/commands"
	"ruffnut/data"
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
	quitting      bool
	maxH          int
	maxW          int
	appH          int
	appW          int
	termH         int
	termW         int
	theme         theme.Theme
	screen        string
	styles        theme.Styles
	input         textinput.Model
	store         *store.Store
	status        string
	programs      []data.Program
	workouts      []data.Workout
	exercises     []data.Exercise
	activeProgram data.Program
	activeWorkout data.Workout
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
		screen: "home",
		status: "hello everything good and great rn",
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
		case "enter":
			{
				line := m.input.Value()
				m.input.SetValue("")
				if line == "" {
					break
				}
				command, ok := commands.Parse(line)
				if !ok {
					m.status = "error parsing command"
					return m, cmd
				}
				resolved, status := commands.Resolve(command)

				if !status {
					m.status = fmt.Sprintf("Command not defined: %v", resolved)
					return m, cmd
				}
				switch resolved {
				case "program":
					m.handleProgram(command.Args)

				case "workout":
					m.handleWorkout(command.Args)
					return m, cmd
				case "exercise":
					m.handleExercise(command.Args)
					return m, cmd
				case "help":
					m.screen = "help"
					return m, cmd

				case "home":
					m.screen = "home"
					return m, cmd

				case "quit":
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

	rawInput := m.input.View()
	input := m.styles.Input.Render(rawInput)
	status := m.styles.Status.Render(m.status)
	screen := ""
	switch m.screen {
	case "home":
		screen = screens.HomeView(m.styles)

	case "help":
		screen = screens.HelpView(m.styles)

	case "program":
		screen = screens.ProgramView(m.styles, m.programs, m.workouts, m.exercises, m.activeProgram, m.activeWorkout)

	}
	join := lipgloss.JoinVertical(lipgloss.Center, screen, input, status)
	box := m.styles.Box.Render(join)
	v := tea.NewView(
		utils.CenterPlace(m.termW, m.termH, box),
	)
	v.BackgroundColor = m.theme.Background
	v.AltScreen = true
	return v
}

func (m *model) handleProgram(args []string) {
	if len(args) == 0 {
		m.status = "usage: program <list|add|select> ..."
		return
	}

	cmd := args[0]
	m.screen = "program"
	switch cmd {
	case "list":
		programs, err := m.store.ListPrograms()
		if err != nil {
			m.status = err.Error()
			return
		}
		m.programs = programs

	case "add":
		if len(args) < 2 {
			m.status = "usage: program add <name>"
			return
		}

		id, err := m.store.CreateProgram(args[1])
		if err != nil {
			m.status = err.Error()
			return
		}
		program := data.Program{ProgramId: id, ProgramName: args[1]}
		m.programs = append(m.programs, program)
		m.status = "Created program"

	case "select":
		if len(args) < 2 {
			m.status = "usage: program select <id|name>"
			return
		}

		program, err := m.store.SelectProgram(args[1])
		if err != nil {
			m.status = err.Error()
			return
		}
		m.activeProgram = program
		workouts, err := m.store.ListWorkouts(program)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.workouts = workouts
		m.activeWorkout = data.Workout{}
		m.exercises = nil
		m.status = "Selected program" + program.ProgramName

	default:
		m.status = fmt.Sprintf("unknown program command: %s", cmd)
	}
}

func (m *model) handleWorkout(args []string) {
	if m.activeProgram.ProgramId == 0 {
		m.status = "select a program first: program select <id|name>"
		return
	}

	if len(args) == 0 {
		m.status = "usage: workout <list|add> ..."
		return
	}
	cmd := args[0]
	m.screen = "program"
	switch cmd {
	case "add":

		if len(args) < 2 {
			m.status = "usage: workout add <name>"
			return
		}

		err := m.store.CreateWorkout(args[1], m.activeProgram)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.workouts = append(m.workouts, data.Workout{
			ProgramId: m.activeProgram.ProgramId,
			Name:      args[1],
		})
		m.status = "Created workout"

	case "list":
		workouts, err := m.store.ListWorkouts(m.activeProgram)

		if err != nil {
			m.status = err.Error()
			return
		}
		m.workouts = workouts

	case "select":
		if len(args) < 2 {
			m.status = "usage: workout select <id|name>"
			return
		}

		workout, err := m.store.SelectWorkout(args[1], m.activeProgram)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.activeWorkout = workout
		exercises, err := m.store.ListExercises(workout)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.exercises = exercises
		m.status = "Selected workout" + workout.Name
	}
}

func (m *model) handleExercise(args []string) {
	if m.activeWorkout.WorkoutId == 0 {
		m.status = "select a workout first: workout select <id|name>"
		return
	}

	if len(args) == 0 {
		m.status = "usage: exercise <list|add> ..."
		return
	}

	cmd := args[0]
	m.screen = "program"
	switch cmd {
	case "add":
		if len(args) < 2 {
			m.status = "usage: exercise add <name>"
			return
		}

		err := m.store.CreateExercise(args[1], m.activeWorkout)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.exercises = append(m.exercises, data.Exercise{
			WorkoutId: m.activeWorkout.WorkoutId,
			Name:      args[1],
		})
		m.status = "Created exercise"

	case "list":
		exercises, err := m.store.ListExercises(m.activeWorkout)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.exercises = exercises

	default:
		m.status = fmt.Sprintf("unknown exercise command: %s", cmd)
	}
}
