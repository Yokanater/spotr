package app

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"database/sql"
	"spotr/data"
	"spotr/store"
	"spotr/ui/screens"
	"spotr/ui/theme"
	"spotr/ui/utils"
)

type mode string

type screen string

type inputPurpose string

type deleteTarget string

const (
	modeNormal mode = "normal"
	modeInput  mode = "input"
	modeCmd    mode = "command"
	modeQuit   mode = "quit"
	modeDelete mode = "delete"
)

const (
	inputNone         inputPurpose = ""
	inputAddProgram   inputPurpose = "add_program"
	inputAddWorkout   inputPurpose = "add_workout"
	inputAddExercise  inputPurpose = "add_exercise"
	inputLogExercise  inputPurpose = "log_exercise"
	inputEditLog      inputPurpose = "edit_log"
	inputEditProgram  inputPurpose = "edit_program"
	inputEditWorkout  inputPurpose = "edit_workout"
	inputEditExercise inputPurpose = "edit_exercise"
)

const (
	deleteNone     deleteTarget = ""
	deleteLog      deleteTarget = "log"
	deleteProgram  deleteTarget = "program"
	deleteWorkout  deleteTarget = "workout"
	deleteExercise deleteTarget = "exercise"
)

const (
	screenHome      screen = "home"
	screenProgram   screen = "program"
	screenPrograms  screen = "programs"
	screenWorkouts  screen = "workouts"
	screenExercises screen = "exercises"
	screenHistory   screen = "history"
	screenTemplates screen = "templates"
	screenHelp      screen = "help"
)

type model struct {
	quitting             bool
	maxH                 int
	maxW                 int
	appH                 int
	appW                 int
	termH                int
	termW                int
	theme                theme.Theme
	screen               screen
	mode                 mode
	inputPurpose         inputPurpose
	styles               theme.Styles
	input                textinput.Model
	store                *store.Store
	status               string
	programCursor        int
	workoutCursor        int
	exerciseCursor       int
	historyCursor        int
	programs             []data.Program
	workouts             []data.Workout
	exercises            []data.Exercise
	historySessions      []data.GymSession
	historyEntries       []data.GymSessionEntry
	historyBackEntries   []data.GymSessionEntry
	historyBackCursor    int
	templateFiles        []programTemplateFile
	templateCursor       int
	activeSession        data.GymSession
	historyTitle         string
	activeProgram        data.Program
	activeWorkout        data.Workout
	activeExercise       data.Exercise
	editingEntry         data.GymSessionEntry
	editingProgram       data.Program
	editingWorkout       data.Workout
	editingExercise      data.Exercise
	deletingEntry        data.GymSessionEntry
	deletingProgram      data.Program
	deletingWorkout      data.Workout
	deletingExercise     data.Exercise
	deleteTarget         deleteTarget
	helpReturnScreen     screen
	templateReturnScreen screen
}

func initialModel(st *store.Store) model {
	ti := textinput.New()
	t := theme.Default()
	ti.Placeholder = ""
	ti.Prompt = ""
	ti.SetWidth(t.InputMax)
	ti.CharLimit = 128
	ti.Blur()
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
		if err := m.restoreActiveProgram(); err != nil {
			m.status = err.Error()
		} else if m.activeProgram.ProgramId != 0 {
			m.status = ""
		} else {
			m.status = ""
		}
	}
	return m
}

func (m *model) restoreActiveProgram() error {
	if len(m.programs) == 0 {
		return nil
	}
	program, err := m.store.ActiveProgram()
	if err == sql.ErrNoRows {
		program = m.programs[0]
	} else if err != nil {
		return err
	}
	return m.activateProgram(program)
}

func (m *model) activateProgram(program data.Program) error {
	m.activeProgram = program
	m.activeWorkout = data.Workout{}
	m.activeExercise = data.Exercise{}
	m.exercises = nil
	m.workoutCursor = 0
	m.exerciseCursor = 0
	if err := m.loadWorkouts(program); err != nil {
		return err
	}
	if err := m.store.SetActiveProgram(program); err != nil {
		return err
	}
	for i := range m.programs {
		if m.programs[i].ProgramId == program.ProgramId {
			m.programCursor = i
			break
		}
	}
	return nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func Run(st *store.Store) error {
	_, err := tea.NewProgram(initialModel(st)).Run()
	return err
}

func (m model) View() tea.View {
	if m.quitting {
		return tea.NewView("bye bye")
	}

	input := m.styles.Input.Render(m.input.View())
	status := renderStatus(m.styles, m.status)
	keyRail := renderKeyRail(m.styles, m.keyHelp())
	screenHeight := max(1, m.appH-lipgloss.Height(input)-lipgloss.Height(status)-lipgloss.Height(keyRail)-2)
	screenStyles := theme.NewStyles(m.theme, m.appW, screenHeight)
	screen := ""
	switch m.screen {
	case screenHome:
		screen = screens.HomeView(screenStyles, m.programs, m.programCursor, m.activeProgram)

	case screenHelp:
		screen = screens.HelpView(screenStyles)

	case screenProgram, screenPrograms:
		screen = screens.ProgramView(screenStyles, m.programs, m.workouts, m.exercises, m.activeProgram, m.activeWorkout, m.activeExercise, m.programCursor, m.workoutCursor, m.exerciseCursor, m.screen == screenPrograms)

	case screenHistory:
		screen = screens.HistoryView(screenStyles, m.activeWorkout, m.historyTitle, m.historySessions, m.historyCursor, m.activeSession, m.historyEntries)

	case screenTemplates:
		screen = screens.TemplatesView(screenStyles, m.templateItems(), m.templateCursor)

	}
	screen = lipgloss.NewStyle().
		Width(m.styles.Box.GetWidth()).
		MaxHeight(screenHeight).
		Render(screen)
	join := lipgloss.JoinVertical(lipgloss.Center, screen, input, status, keyRail)
	box := m.styles.Box.Render(join)
	v := tea.NewView(
		utils.CenterPlace(m.termW, m.termH, box),
	)
	v.BackgroundColor = m.theme.Background
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}
