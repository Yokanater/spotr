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
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func main() {
	st, err := store.NewSQLite("ruffnut.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "spotr: open db: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	if _, err := tea.NewProgram(initialModel(st)).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "spotr: %v\n", err)
		os.Exit(1)
	}
}

type mode string
type screen string
type inputPurpose string

const (
	modeNormal mode = "normal"
	modeInput  mode = "input"
	modeCmd    mode = "command"
)

const (
	inputNone        inputPurpose = ""
	inputAddProgram  inputPurpose = "add_program"
	inputAddWorkout  inputPurpose = "add_workout"
	inputAddExercise inputPurpose = "add_exercise"
)

const (
	screenPrograms  screen = "programs"
	screenWorkouts  screen = "workouts"
	screenExercises screen = "exercises"
	screenHelp      screen = "help"
)

type model struct {
	quitting       bool
	maxH           int
	maxW           int
	appH           int
	appW           int
	termH          int
	termW          int
	theme          theme.Theme
	screen         screen
	mode           mode
	inputPurpose   inputPurpose
	styles         theme.Styles
	input          textinput.Model
	store          *store.Store
	status         string
	programs       []data.Program
	workouts       []data.Workout
	exercises      []data.Exercise
	activeProgram  data.Program
	activeWorkout  data.Workout
	activeExercise data.Exercise
}

func initialModel(st *store.Store) model {
	ti := textinput.New()
	t := theme.Default()
	ti.Placeholder = "press : to type a command"
	ti.Prompt = ""
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
		mode:   modeNormal,
		screen: "home",
		status: "press : for commands, a to add, ? for help",
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
		m.appW = min(m.termW, min(m.maxW, m.theme.InputMax))
		m.appH = min(m.termH, m.maxH)
		m.styles = theme.NewStyles(m.theme, m.appW, m.appH)
		m.input.SetWidth(max(1, min(m.theme.InputMax, m.appW-m.theme.PadX)))
	case tea.KeyPressMsg:
		switch m.mode {
		case modeCmd:
			return m.handleCommandKey(msg)

		case modeInput:
			return m.handleInputKey(msg)

		case modeNormal:
			return m.handleNormalKey(msg)
		}
	}
	return m, cmd
}

func (m model) handleCommandKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		line := m.input.Value()
		m.input.SetValue("")
		m.mode = modeNormal
		m.inputPurpose = inputNone
		m.resetInputPrompt()
		if line == "" {
			m.status = "command cancelled"
			return m, cmd
		}
		return m.runCommandLine(line)
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.input.SetValue("")
		m.mode = modeNormal
		m.inputPurpose = inputNone
		m.resetInputPrompt()
		m.status = "command cancelled"
		return m, cmd
	default:
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m model) handleInputKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "enter":
		value := strings.TrimSpace(m.input.Value())
		m.input.SetValue("")
		purpose := m.inputPurpose
		m.mode = modeNormal
		m.inputPurpose = inputNone
		m.resetInputPrompt()
		if value == "" {
			m.status = "add cancelled"
			return m, cmd
		}
		m.submitInput(purpose, value)
		return m, cmd
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.input.SetValue("")
		m.mode = modeNormal
		m.inputPurpose = inputNone
		m.resetInputPrompt()
		m.status = "add cancelled"
		return m, cmd
	default:
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m model) handleNormalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case ":":
		m.mode = modeCmd
		m.inputPurpose = inputNone
		m.input.SetValue("")
		m.input.Placeholder = "program list"
		m.input.Prompt = "spotr $ "
		m.status = "command mode"
	case "a":
		m.startAdd()
	case "?":
		m.screen = "help"
		m.status = "help"
	case "home", "b", "esc":
		m.screen = "home"
		m.status = "home"
	case "q", "ctrl+c":
		return m, tea.Quit
	default:
		m.status = "press : for commands, a to add, ? for help"
	}
	return m, cmd
}

func (m *model) startAdd() {
	m.mode = modeInput
	m.input.SetValue("")
	m.screen = "program"

	switch {
	case m.activeWorkout.WorkoutId != 0:
		m.inputPurpose = inputAddExercise
		m.input.Placeholder = "exercise name [sets] [reps]"
		m.input.Prompt = "add exercise $ "
		m.status = "adding exercise"
	case m.activeProgram.ProgramId != 0:
		m.inputPurpose = inputAddWorkout
		m.input.Placeholder = "workout name"
		m.input.Prompt = "add workout $ "
		m.status = "adding workout"
	default:
		m.inputPurpose = inputAddProgram
		m.input.Placeholder = "program name"
		m.input.Prompt = "add program $ "
		m.status = "adding program"
	}
}

func (m *model) submitInput(purpose inputPurpose, value string) {
	args := strings.Fields(value)
	switch purpose {
	case inputAddProgram:
		m.handleProgram(append([]string{"add"}, args...))
	case inputAddWorkout:
		m.handleWorkout(append([]string{"add"}, args...))
	case inputAddExercise:
		m.handleExercise(append([]string{"add"}, args...))
	default:
		m.status = "nothing to submit"
	}
}

func (m *model) resetInputPrompt() {
	m.input.Placeholder = "press : to type a command"
	m.input.Prompt = ""
}

func (m model) runCommandLine(line string) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
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
		screen = screens.ProgramView(m.styles, m.programs, m.workouts, m.exercises, m.activeProgram, m.activeWorkout, m.activeExercise)

	}
	screenHeight := max(1, m.appH-lipgloss.Height(input)-lipgloss.Height(status)-2)
	screen = lipgloss.NewStyle().
		Width(m.styles.Box.GetWidth()).
		MaxHeight(screenHeight).
		Render(screen)
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

		name := strings.Join(args[1:], " ")
		id, err := m.store.CreateProgram(name)
		if err != nil {
			m.status = err.Error()
			return
		}
		program := data.Program{ProgramId: id, ProgramName: name}
		m.programs = append(m.programs, program)
		m.status = "Created program"

	case "select":
		if len(args) < 2 {
			m.status = "usage: program select <id|name>"
			return
		}

		name := strings.Join(args[1:], " ")
		program, err := m.store.SelectProgram(name)
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
		m.activeExercise = data.Exercise{}
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

		name := strings.Join(args[1:], " ")
		err := m.store.CreateWorkout(name, m.activeProgram)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.workouts = append(m.workouts, data.Workout{
			ProgramId: m.activeProgram.ProgramId,
			Name:      name,
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

		name := strings.Join(args[1:], " ")
		workout, err := m.store.SelectWorkout(name, m.activeProgram)
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
		m.activeExercise = data.Exercise{}
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
			m.status = "usage: exercise add <name> [sets] [reps]"
			return
		}

		name, sets, reps, err := parseExerciseAddArgs(args)
		if err != nil {
			m.status = err.Error()
			return
		}

		err = m.store.CreateExercise(name, sets, reps, m.activeWorkout)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.exercises = append(m.exercises, data.Exercise{
			WorkoutId: m.activeWorkout.WorkoutId,
			Name:      name,
			Sets:      sets,
			Reps:      reps,
		})
		m.status = "Created exercise"

	case "list":
		exercises, err := m.store.ListExercises(m.activeWorkout)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.exercises = exercises

	case "select":
		if len(args) < 2 {
			m.status = "usage: exercise select <id|name>"
			return
		}

		name := strings.Join(args[1:], " ")
		exercise, err := m.store.SelectExercise(name, m.activeWorkout)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.activeExercise = exercise
		m.status = "Selected exercise" + exercise.Name

	case "set":
		if m.activeExercise.ExerciseId == 0 {
			m.status = "select an exercise first: exercise select <id|name>"
			return
		}
		if len(args) < 3 {
			m.status = "usage: exercise set <sets> <reps>"
			return
		}

		sets, reps, err := parseSetReps(args[1], args[2])
		if err != nil {
			m.status = err.Error()
			return
		}
		err = m.store.UpdateExerciseDefaults(m.activeExercise, sets, reps)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.activeExercise.Sets = sets
		m.activeExercise.Reps = reps
		for i := range m.exercises {
			if m.exercises[i].ExerciseId == m.activeExercise.ExerciseId {
				m.exercises[i].Sets = sets
				m.exercises[i].Reps = reps
				break
			}
		}
		m.status = "Updated exercise"

	default:
		m.status = fmt.Sprintf("unknown exercise command: %s", cmd)
	}
}

func parseSetReps(setsArg string, repsArg string) (int, int, error) {
	sets, err := strconv.Atoi(setsArg)
	if err != nil {
		return 0, 0, fmt.Errorf("sets must be a number")
	}
	reps, err := strconv.Atoi(repsArg)
	if err != nil {
		return 0, 0, fmt.Errorf("reps must be a number")
	}
	return sets, reps, nil
}

func parseExerciseAddArgs(args []string) (string, int, int, error) {
	nameEnd := len(args)
	sets := 0
	reps := 0

	if len(args) >= 4 {
		parsedSets, err := strconv.Atoi(args[len(args)-2])
		if err != nil {
			return "", 0, 0, fmt.Errorf("sets must be a number")
		}
		parsedReps, err := strconv.Atoi(args[len(args)-1])
		if err != nil {
			return "", 0, 0, fmt.Errorf("reps must be a number")
		}
		nameEnd = len(args) - 2
		sets = parsedSets
		reps = parsedReps
	}

	name := strings.Join(args[1:nameEnd], " ")
	if strings.TrimSpace(name) == "" {
		return "", 0, 0, fmt.Errorf("exercise name is required")
	}

	return name, sets, reps, nil
}
