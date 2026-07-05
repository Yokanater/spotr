package main

import (
	"database/sql"
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
	modeQuit   mode = "quit"
)

const (
	inputNone        inputPurpose = ""
	inputAddProgram  inputPurpose = "add_program"
	inputAddWorkout  inputPurpose = "add_workout"
	inputAddExercise inputPurpose = "add_exercise"
)

const (
	screenHome      screen = "home"
	screenProgram   screen = "program"
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
	programCursor  int
	workoutCursor  int
	exerciseCursor int
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
	m := model{
		maxW:   utils.DefaultStruct.MaxW,
		maxH:   utils.DefaultStruct.MaxH,
		appW:   utils.DefaultStruct.W,
		appH:   utils.DefaultStruct.H,
		theme:  t,
		styles: theme.NewStyles(t, utils.DefaultStruct.MaxW, utils.DefaultStruct.MaxH),
		input:  ti,
		store:  st,
		mode:   modeNormal,
		screen: screenHome,
		status: "press : for commands, a to add, ? for help",
	}
	if err := m.loadPrograms(); err != nil {
		m.status = err.Error()
	}
	return m
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
		case modeQuit:
			return m.handleQuitKey(msg)

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
		m.requestQuit()
		return m, cmd
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
		m.requestQuit()
		return m, cmd
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
	case "down":
		m.moveCursor(1)
	case "up":
		m.moveCursor(-1)
	case "enter":
		m.openSelected()
	case "?":
		m.screen = screenHelp
		m.status = "help"
	case "home":
		m.goHome()
	case "b", "esc":
		m.goBack()
	case "q", "ctrl+c":
		m.requestQuit()
	default:
		m.status = m.normalHelp()
	}
	return m, cmd
}

func (m model) handleQuitKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "y", "Y", "enter":
		return m, tea.Quit
	case "n", "N", "esc", "q":
		m.mode = modeNormal
		m.status = m.normalHelp()
		return m, cmd
	default:
		m.status = "quit? y/n"
		return m, cmd
	}
}

func (m *model) requestQuit() {
	m.mode = modeQuit
	m.inputPurpose = inputNone
	m.input.SetValue("")
	m.resetInputPrompt()
	m.status = "quit? y/n"
}

func (m *model) moveCursor(delta int) {
	m.screen = screenProgram
	switch m.currentLevel() {
	case screenPrograms:
		m.programCursor = moveIndex(m.programCursor, delta, len(m.programs))
	case screenWorkouts:
		m.workoutCursor = moveIndex(m.workoutCursor, delta, len(m.workouts))
	case screenExercises:
		m.exerciseCursor = moveIndex(m.exerciseCursor, delta, len(m.exercises))
	}
}

func moveIndex(current int, delta int, length int) int {
	if length == 0 {
		return 0
	}
	next := current + delta
	if next < 0 {
		return 0
	}
	if next >= length {
		return length - 1
	}
	return next
}

func (m *model) openSelected() {
	m.screen = screenProgram
	switch m.currentLevel() {
	case screenPrograms:
		if len(m.programs) == 0 {
			m.status = "no programs yet. press a to add one"
			return
		}
		m.programCursor = clampIndex(m.programCursor, len(m.programs))
		program := m.programs[m.programCursor]
		m.activeProgram = program
		if err := m.loadWorkouts(program); err != nil {
			m.status = err.Error()
			return
		}
		m.activeWorkout = data.Workout{}
		m.activeExercise = data.Exercise{}
		m.workoutCursor = 0
		m.exerciseCursor = 0
		m.exercises = nil
		if len(m.workouts) == 0 {
			m.startAddWorkout()
			m.status = "no workouts in " + program.ProgramName + ". add the first workout"
			return
		}
		m.status = "selected program " + program.ProgramName + ". " + m.normalHelp()
	case screenWorkouts:
		if len(m.workouts) == 0 {
			m.status = "no workouts yet. press a to add one"
			return
		}
		m.workoutCursor = clampIndex(m.workoutCursor, len(m.workouts))
		workout := m.workouts[m.workoutCursor]
		m.activeWorkout = workout
		if err := m.loadExercises(workout); err != nil {
			m.status = err.Error()
			return
		}
		m.activeExercise = data.Exercise{}
		m.exerciseCursor = 0
		if len(m.exercises) == 0 {
			m.status = "selected workout " + workout.Name + ". press a to add an exercise"
			return
		}
		m.status = "selected workout " + workout.Name + ". " + m.normalHelp()
	case screenExercises:
		if len(m.exercises) == 0 {
			m.status = "no exercises yet. press a to add one"
			return
		}
		m.exerciseCursor = clampIndex(m.exerciseCursor, len(m.exercises))
		exercise := m.exercises[m.exerciseCursor]
		m.activeExercise = exercise
		m.status = "selected exercise " + exercise.Name + ". " + m.normalHelp()
	}
}

func clampIndex(current int, length int) int {
	if length == 0 || current < 0 {
		return 0
	}
	if current >= length {
		return length - 1
	}
	return current
}

func (m *model) goBack() {
	if m.screen == screenHelp {
		m.screen = screenProgram
		m.status = "back"
		return
	}
	switch {
	case m.activeExercise.ExerciseId != 0:
		m.activeExercise = data.Exercise{}
		m.status = "back to exercises"
	case m.activeWorkout.WorkoutId != 0:
		m.activeWorkout = data.Workout{}
		m.activeExercise = data.Exercise{}
		m.exercises = nil
		m.exerciseCursor = 0
		m.status = "back to workouts"
	case m.activeProgram.ProgramId != 0:
		m.activeProgram = data.Program{}
		m.activeWorkout = data.Workout{}
		m.activeExercise = data.Exercise{}
		m.workouts = nil
		m.exercises = nil
		m.workoutCursor = 0
		m.exerciseCursor = 0
		m.status = "back to programs"
	default:
		m.goHome()
	}
	m.screen = screenProgram
}

func (m *model) goHome() {
	m.screen = screenHome
	m.status = "home"
}

func (m model) currentLevel() screen {
	if m.activeWorkout.WorkoutId != 0 {
		return screenExercises
	}
	if m.activeProgram.ProgramId != 0 {
		return screenWorkouts
	}
	return screenPrograms
}

func (m model) normalHelp() string {
	switch m.currentLevel() {
	case screenPrograms:
		return "up/down move, enter open, a add program, : command"
	case screenWorkouts:
		return "up/down move, enter open, b programs, a add workout, : command"
	case screenExercises:
		return "up/down move, enter select, b workouts, a add exercise, : command"
	default:
		return "up/down move, enter open, b back, a add, : command"
	}
}

func (m *model) loadPrograms() error {
	programs, err := m.store.ListPrograms()
	if err != nil {
		return err
	}
	m.programs = programs
	m.programCursor = clampIndex(m.programCursor, len(m.programs))
	return nil
}

func (m *model) loadWorkouts(program data.Program) error {
	workouts, err := m.store.ListWorkouts(program)
	if err != nil {
		return err
	}
	m.workouts = workouts
	m.workoutCursor = clampIndex(m.workoutCursor, len(m.workouts))
	return nil
}

func (m *model) loadExercises(workout data.Workout) error {
	exercises, err := m.store.ListExercises(workout)
	if err != nil {
		return err
	}
	m.exercises = exercises
	m.exerciseCursor = clampIndex(m.exerciseCursor, len(m.exercises))
	return nil
}

func (m *model) startAdd() {
	m.mode = modeInput
	m.input.SetValue("")
	m.screen = screenProgram

	switch {
	case m.activeWorkout.WorkoutId != 0:
		m.startAddExercise()
	case m.activeProgram.ProgramId != 0:
		m.startAddWorkout()
	default:
		m.startAddProgram()
	}
}

func (m *model) startAddProgram() {
	m.mode = modeInput
	m.input.SetValue("")
	m.screen = screenProgram
	m.inputPurpose = inputAddProgram
	m.input.Placeholder = "program name"
	m.input.Prompt = "add program $ "
	m.status = "adding program"
}

func (m *model) startAddWorkout() {
	m.mode = modeInput
	m.input.SetValue("")
	m.screen = screenProgram
	m.inputPurpose = inputAddWorkout
	m.input.Placeholder = "workout name"
	m.input.Prompt = "add workout $ "
	m.status = "adding workout"
}

func (m *model) startAddExercise() {
	m.mode = modeInput
	m.input.SetValue("")
	m.screen = screenProgram
	m.inputPurpose = inputAddExercise
	m.input.Placeholder = "exercise name [sets] [reps]"
	m.input.Prompt = "add exercise $ "
	m.status = "adding exercise"
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
	case "log":
		m.handleLog(command.Args)
		return m, cmd
	case "history":
		m.handleHistory(command.Args)
		return m, cmd
	case "help":
		m.screen = screenHelp
		return m, cmd

	case "home":
		m.screen = screenHome
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
	case screenHome:
		screen = screens.HomeView(m.styles)

	case screenHelp:
		screen = screens.HelpView(m.styles)

	case screenProgram:
		screen = screens.ProgramView(m.styles, m.programs, m.workouts, m.exercises, m.activeProgram, m.activeWorkout, m.activeExercise, m.programCursor, m.workoutCursor, m.exerciseCursor)

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
	m.screen = screenProgram
	switch cmd {
	case "list":
		if err := m.loadPrograms(); err != nil {
			m.status = err.Error()
			return
		}

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
		m.programCursor = len(m.programs) - 1
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
		if err := m.loadWorkouts(program); err != nil {
			m.status = err.Error()
			return
		}
		m.activeWorkout = data.Workout{}
		m.activeExercise = data.Exercise{}
		m.exercises = nil
		m.workoutCursor = 0
		m.exerciseCursor = 0
		if len(m.workouts) == 0 {
			m.startAddWorkout()
			m.status = "no workouts in " + program.ProgramName + ". add the first workout"
			return
		}
		m.status = "Selected program " + program.ProgramName

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
	m.screen = screenProgram
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
		if err := m.loadWorkouts(m.activeProgram); err != nil {
			m.status = err.Error()
			return
		}
		m.workoutCursor = len(m.workouts) - 1
		m.status = "Created workout"

	case "list":
		if err := m.loadWorkouts(m.activeProgram); err != nil {
			m.status = err.Error()
			return
		}

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
		if err := m.loadExercises(workout); err != nil {
			m.status = err.Error()
			return
		}
		m.activeExercise = data.Exercise{}
		m.exerciseCursor = 0
		m.status = "Selected workout " + workout.Name
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
	m.screen = screenProgram
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
		if err := m.loadExercises(m.activeWorkout); err != nil {
			m.status = err.Error()
			return
		}
		m.exerciseCursor = len(m.exercises) - 1
		m.status = "Created exercise"

	case "list":
		if err := m.loadExercises(m.activeWorkout); err != nil {
			m.status = err.Error()
			return
		}

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
		m.status = "Selected exercise " + exercise.Name

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

func (m *model) handleLog(args []string) {
	if m.activeWorkout.WorkoutId == 0 {
		m.status = "select a workout first: workout select <id|name>"
		return
	}
	if len(args) == 0 {
		m.status = "usage: log start | log add [exercise] <sets> <reps> [weight] [notes] | log finish [notes] | log current"
		return
	}

	m.screen = screenProgram
	switch args[0] {
	case "start":
		session, err := m.store.StartGymSession(m.activeWorkout)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.status = fmt.Sprintf("Started session #%d for %s", session.SessionId, m.activeWorkout.Name)

	case "add":
		exercise, sets, reps, weight, notes, err := m.parseLogAddArgs(args[1:])
		if err != nil {
			m.status = err.Error()
			return
		}
		session, err := m.store.ActiveGymSession(m.activeWorkout)
		if err == sql.ErrNoRows {
			m.status = "start a session first: log start"
			return
		}
		if err != nil {
			m.status = err.Error()
			return
		}
		if err := m.store.AddGymSessionEntry(session, exercise, sets, reps, weight, notes); err != nil {
			m.status = err.Error()
			return
		}
		m.status = "Logged " + formatSessionEntry(data.GymSessionEntry{Exercise: exercise.Name, Sets: sets, Reps: reps, Weight: weight, Notes: notes})

	case "current":
		session, err := m.store.ActiveGymSession(m.activeWorkout)
		if err == sql.ErrNoRows {
			m.status = "no active session"
			return
		}
		if err != nil {
			m.status = err.Error()
			return
		}
		entries, err := m.store.ListGymSessionEntries(session)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.status = fmt.Sprintf("Session #%d started %s, %d entries", session.SessionId, session.StartedAt, len(entries))

	case "finish", "save":
		session, err := m.store.ActiveGymSession(m.activeWorkout)
		if err == sql.ErrNoRows {
			m.status = "no active session"
			return
		}
		if err != nil {
			m.status = err.Error()
			return
		}
		notes := strings.Join(args[1:], " ")
		if err := m.store.FinishGymSession(session, notes); err != nil {
			m.status = err.Error()
			return
		}
		m.status = fmt.Sprintf("Finished session #%d", session.SessionId)

	default:
		m.status = fmt.Sprintf("unknown log command: %s", args[0])
	}
}

func (m *model) handleHistory(args []string) {
	if m.activeWorkout.WorkoutId == 0 {
		m.status = "select a workout first: workout select <id|name>"
		return
	}
	if len(args) == 0 {
		m.status = "usage: history list [limit] | history show <session-id>"
		return
	}

	m.screen = screenProgram
	switch args[0] {
	case "list":
		limit := 5
		if len(args) >= 2 {
			parsed, err := strconv.Atoi(args[1])
			if err != nil {
				m.status = "limit must be a number"
				return
			}
			limit = parsed
		}
		sessions, err := m.store.ListGymSessions(m.activeWorkout, limit)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.status = formatSessionList(sessions)

	case "show":
		if len(args) < 2 {
			m.status = "usage: history show <session-id>"
			return
		}
		session, err := m.store.SelectGymSession(args[1], m.activeWorkout)
		if err != nil {
			m.status = err.Error()
			return
		}
		entries, err := m.store.ListGymSessionEntries(session)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.status = formatSessionDetail(session, entries)

	default:
		m.status = fmt.Sprintf("unknown history command: %s", args[0])
	}
}

func (m *model) parseLogAddArgs(args []string) (data.Exercise, int, int, float64, string, error) {
	if len(args) < 2 {
		return data.Exercise{}, 0, 0, 0, "", fmt.Errorf("usage: log add [exercise] <sets> <reps> [weight] [notes]")
	}

	exercise := m.activeExercise
	valueStart := 0
	if exercise.ExerciseId == 0 || !isInt(args[0]) {
		if len(args) < 3 {
			return data.Exercise{}, 0, 0, 0, "", fmt.Errorf("usage: log add [exercise] <sets> <reps> [weight] [notes]")
		}
		valueStart = findFirstNumberPair(args)
		if valueStart <= 0 {
			return data.Exercise{}, 0, 0, 0, "", fmt.Errorf("sets and reps must be numbers")
		}
		selected, err := m.store.SelectExercise(strings.Join(args[:valueStart], " "), m.activeWorkout)
		if err != nil {
			return data.Exercise{}, 0, 0, 0, "", err
		}
		exercise = selected
	}
	if exercise.ExerciseId == 0 {
		return data.Exercise{}, 0, 0, 0, "", fmt.Errorf("select an exercise first or pass one to log add")
	}

	sets, reps, err := parseSetReps(args[valueStart], args[valueStart+1])
	if err != nil {
		return data.Exercise{}, 0, 0, 0, "", err
	}

	weight := 0.0
	notesStart := valueStart + 2
	if len(args) > notesStart {
		parsedWeight, err := strconv.ParseFloat(args[notesStart], 64)
		if err == nil {
			weight = parsedWeight
			notesStart++
		}
	}

	notes := ""
	if len(args) > notesStart {
		notes = strings.Join(args[notesStart:], " ")
	}
	return exercise, sets, reps, weight, notes, nil
}

func findFirstNumberPair(args []string) int {
	for i := 0; i < len(args)-1; i++ {
		if isInt(args[i]) && isInt(args[i+1]) {
			return i
		}
	}
	return -1
}

func isInt(value string) bool {
	_, err := strconv.Atoi(value)
	return err == nil
}

func formatSessionList(sessions []data.GymSession) string {
	if len(sessions) == 0 {
		return "no sessions yet"
	}
	parts := make([]string, 0, len(sessions))
	for _, session := range sessions {
		state := "active"
		if session.EndedAt != "" {
			state = "done"
		}
		parts = append(parts, fmt.Sprintf("#%d %s %s", session.SessionId, state, session.StartedAt))
	}
	return strings.Join(parts, " | ")
}

func formatSessionDetail(session data.GymSession, entries []data.GymSessionEntry) string {
	if len(entries) == 0 {
		return fmt.Sprintf("Session #%d has no entries", session.SessionId)
	}
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		parts = append(parts, formatSessionEntry(entry))
	}
	return fmt.Sprintf("Session #%d: %s", session.SessionId, strings.Join(parts, " | "))
}

func formatSessionEntry(entry data.GymSessionEntry) string {
	value := fmt.Sprintf("%s %dx%d", entry.Exercise, entry.Sets, entry.Reps)
	if entry.Weight > 0 {
		value += fmt.Sprintf(" @ %.1f", entry.Weight)
	}
	if entry.Notes != "" {
		value += " (" + entry.Notes + ")"
	}
	return value
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
