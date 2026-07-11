package app

import (
	"database/sql"
	"fmt"
	"github.com/Yokanater/spotr/data"
)

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
	m.input.Focus()
	m.inputPurpose = inputLogExercise
	m.input.SetValue(suggestion)
	m.input.Placeholder = "sets reps or reps/reps [weight] [notes]"
	m.input.Prompt = "log " + exercise.Name + " $ "
	if suggestion != "" {
		m.status = fmt.Sprintf("Suggested %dx%d · mixed reps: 6/4", exercise.Sets, exercise.Reps)
		return
	}
	m.status = "Enter sets and reps for " + exercise.Name
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
		m.status = exercise.Name + " history"
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
	m.status = workout.Name + " sessions"
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
