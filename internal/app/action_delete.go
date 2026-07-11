package app

import "github.com/Yokanater/spotr/data"

func (m *model) requestDeleteSelected() {
	if m.screen == screenHistory {
		m.requestDeleteLogEntry()
		return
	}
	switch m.currentLevel() {
	case screenPrograms:
		m.requestDeleteProgram()
	case screenWorkouts:
		m.requestDeleteWorkout()
	case screenExercises:
		m.requestDeleteExercise()
	default:
		m.status = "delete is available for programs, workouts, exercises, and logs"
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

func (m *model) requestDeleteWorkout() {
	workout, ok := m.selectedWorkout()
	if !ok {
		m.status = "select a workout first"
		return
	}
	m.mode = modeDelete
	m.deleteTarget = deleteWorkout
	m.deletingWorkout = workout
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
	case deleteWorkout:
		m.confirmDeleteWorkout()
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
		if len(m.programs) > 0 {
			if err := m.restoreActiveProgram(); err != nil {
				m.status = err.Error()
				return
			}
		}
	}
	m.status = "Deleted program " + deleted.ProgramName
}

func (m *model) confirmDeleteWorkout() {
	if m.deletingWorkout.WorkoutId == 0 {
		m.mode = modeNormal
		m.clearDeleteState()
		m.status = "no workout selected to delete"
		return
	}
	deleted := m.deletingWorkout
	if err := m.store.DeleteWorkout(deleted); err != nil {
		m.mode = modeNormal
		m.clearDeleteState()
		m.status = err.Error()
		return
	}

	m.mode = modeNormal
	m.clearDeleteState()
	if err := m.loadWorkouts(m.activeProgram); err != nil {
		m.status = err.Error()
		return
	}
	if m.activeWorkout.WorkoutId == deleted.WorkoutId {
		m.activeWorkout = data.Workout{}
		m.activeExercise = data.Exercise{}
		m.exercises = nil
		m.exerciseCursor = 0
	}
	m.status = "Deleted workout " + deleted.Name
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
		return "Delete " + formatSessionEntry(m.deletingEntry) + "?"
	case deleteProgram:
		return "Delete program " + m.deletingProgram.ProgramName + "?"
	case deleteWorkout:
		return "Delete workout " + m.deletingWorkout.Name + "?"
	case deleteExercise:
		return "Delete exercise " + m.deletingExercise.Name + "?"
	default:
		return "Delete selected item?"
	}
}

func (m *model) clearDeleteState() {
	m.deleteTarget = deleteNone
	m.deletingEntry = data.GymSessionEntry{}
	m.deletingProgram = data.Program{}
	m.deletingWorkout = data.Workout{}
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

func (m *model) selectedWorkout() (data.Workout, bool) {
	if m.activeProgram.ProgramId == 0 || len(m.workouts) == 0 {
		return data.Workout{}, false
	}
	m.workoutCursor = clampIndex(m.workoutCursor, len(m.workouts))
	return m.workouts[m.workoutCursor], true
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
