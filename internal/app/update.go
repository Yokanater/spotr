package app

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"ruffnut/data"
	"ruffnut/ui/theme"
	"strings"
)

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
	case tea.MouseWheelMsg:
		return m.handleMouseWheel(msg)
	case tea.KeyPressMsg:
		switch m.mode {
		case modeQuit:
			return m.handleQuitKey(msg)

		case modeDelete:
			return m.handleDeleteKey(msg)

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
			m.status = inputCancelledStatus(purpose)
			return m, cmd
		}
		m.submitInput(purpose, value)
		return m, cmd
	case "ctrl+c":
		m.requestQuit()
		return m, cmd
	case "esc":
		purpose := m.inputPurpose
		m.input.SetValue("")
		m.mode = modeNormal
		m.inputPurpose = inputNone
		m.resetInputPrompt()
		m.status = inputCancelledStatus(purpose)
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
		m.status = helperMessage("type a command", "enter run", "esc cancel")
	case "a":
		m.startAdd()
	case "s":
		m.startLogSession()
	case "l":
		m.startLogExerciseInput()
	case "e":
		m.startEditSelectedInput()
	case "d":
		m.requestDeleteSelected()
	case "v":
		m.viewRecentLogs()
	case "f":
		m.finishLogSession()
	case "down", "j":
		m.moveCursor(1)
	case "up", "k":
		m.moveCursor(-1)
	case "pgdown", "ctrl+f":
		m.moveCursor(5)
	case "pgup", "ctrl+b":
		m.moveCursor(-5)
	case "enter":
		m.openSelected()
	case "?":
		m.screen = screenHelp
		m.status = helperMessage("b back", ": command")
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

func (m model) handleMouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.Mouse().Button {
	case tea.MouseWheelDown:
		m.moveCursor(1)
	case tea.MouseWheelUp:
		m.moveCursor(-1)
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
		m.status = quitConfirmStatus()
		return m, cmd
	}
}

func (m model) handleDeleteKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "y", "Y", "enter":
		m.confirmDeleteSelected()
		return m, cmd
	case "n", "N", "esc", "q":
		m.clearDeleteState()
		m.status = "delete cancelled"
		return m, cmd
	default:
		m.status = m.deleteConfirmStatus()
		return m, cmd
	}
}

func (m *model) requestQuit() {
	m.mode = modeQuit
	m.inputPurpose = inputNone
	m.input.SetValue("")
	m.resetInputPrompt()
	m.status = quitConfirmStatus()
}

func quitConfirmStatus() string {
	return helperMessage("quit spotr?", "y confirm", "n cancel")
}

func (m *model) moveCursor(delta int) {
	if m.screen == screenHistory {
		switch {
		case m.activeSession.SessionId != 0:
			m.historyCursor = moveIndex(m.historyCursor, delta, len(m.historyEntries))
		case m.historyEntries != nil:
			m.historyCursor = moveIndex(m.historyCursor, delta, len(m.historyEntries))
		default:
			m.historyCursor = moveIndex(m.historyCursor, delta, len(m.historySessions))
		}
		return
	}
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
	if m.screen == screenHistory {
		m.openSelectedHistory()
		return
	}
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

func (m *model) openSelectedHistory() {
	if m.activeSession.SessionId != 0 {
		m.status = helperMessage("e edit log", "d delete log", "b back to logs", ": command")
		return
	}
	if m.historyEntries != nil {
		if len(m.historyEntries) == 0 {
			m.status = "no linked logs yet"
			return
		}
		m.historyCursor = clampIndex(m.historyCursor, len(m.historyEntries))
		entry := m.historyEntries[m.historyCursor]
		session, err := m.store.SelectGymSessionByID(entry.SessionId)
		if err != nil {
			m.status = err.Error()
			return
		}
		entries, err := m.store.ListGymSessionEntries(session)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.activeSession = session
		m.historyBackEntries = m.historyEntries
		m.historyBackCursor = m.historyCursor
		m.historyEntries = entries
		m.historyCursor = 0
		m.status = helperMessage("j/k scroll", "e edit", "d delete", "b back to logs")
		return
	}
	if len(m.historySessions) == 0 {
		m.status = "no logs yet"
		return
	}
	m.historyCursor = clampIndex(m.historyCursor, len(m.historySessions))
	session := m.historySessions[m.historyCursor]
	entries, err := m.store.ListGymSessionEntries(session)
	if err != nil {
		m.status = err.Error()
		return
	}
	m.historyEntries = entries
	m.historyBackEntries = nil
	m.activeSession = session
	m.historyCursor = 0
	m.status = helperMessage("j/k scroll", "e edit", "d delete", "b back to logs")
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
	if m.screen == screenHistory {
		if m.activeSession.SessionId != 0 {
			if m.historyBackEntries != nil {
				m.activeSession = data.GymSession{}
				m.historyEntries = m.historyBackEntries
				m.historyBackEntries = nil
				m.historyCursor = clampIndex(m.historyBackCursor, len(m.historyEntries))
				m.status = helperMessage("j/k scroll", "enter open", "e edit", "d delete", "b training")
				return
			}
			if m.historySessions == nil {
				m.activeSession = data.GymSession{}
				m.historyEntries = nil
				m.screen = screenProgram
				m.status = m.normalHelp()
				return
			}
			m.activeSession = data.GymSession{}
			m.historyEntries = nil
			m.historyCursor = clampIndex(m.historyCursor, len(m.historySessions))
			m.status = helperMessage("j/k scroll", "enter open", "b training")
			return
		}
		m.screen = screenProgram
		m.status = m.normalHelp()
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
	if m.screen == screenHistory {
		if m.historyEntries != nil {
			return helperMessage("j/k scroll", "enter open", "e edit", "d delete", "b back")
		}
		return helperMessage("j/k scroll", "enter open", "b back")
	}

	switch m.currentLevel() {
	case screenPrograms:
		if len(m.programs) == 0 {
			return helperMessage("a add your first program", ": command", "? help")
		}
		return helperMessage("up/down move", "enter open", "e edit", "d delete", "a add")
	case screenWorkouts:
		if len(m.workouts) == 0 {
			return helperMessage("a add your first workout", "b programs", ": command")
		}
		return helperMessage("up/down move", "enter open workout", "v workout logs", "a add workout", "b programs")
	case screenExercises:
		if len(m.exercises) == 0 {
			return helperMessage("a add your first exercise", "b workouts", ": command")
		}
		return helperMessage("up/down move", "l log", "v logs", "e edit", "d delete", "b workouts")
	default:
		return helperMessage("up/down move", "enter open", "b back", "a add", ": command")
	}
}

func helperMessage(parts ...string) string {
	return strings.Join(parts, " · ")
}

func renderStatus(styles theme.Styles, status string) string {
	parts := strings.Split(status, " · ")
	if len(parts) == 1 {
		return styles.Status.Render(status)
	}

	rendered := make([]string, 0, len(parts)*2-1)
	for i, part := range parts {
		if i > 0 {
			rendered = append(rendered, styles.HelperSeparator.Render(" · "))
		}
		rendered = append(rendered, renderHelperPart(styles, part))
	}

	return styles.Status.Render(lipgloss.JoinHorizontal(lipgloss.Top, rendered...))
}

func renderHelperPart(styles theme.Styles, part string) string {
	key, rest, ok := strings.Cut(part, " ")
	if !ok || !isHelperKey(key) {
		return part
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, styles.HelperKey.Render(key), " "+rest)
}

func isHelperKey(value string) bool {
	switch value {
	case ":", "?", "a", "b", "d", "e", "enter", "esc", "f", "j/k", "l", "n", "s", "v", "up/down", "y":
		return true
	default:
		return false
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
	m.status = helperMessage("type a program name", "enter create", "esc cancel")
}

func (m *model) startAddWorkout() {
	m.mode = modeInput
	m.input.SetValue("")
	m.screen = screenProgram
	m.inputPurpose = inputAddWorkout
	m.input.Placeholder = "workout name"
	m.input.Prompt = "add workout $ "
	m.status = helperMessage("type a workout name", "enter create", "esc cancel")
}

func (m *model) startAddExercise() {
	m.mode = modeInput
	m.input.SetValue("")
	m.screen = screenProgram
	m.inputPurpose = inputAddExercise
	m.input.Placeholder = "exercise name [sets] [reps]"
	m.input.Prompt = "add exercise $ "
	m.status = helperMessage("type exercise name plus optional sets and reps", "enter create", "esc cancel")
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
	case inputLogExercise:
		m.submitLoggedExercise(value)
	case inputEditLog:
		m.submitEditedLogEntry(value)
	case inputEditProgram:
		m.submitEditedProgram(value)
	case inputEditExercise:
		m.submitEditedExercise(value)
	default:
		m.status = "nothing to submit"
	}
}

func inputCancelledStatus(purpose inputPurpose) string {
	switch purpose {
	case inputEditLog, inputEditProgram, inputEditExercise:
		return "edit cancelled"
	case inputLogExercise:
		return "log cancelled"
	default:
		return "add cancelled"
	}
}

func (m *model) resetInputPrompt() {
	m.input.Placeholder = ""
	m.input.Prompt = ""
}
