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
	modeDelete mode = "delete"
)

const (
	inputNone        inputPurpose = ""
	inputAddProgram  inputPurpose = "add_program"
	inputAddWorkout  inputPurpose = "add_workout"
	inputAddExercise inputPurpose = "add_exercise"
	inputLogExercise inputPurpose = "log_exercise"
	inputEditLog     inputPurpose = "edit_log"
)

const (
	screenHome      screen = "home"
	screenProgram   screen = "program"
	screenPrograms  screen = "programs"
	screenWorkouts  screen = "workouts"
	screenExercises screen = "exercises"
	screenHistory   screen = "history"
	screenHelp      screen = "help"
)

type model struct {
	quitting           bool
	maxH               int
	maxW               int
	appH               int
	appW               int
	termH              int
	termW              int
	theme              theme.Theme
	screen             screen
	mode               mode
	inputPurpose       inputPurpose
	styles             theme.Styles
	input              textinput.Model
	store              *store.Store
	status             string
	programCursor      int
	workoutCursor      int
	exerciseCursor     int
	historyCursor      int
	programs           []data.Program
	workouts           []data.Workout
	exercises          []data.Exercise
	historySessions    []data.GymSession
	historyEntries     []data.GymSessionEntry
	historyBackEntries []data.GymSessionEntry
	historyBackCursor  int
	activeSession      data.GymSession
	historyTitle       string
	activeProgram      data.Program
	activeWorkout      data.Workout
	activeExercise     data.Exercise
	editingEntry       data.GymSessionEntry
	deletingEntry      data.GymSessionEntry
}

func initialModel(st *store.Store) model {
	ti := textinput.New()
	t := theme.Default()
	ti.Placeholder = ""
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
		status: "",
	}
	if err := m.loadPrograms(); err != nil {
		m.status = err.Error()
	} else {
		m.status = m.normalHelp()
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
		m.startEditLogEntryInput()
	case "d":
		m.requestDeleteLogEntry()
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
		m.status = "quit? y/n"
		return m, cmd
	}
}

func (m model) handleDeleteKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case "y", "Y", "enter":
		m.confirmDeleteLogEntry()
		return m, cmd
	case "n", "N", "esc", "q":
		m.mode = modeNormal
		m.deletingEntry = data.GymSessionEntry{}
		m.status = "delete cancelled"
		return m, cmd
	default:
		m.status = helperMessage("delete selected log?", "y confirm", "n cancel")
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
		return helperMessage("up/down move", "enter open program", "a add program", ": command")
	case screenWorkouts:
		if len(m.workouts) == 0 {
			return helperMessage("a add your first workout", "b programs", ": command")
		}
		return helperMessage("up/down move", "enter open workout", "v workout logs", "a add workout", "b programs")
	case screenExercises:
		if len(m.exercises) == 0 {
			return helperMessage("a add your first exercise", "b workouts", ": command")
		}
		return helperMessage("up/down move", "l log actual sets", "v exercise logs", "s start log", "b workouts")
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

func (m *model) startLogSession() {
	m.screen = screenProgram
	if m.activeWorkout.WorkoutId == 0 {
		m.status = "select a workout first"
		return
	}

	session, err := m.store.StartGymSession(m.activeWorkout)
	if err != nil {
		m.status = err.Error()
		return
	}
	m.status = fmt.Sprintf("Started session #%d for %s", session.SessionId, m.activeWorkout.Name)
}

func (m *model) startLogExerciseInput() {
	m.screen = screenProgram
	exercise, ok := m.exerciseForLogging()
	if !ok {
		return
	}

	suggestion := ""
	if exercise.Sets > 0 || exercise.Reps > 0 {
		suggestion = fmt.Sprintf("%d %d", exercise.Sets, exercise.Reps)
	}

	m.mode = modeInput
	m.inputPurpose = inputLogExercise
	m.input.SetValue(suggestion)
	m.input.Placeholder = "sets reps or reps/reps [weight] [notes]"
	m.input.Prompt = "log " + exercise.Name + " $ "
	if suggestion != "" {
		m.status = helperMessage(fmt.Sprintf("suggested %dx%d", exercise.Sets, exercise.Reps), "use 6/4 for mixed reps", "enter log", "esc cancel")
		return
	}
	m.status = helperMessage("enter actual sets and reps for "+exercise.Name, "enter log", "esc cancel")
}

func (m *model) submitLoggedExercise(value string) {
	exercise := m.activeExercise
	if exercise.ExerciseId == 0 {
		m.status = "choose an exercise to log"
		return
	}

	sets, reps, repsDetail, weight, notes, err := parseLoggedExerciseValue(value)
	if err != nil {
		m.status = err.Error()
		return
	}

	session, started, err := m.activeOrStartedSession()
	if err != nil {
		m.status = err.Error()
		return
	}
	if err := m.store.AddGymSessionEntry(session, exercise, sets, reps, repsDetail, weight, notes); err != nil {
		m.status = err.Error()
		return
	}

	entry := data.GymSessionEntry{Exercise: exercise.Name, Sets: sets, Reps: reps, RepsDetail: repsDetail, Weight: weight, Notes: notes}
	if started {
		m.status = fmt.Sprintf("Started session #%d. Logged %s", session.SessionId, formatSessionEntry(entry))
		return
	}
	m.status = "Logged " + formatSessionEntry(entry)
}

func (m *model) startEditLogEntryInput() {
	entry, ok := m.selectedLogEntry()
	if !ok {
		m.status = "select a logged entry first"
		return
	}

	m.mode = modeInput
	m.inputPurpose = inputEditLog
	m.editingEntry = entry
	m.input.SetValue(logEntryInputValue(entry))
	m.input.Placeholder = "sets reps or reps/reps [weight] [notes]"
	m.input.Prompt = "edit log #" + strconv.FormatInt(entry.EntryId, 10) + " $ "
	m.status = helperMessage("edit sets reps weight notes", "enter save", "esc cancel")
}

func (m *model) submitEditedLogEntry(value string) {
	if m.editingEntry.EntryId == 0 {
		m.status = "no log selected to edit"
		return
	}

	sets, reps, repsDetail, weight, notes, err := parseLoggedExerciseValue(value)
	if err != nil {
		m.status = err.Error()
		return
	}
	if err := m.store.UpdateGymSessionEntry(m.editingEntry, sets, reps, repsDetail, weight, notes); err != nil {
		m.status = err.Error()
		return
	}

	updated := m.editingEntry
	updated.Sets = sets
	updated.Reps = reps
	updated.RepsDetail = repsDetail
	updated.Weight = weight
	updated.Notes = notes
	m.editingEntry = data.GymSessionEntry{}
	m.refreshHistoryAfterEntryUpdate(updated)
	m.status = "Updated " + formatSessionEntry(updated)
}

func (m *model) requestDeleteLogEntry() {
	entry, ok := m.selectedLogEntry()
	if !ok {
		m.status = "select a logged entry first"
		return
	}
	m.mode = modeDelete
	m.deletingEntry = entry
	m.inputPurpose = inputNone
	m.input.SetValue("")
	m.resetInputPrompt()
	m.status = helperMessage("delete "+formatSessionEntry(entry)+"?", "y confirm", "n cancel")
}

func (m *model) confirmDeleteLogEntry() {
	if m.deletingEntry.EntryId == 0 {
		m.mode = modeNormal
		m.status = "no log selected to delete"
		return
	}
	deleted := m.deletingEntry
	if err := m.store.DeleteGymSessionEntry(deleted); err != nil {
		m.mode = modeNormal
		m.deletingEntry = data.GymSessionEntry{}
		m.status = err.Error()
		return
	}

	m.mode = modeNormal
	m.deletingEntry = data.GymSessionEntry{}
	m.refreshHistoryAfterEntryDelete(deleted)
	m.status = "Deleted " + formatSessionEntry(deleted)
}

func (m *model) finishLogSession() {
	m.screen = screenProgram
	if m.activeWorkout.WorkoutId == 0 {
		m.status = "select a workout first"
		return
	}

	session, err := m.store.ActiveGymSession(m.activeWorkout)
	if err == sql.ErrNoRows {
		m.status = "no active session"
		return
	}
	if err != nil {
		m.status = err.Error()
		return
	}
	if err := m.store.FinishGymSession(session, ""); err != nil {
		m.status = err.Error()
		return
	}
	m.status = fmt.Sprintf("Finished session #%d", session.SessionId)
}

func (m *model) viewRecentLogs() {
	if m.activeWorkout.WorkoutId == 0 {
		workout, ok := m.workoutForHistory()
		if !ok {
			m.status = "select a workout first"
			return
		}
		m.viewWorkoutSessions(workout)
		return
	}

	if exercise, ok := m.exerciseForHistory(); ok {
		entries, err := m.store.ListExerciseLogEntries(m.activeProgram, exercise.Name, 20)
		if err != nil {
			m.status = err.Error()
			return
		}
		m.historySessions = nil
		m.historyEntries = entries
		m.historyBackEntries = nil
		m.activeSession = data.GymSession{}
		m.historyTitle = exercise.Name + " across " + m.activeProgram.ProgramName
		m.historyCursor = 0
		m.screen = screenHistory
		m.status = helperMessage("j/k scroll", "enter open", "e edit", "d delete", "b back")
		return
	}

	m.viewWorkoutSessions(m.activeWorkout)
}

func (m *model) viewWorkoutSessions(workout data.Workout) {
	sessions, err := m.store.ListGymSessions(workout, 8)
	if err != nil {
		m.status = err.Error()
		return
	}
	m.historySessions = sessions
	m.historyEntries = nil
	m.historyBackEntries = nil
	m.activeSession = data.GymSession{}
	m.historyTitle = workout.Name + " sessions"
	m.historyCursor = clampIndex(m.historyCursor, len(sessions))
	m.screen = screenHistory
	m.status = helperMessage("j/k scroll", "enter open", "b back")
}

func (m *model) workoutForHistory() (data.Workout, bool) {
	if m.activeWorkout.WorkoutId != 0 {
		return m.activeWorkout, true
	}
	if m.activeProgram.ProgramId == 0 || len(m.workouts) == 0 {
		return data.Workout{}, false
	}
	m.workoutCursor = clampIndex(m.workoutCursor, len(m.workouts))
	return m.workouts[m.workoutCursor], true
}

func (m *model) exerciseForHistory() (data.Exercise, bool) {
	if m.activeExercise.ExerciseId != 0 {
		return m.activeExercise, true
	}
	if m.activeWorkout.WorkoutId == 0 || len(m.exercises) == 0 {
		return data.Exercise{}, false
	}
	m.exerciseCursor = clampIndex(m.exerciseCursor, len(m.exercises))
	return m.exercises[m.exerciseCursor], true
}

func (m *model) exerciseForLogging() (data.Exercise, bool) {
	if m.activeWorkout.WorkoutId == 0 {
		m.status = "select a workout first"
		return data.Exercise{}, false
	}
	if len(m.exercises) == 0 {
		m.status = "no exercises yet. press a to add one"
		return data.Exercise{}, false
	}

	m.exerciseCursor = clampIndex(m.exerciseCursor, len(m.exercises))
	exercise := m.exercises[m.exerciseCursor]
	m.activeExercise = exercise
	return exercise, true
}

func (m *model) activeOrStartedSession() (data.GymSession, bool, error) {
	session, err := m.store.ActiveGymSession(m.activeWorkout)
	if err == nil {
		return session, false, nil
	}
	if err != sql.ErrNoRows {
		return data.GymSession{}, false, err
	}

	session, err = m.store.StartGymSession(m.activeWorkout)
	if err != nil {
		return data.GymSession{}, false, err
	}
	return session, true, nil
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
	default:
		m.status = "nothing to submit"
	}
}

func inputCancelledStatus(purpose inputPurpose) string {
	if purpose == inputEditLog {
		return "edit cancelled"
	}
	if purpose == inputLogExercise {
		return "log cancelled"
	}
	return "add cancelled"
}

func (m *model) resetInputPrompt() {
	m.input.Placeholder = ""
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
	status := renderStatus(m.styles, m.status)
	screenHeight := max(1, m.appH-lipgloss.Height(input)-lipgloss.Height(status)-2)
	screenStyles := theme.NewStyles(m.theme, m.appW, screenHeight)
	screen := ""
	switch m.screen {
	case screenHome:
		screen = screens.HomeView(screenStyles)

	case screenHelp:
		screen = screens.HelpView(screenStyles)

	case screenProgram:
		screen = screens.ProgramView(screenStyles, m.programs, m.workouts, m.exercises, m.activeProgram, m.activeWorkout, m.activeExercise, m.programCursor, m.workoutCursor, m.exerciseCursor)

	case screenHistory:
		screen = screens.HistoryView(screenStyles, m.activeWorkout, m.historyTitle, m.historySessions, m.historyCursor, m.activeSession, m.historyEntries)

	}
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
	v.MouseMode = tea.MouseModeCellMotion
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
		m.status = "usage: log start | log add [exercise] <sets> <reps> | log add [exercise] <reps/reps> | log finish [notes] | log current"
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
		exercise, sets, reps, repsDetail, weight, notes, err := m.parseLogAddArgs(args[1:])
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
		if err := m.store.AddGymSessionEntry(session, exercise, sets, reps, repsDetail, weight, notes); err != nil {
			m.status = err.Error()
			return
		}
		m.status = "Logged " + formatSessionEntry(data.GymSessionEntry{Exercise: exercise.Name, Sets: sets, Reps: reps, RepsDetail: repsDetail, Weight: weight, Notes: notes})

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
		m.historySessions = sessions
		m.historyEntries = nil
		m.activeSession = data.GymSession{}
		m.historyTitle = m.activeWorkout.Name + " sessions"
		m.historyCursor = clampIndex(m.historyCursor, len(sessions))
		m.screen = screenHistory
		m.status = helperMessage("j/k scroll", "enter open", "b back")

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
		m.historySessions = nil
		m.historyEntries = entries
		m.activeSession = session
		m.historyTitle = m.activeWorkout.Name + " sessions"
		m.screen = screenHistory
		m.status = helperMessage("j/k scroll", "e edit", "d delete", "b back")

	default:
		m.status = fmt.Sprintf("unknown history command: %s", args[0])
	}
}

func (m *model) selectedLogEntry() (data.GymSessionEntry, bool) {
	if m.screen != screenHistory || len(m.historyEntries) == 0 {
		return data.GymSessionEntry{}, false
	}
	m.historyCursor = clampIndex(m.historyCursor, len(m.historyEntries))
	return m.historyEntries[m.historyCursor], true
}

func (m *model) refreshHistoryAfterEntryUpdate(entry data.GymSessionEntry) {
	for i := range m.historyEntries {
		if m.historyEntries[i].EntryId == entry.EntryId {
			m.historyEntries[i].Sets = entry.Sets
			m.historyEntries[i].Reps = entry.Reps
			m.historyEntries[i].RepsDetail = entry.RepsDetail
			m.historyEntries[i].Weight = entry.Weight
			m.historyEntries[i].Notes = entry.Notes
			break
		}
	}
	for i := range m.historyBackEntries {
		if m.historyBackEntries[i].EntryId == entry.EntryId {
			m.historyBackEntries[i].Sets = entry.Sets
			m.historyBackEntries[i].Reps = entry.Reps
			m.historyBackEntries[i].RepsDetail = entry.RepsDetail
			m.historyBackEntries[i].Weight = entry.Weight
			m.historyBackEntries[i].Notes = entry.Notes
			break
		}
	}

	if m.activeSession.SessionId != 0 {
		m.reloadActiveSessionEntries()
	}
}

func (m *model) refreshHistoryAfterEntryDelete(entry data.GymSessionEntry) {
	m.historyEntries = removeHistoryEntry(m.historyEntries, entry.EntryId)
	m.historyBackEntries = removeHistoryEntry(m.historyBackEntries, entry.EntryId)
	m.historyCursor = clampIndex(m.historyCursor, len(m.historyEntries))

	if m.activeSession.SessionId != 0 {
		m.reloadActiveSessionEntries()
	}
}

func (m *model) reloadActiveSessionEntries() {
	entries, err := m.store.ListGymSessionEntries(m.activeSession)
	if err != nil {
		m.status = err.Error()
		return
	}
	m.historyEntries = entries
	m.historyCursor = clampIndex(m.historyCursor, len(m.historyEntries))
}

func removeHistoryEntry(entries []data.GymSessionEntry, entryID int64) []data.GymSessionEntry {
	if entries == nil {
		return nil
	}
	filtered := entries[:0]
	for _, entry := range entries {
		if entry.EntryId != entryID {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (m *model) parseLogAddArgs(args []string) (data.Exercise, int, int, string, float64, string, error) {
	if len(args) < 1 {
		return data.Exercise{}, 0, 0, "", 0, "", fmt.Errorf("usage: log add [exercise] <sets> <reps> or <reps/reps> [weight] [notes]")
	}

	exercise := m.activeExercise
	valueStart := 0
	if exercise.ExerciseId == 0 || !isLogValueStart(args, 0) {
		if len(args) < 2 {
			return data.Exercise{}, 0, 0, "", 0, "", fmt.Errorf("usage: log add [exercise] <sets> <reps> or <reps/reps> [weight] [notes]")
		}
		valueStart = findFirstLogValue(args)
		if valueStart <= 0 {
			return data.Exercise{}, 0, 0, "", 0, "", fmt.Errorf("sets and reps must be numbers")
		}
		selected, err := m.store.SelectExercise(strings.Join(args[:valueStart], " "), m.activeWorkout)
		if err != nil {
			return data.Exercise{}, 0, 0, "", 0, "", err
		}
		exercise = selected
	}
	if exercise.ExerciseId == 0 {
		return data.Exercise{}, 0, 0, "", 0, "", fmt.Errorf("select an exercise first or pass one to log add")
	}

	sets, reps, repsDetail, weight, notes, err := parseLoggedExerciseValue(strings.Join(args[valueStart:], " "))
	if err != nil {
		return data.Exercise{}, 0, 0, "", 0, "", err
	}
	return exercise, sets, reps, repsDetail, weight, notes, nil
}

func findFirstLogValue(args []string) int {
	for i := range args {
		if isLogValueStart(args, i) {
			return i
		}
	}
	return -1
}

func isLogValueStart(args []string, index int) bool {
	if index >= len(args) {
		return false
	}
	if isValidRepsDetailToken(args[index]) {
		return true
	}
	return index+1 < len(args) && isInt(args[index]) && isInt(args[index+1])
}

func isInt(value string) bool {
	_, err := strconv.Atoi(value)
	return err == nil
}

func parseLoggedExerciseValue(value string) (int, int, string, float64, string, error) {
	args := strings.Fields(value)
	if len(args) < 1 {
		return 0, 0, "", 0, "", fmt.Errorf("usage: sets reps [weight] [notes] or reps/reps [weight] [notes]")
	}

	sets := 0
	reps := 0
	repsDetail := ""
	notesStart := 0
	if isRepsDetailToken(args[0]) {
		parsedReps, err := parseRepsDetail(args[0])
		if err != nil {
			return 0, 0, "", 0, "", err
		}
		sets = len(parsedReps)
		reps = parsedReps[len(parsedReps)-1]
		repsDetail = strings.Join(intStrings(parsedReps), "/")
		notesStart = 1
	} else {
		if len(args) < 2 {
			return 0, 0, "", 0, "", fmt.Errorf("usage: sets reps [weight] [notes] or reps/reps [weight] [notes]")
		}
		var err error
		sets, reps, err = parseSetReps(args[0], args[1])
		if err != nil {
			return 0, 0, "", 0, "", err
		}
		notesStart = 2
	}

	if sets <= 0 {
		return 0, 0, "", 0, "", fmt.Errorf("sets must be greater than zero")
	}
	if reps <= 0 {
		return 0, 0, "", 0, "", fmt.Errorf("reps must be greater than zero")
	}

	weight := 0.0
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
	return sets, reps, repsDetail, weight, notes, nil
}

func isRepsDetailToken(value string) bool {
	return strings.Contains(value, "/")
}

func isValidRepsDetailToken(value string) bool {
	_, err := parseRepsDetail(value)
	return err == nil
}

func parseRepsDetail(value string) ([]int, error) {
	parts := strings.Split(value, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("use reps/reps for per-set reps")
	}
	reps := make([]int, 0, len(parts))
	for _, part := range parts {
		parsed, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("per-set reps must be numbers")
		}
		if parsed <= 0 {
			return nil, fmt.Errorf("per-set reps must be greater than zero")
		}
		reps = append(reps, parsed)
	}
	return reps, nil
}

func intStrings(values []int) []string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.Itoa(value))
	}
	return parts
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
	value := fmt.Sprintf("%s %s", entry.Exercise, setRepLabel(entry))
	if entry.Weight > 0 {
		value += fmt.Sprintf(" @ %.1f", entry.Weight)
	}
	if entry.Notes != "" {
		value += " (" + entry.Notes + ")"
	}
	return value
}

func setRepLabel(entry data.GymSessionEntry) string {
	if entry.RepsDetail != "" {
		return fmt.Sprintf("%dx%s", entry.Sets, entry.RepsDetail)
	}
	return fmt.Sprintf("%dx%d", entry.Sets, entry.Reps)
}

func logEntryInputValue(entry data.GymSessionEntry) string {
	parts := []string{}
	if entry.RepsDetail != "" {
		parts = append(parts, entry.RepsDetail)
	} else {
		parts = append(parts, strconv.Itoa(entry.Sets), strconv.Itoa(entry.Reps))
	}
	if entry.Weight > 0 {
		parts = append(parts, strconv.FormatFloat(entry.Weight, 'f', -1, 64))
	}
	if entry.Notes != "" {
		parts = append(parts, entry.Notes)
	}
	return strings.Join(parts, " ")
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
