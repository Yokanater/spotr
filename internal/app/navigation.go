package app

import "spotr/data"

func (m *model) moveCursor(delta int) {
	if m.screen == screenHome {
		m.programCursor = moveIndex(m.programCursor, delta, len(m.programs))
		return
	}
	if m.screen == screenHelp {
		return
	}
	if m.screen == screenTemplates {
		m.templateCursor = moveIndex(m.templateCursor, delta, len(m.templateFiles))
		return
	}
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
	if m.screen == screenTemplates {
		m.importSelectedTemplate()
		return
	}
	if m.screen == screenHistory {
		m.openSelectedHistory()
		return
	}
	if m.screen == screenHome {
		m.openHomeProgram()
		return
	}
	level := m.currentLevel()
	m.screen = screenProgram
	switch level {
	case screenPrograms:
		if len(m.programs) == 0 {
			m.status = "no programs yet. press a to add one"
			return
		}
		m.programCursor = clampIndex(m.programCursor, len(m.programs))
		program := m.programs[m.programCursor]
		if err := m.activateProgram(program); err != nil {
			m.status = err.Error()
			return
		}
		m.screen = screenProgram
		if len(m.workouts) == 0 {
			m.startAddWorkout()
			m.status = "no workouts in " + program.ProgramName + ". add the first workout"
			return
		}
		m.status = "Using program " + program.ProgramName
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
		m.status = "Opened workout " + workout.Name
	case screenExercises:
		if len(m.exercises) == 0 {
			m.status = "no exercises yet. press a to add one"
			return
		}
		m.exerciseCursor = clampIndex(m.exerciseCursor, len(m.exercises))
		exercise := m.exercises[m.exerciseCursor]
		m.activeExercise = exercise
		m.status = "Selected " + exercise.Name
	}
}

func (m *model) openHomeProgram() {
	if len(m.programs) == 0 {
		m.status = "Create a program or browse templates"
		return
	}
	m.programCursor = clampIndex(m.programCursor, len(m.programs))
	program := m.programs[m.programCursor]
	if err := m.activateProgram(program); err != nil {
		m.status = err.Error()
		return
	}
	m.screen = screenProgram
	if len(m.workouts) == 0 {
		m.status = "No workouts in " + program.ProgramName + ". Press a to add one."
		return
	}
	m.status = "Opened " + program.ProgramName
}

func (m *model) openSelectedHistory() {
	if m.activeSession.SessionId != 0 {
		m.status = "Session details"
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
		m.status = "Session details"
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
	m.status = "Session details"
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
	if m.screen == screenPrograms {
		if m.activeProgram.ProgramId == 0 {
			m.goHome()
			return
		}
		m.screen = screenProgram
		m.status = "Back to workouts."
		return
	}
	if m.screen == screenHelp {
		m.screen = m.returnDestination(m.helpReturnScreen)
		m.status = ""
		return
	}
	if m.screen == screenTemplates {
		m.screen = m.returnDestination(m.templateReturnScreen)
		m.status = ""
		return
	}
	if m.screen == screenHistory {
		if m.activeSession.SessionId != 0 {
			if m.historyBackEntries != nil {
				m.activeSession = data.GymSession{}
				m.historyEntries = m.historyBackEntries
				m.historyBackEntries = nil
				m.historyCursor = clampIndex(m.historyBackCursor, len(m.historyEntries))
				m.status = "Back to movement history"
				return
			}
			if m.historySessions == nil {
				m.activeSession = data.GymSession{}
				m.historyEntries = nil
				m.screen = screenProgram
				m.status = "Back to workouts"
				return
			}
			m.activeSession = data.GymSession{}
			m.historyEntries = nil
			m.historyCursor = clampIndex(m.historyCursor, len(m.historySessions))
			m.status = "Back to recent sessions"
			return
		}
		m.screen = screenProgram
		m.status = "Back to workouts"
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
		m.goHome()
		return
	default:
		m.goHome()
	}
	m.screen = screenProgram
}

func (m model) returnDestination(destination screen) screen {
	switch destination {
	case screenHome, screenProgram, screenPrograms, screenHistory, screenTemplates:
		return destination
	default:
		if m.activeProgram.ProgramId != 0 {
			return screenProgram
		}
		return screenHome
	}
}

func (m *model) goHome() {
	m.screen = screenHome
	m.status = "home"
}

func (m model) currentLevel() screen {
	if m.screen == screenPrograms {
		return screenPrograms
	}
	if m.activeWorkout.WorkoutId != 0 {
		return screenExercises
	}
	if m.activeProgram.ProgramId != 0 {
		return screenWorkouts
	}
	return screenPrograms
}

func (m *model) openProgramPicker() {
	m.screen = screenPrograms
	m.activeWorkout = data.Workout{}
	m.activeExercise = data.Exercise{}
	m.exercises = nil
	m.exerciseCursor = 0
	m.status = "Choose a program. Spotr will remember it next time."
}
