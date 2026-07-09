package app

import "ruffnut/data"

func (m *model) moveCursor(delta int) {
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
	if m.screen == screenTemplates {
		m.importSelectedTemplate()
		return
	}
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
	if m.screen == screenTemplates {
		m.screen = screenProgram
		m.status = m.normalHelp()
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
