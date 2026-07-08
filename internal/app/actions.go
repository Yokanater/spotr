package app

import (
	"database/sql"
	"fmt"
	"ruffnut/data"
	"strconv"
	"strings"
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

func (m *model) startEditSelectedInput() {
	if m.screen == screenHistory {
		m.startEditLogEntryInput()
		return
	}
	m.screen = screenProgram
	switch m.currentLevel() {
	case screenPrograms:
		m.startEditProgramInput()
	case screenExercises:
		m.startEditExerciseInput()
	default:
		m.status = "edit is available for programs, exercises, and logs"
	}
}

func (m *model) startEditProgramInput() {
	program, ok := m.selectedProgram()
	if !ok {
		m.status = "select a program first"
		return
	}

	m.mode = modeInput
	m.inputPurpose = inputEditProgram
	m.editingProgram = program
	m.input.SetValue(program.ProgramName)
	m.input.Placeholder = "program name"
	m.input.Prompt = "edit program #" + strconv.FormatInt(program.ProgramId, 10) + " $ "
	m.status = helperMessage("edit program name", "enter save", "esc cancel")
}

func (m *model) submitEditedProgram(value string) {
	if m.editingProgram.ProgramId == 0 {
		m.status = "no program selected to edit"
		return
	}
	name := strings.TrimSpace(value)
	if name == "" {
		m.status = "program name is required"
		return
	}
	if err := m.store.UpdateProgram(m.editingProgram, name); err != nil {
		m.status = err.Error()
		return
	}

	updated := m.editingProgram
	updated.ProgramName = name
	m.editingProgram = data.Program{}
	for i := range m.programs {
		if m.programs[i].ProgramId == updated.ProgramId {
			m.programs[i] = updated
			m.programCursor = i
			break
		}
	}
	if m.activeProgram.ProgramId == updated.ProgramId {
		m.activeProgram = updated
	}
	m.status = "Updated program " + updated.ProgramName
}

func (m *model) startEditExerciseInput() {
	exercise, ok := m.selectedExercise()
	if !ok {
		m.status = "select an exercise first"
		return
	}

	m.mode = modeInput
	m.inputPurpose = inputEditExercise
	m.editingExercise = exercise
	m.input.SetValue(exerciseInputValue(exercise))
	m.input.Placeholder = "exercise name [sets] [reps]"
	m.input.Prompt = "edit exercise #" + strconv.FormatInt(exercise.ExerciseId, 10) + " $ "
	m.status = helperMessage("edit exercise name sets reps", "enter save", "esc cancel")
}

func (m *model) submitEditedExercise(value string) {
	if m.editingExercise.ExerciseId == 0 {
		m.status = "no exercise selected to edit"
		return
	}
	name, sets, reps, err := parseExerciseValue(strings.Fields(value))
	if err != nil {
		m.status = err.Error()
		return
	}
	if err := m.store.UpdateExercise(m.editingExercise, name, sets, reps); err != nil {
		m.status = err.Error()
		return
	}

	updated := m.editingExercise
	updated.Name = name
	updated.Sets = sets
	updated.Reps = reps
	m.editingExercise = data.Exercise{}
	for i := range m.exercises {
		if m.exercises[i].ExerciseId == updated.ExerciseId {
			m.exercises[i] = updated
			m.exerciseCursor = i
			break
		}
	}
	if m.activeExercise.ExerciseId == updated.ExerciseId {
		m.activeExercise = updated
	}
	m.status = "Updated exercise " + exerciseLabelForStatus(updated)
}

func (m *model) requestDeleteSelected() {
	if m.screen == screenHistory {
		m.requestDeleteLogEntry()
		return
	}
	m.screen = screenProgram
	switch m.currentLevel() {
	case screenPrograms:
		m.requestDeleteProgram()
	case screenExercises:
		m.requestDeleteExercise()
	default:
		m.status = "delete is available for programs, exercises, and logs"
	}
}

func (m *model) requestDeleteLogEntry() {
	entry, ok := m.selectedLogEntry()
	if !ok {
		m.status = "select a logged entry first"
		return
	}
	m.mode = modeDelete
	m.deleteTarget = deleteLog
	m.deletingEntry = entry
	m.inputPurpose = inputNone
	m.input.SetValue("")
	m.resetInputPrompt()
	m.status = m.deleteConfirmStatus()
}

func (m *model) requestDeleteProgram() {
	program, ok := m.selectedProgram()
	if !ok {
		m.status = "select a program first"
		return
	}
	m.mode = modeDelete
	m.deleteTarget = deleteProgram
	m.deletingProgram = program
	m.inputPurpose = inputNone
	m.input.SetValue("")
	m.resetInputPrompt()
	m.status = m.deleteConfirmStatus()
}

func (m *model) requestDeleteExercise() {
	exercise, ok := m.selectedExercise()
	if !ok {
		m.status = "select an exercise first"
		return
	}
	m.mode = modeDelete
	m.deleteTarget = deleteExercise
	m.deletingExercise = exercise
	m.inputPurpose = inputNone
	m.input.SetValue("")
	m.resetInputPrompt()
	m.status = m.deleteConfirmStatus()
}

func (m *model) confirmDeleteSelected() {
	switch m.deleteTarget {
	case deleteLog:
		m.confirmDeleteLogEntry()
	case deleteProgram:
		m.confirmDeleteProgram()
	case deleteExercise:
		m.confirmDeleteExercise()
	default:
		m.mode = modeNormal
		m.status = "nothing selected to delete"
	}
}

func (m *model) confirmDeleteLogEntry() {
	if m.deletingEntry.EntryId == 0 {
		m.mode = modeNormal
		m.clearDeleteState()
		m.status = "no log selected to delete"
		return
	}
	deleted := m.deletingEntry
	if err := m.store.DeleteGymSessionEntry(deleted); err != nil {
		m.mode = modeNormal
		m.clearDeleteState()
		m.status = err.Error()
		return
	}

	m.mode = modeNormal
	m.clearDeleteState()
	m.refreshHistoryAfterEntryDelete(deleted)
	m.status = "Deleted " + formatSessionEntry(deleted)
}

func (m *model) confirmDeleteProgram() {
	if m.deletingProgram.ProgramId == 0 {
		m.mode = modeNormal
		m.clearDeleteState()
		m.status = "no program selected to delete"
		return
	}
	deleted := m.deletingProgram
	if err := m.store.DeleteProgram(deleted); err != nil {
		m.mode = modeNormal
		m.clearDeleteState()
		m.status = err.Error()
		return
	}

	m.mode = modeNormal
	m.clearDeleteState()
	if err := m.loadPrograms(); err != nil {
		m.status = err.Error()
		return
	}
	if m.activeProgram.ProgramId == deleted.ProgramId {
		m.activeProgram = data.Program{}
		m.activeWorkout = data.Workout{}
		m.activeExercise = data.Exercise{}
		m.workouts = nil
		m.exercises = nil
		m.workoutCursor = 0
		m.exerciseCursor = 0
	}
	m.status = "Deleted program " + deleted.ProgramName
}

func (m *model) confirmDeleteExercise() {
	if m.deletingExercise.ExerciseId == 0 {
		m.mode = modeNormal
		m.clearDeleteState()
		m.status = "no exercise selected to delete"
		return
	}
	deleted := m.deletingExercise
	if err := m.store.DeleteExercise(deleted); err != nil {
		m.mode = modeNormal
		m.clearDeleteState()
		m.status = err.Error()
		return
	}

	m.mode = modeNormal
	m.clearDeleteState()
	if err := m.loadExercises(m.activeWorkout); err != nil {
		m.status = err.Error()
		return
	}
	if m.activeExercise.ExerciseId == deleted.ExerciseId {
		m.activeExercise = data.Exercise{}
	}
	m.status = "Deleted exercise " + deleted.Name
}

func (m *model) deleteConfirmStatus() string {
	switch m.deleteTarget {
	case deleteLog:
		return helperMessage("delete "+formatSessionEntry(m.deletingEntry)+"?", "y confirm", "n cancel")
	case deleteProgram:
		return helperMessage("delete program "+m.deletingProgram.ProgramName+"?", "y confirm", "n cancel")
	case deleteExercise:
		return helperMessage("delete exercise "+m.deletingExercise.Name+"?", "y confirm", "n cancel")
	default:
		return helperMessage("delete selected item?", "y confirm", "n cancel")
	}
}

func (m *model) clearDeleteState() {
	m.deleteTarget = deleteNone
	m.deletingEntry = data.GymSessionEntry{}
	m.deletingProgram = data.Program{}
	m.deletingExercise = data.Exercise{}
}

func (m *model) selectedLogEntry() (data.GymSessionEntry, bool) {
	if m.screen != screenHistory || len(m.historyEntries) == 0 {
		return data.GymSessionEntry{}, false
	}
	m.historyCursor = clampIndex(m.historyCursor, len(m.historyEntries))
	return m.historyEntries[m.historyCursor], true
}

func (m *model) selectedProgram() (data.Program, bool) {
	if len(m.programs) == 0 {
		return data.Program{}, false
	}
	m.programCursor = clampIndex(m.programCursor, len(m.programs))
	return m.programs[m.programCursor], true
}

func (m *model) selectedExercise() (data.Exercise, bool) {
	if m.activeWorkout.WorkoutId == 0 || len(m.exercises) == 0 {
		return data.Exercise{}, false
	}
	m.exerciseCursor = clampIndex(m.exerciseCursor, len(m.exercises))
	return m.exercises[m.exerciseCursor], true
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
