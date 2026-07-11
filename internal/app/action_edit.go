package app

import (
	"github.com/Yokanater/spotr/data"
	"strconv"
	"strings"
)

func (m *model) startEditLogEntryInput() {
	entry, ok := m.selectedLogEntry()
	if !ok {
		m.status = "select a logged entry first"
		return
	}

	m.mode = modeInput
	m.input.Focus()
	m.inputPurpose = inputEditLog
	m.editingEntry = entry
	m.input.SetValue(logEntryInputValue(entry))
	m.input.Placeholder = "sets reps or reps/reps [weight] [notes]"
	m.input.Prompt = "edit log #" + strconv.FormatInt(entry.EntryId, 10) + " $ "
	m.status = "Sets, reps, load, then optional notes"
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
	switch m.currentLevel() {
	case screenPrograms:
		m.startEditProgramInput()
	case screenWorkouts:
		m.startEditWorkoutInput()
	case screenExercises:
		m.startEditExerciseInput()
	default:
		m.status = "edit is available for programs, workouts, exercises, and logs"
	}
}

func (m *model) startEditProgramInput() {
	program, ok := m.selectedProgram()
	if !ok {
		m.status = "select a program first"
		return
	}

	m.mode = modeInput
	m.input.Focus()
	m.inputPurpose = inputEditProgram
	m.editingProgram = program
	m.input.SetValue(program.ProgramName)
	m.input.Placeholder = "program name"
	m.input.Prompt = "edit program #" + strconv.FormatInt(program.ProgramId, 10) + " $ "
	m.status = "Rename program"
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

func (m *model) startEditWorkoutInput() {
	workout, ok := m.selectedWorkout()
	if !ok {
		m.status = "select a workout first"
		return
	}

	m.mode = modeInput
	m.input.Focus()
	m.inputPurpose = inputEditWorkout
	m.editingWorkout = workout
	m.input.SetValue(workout.Name)
	m.input.Placeholder = "workout name"
	m.input.Prompt = "edit workout #" + strconv.FormatInt(workout.WorkoutId, 10) + " $ "
	m.status = "Rename workout"
}

func (m *model) submitEditedWorkout(value string) {
	if m.editingWorkout.WorkoutId == 0 {
		m.status = "no workout selected to edit"
		return
	}
	name := strings.TrimSpace(value)
	if name == "" {
		m.status = "workout name is required"
		return
	}
	if err := m.store.UpdateWorkout(m.editingWorkout, name); err != nil {
		m.status = err.Error()
		return
	}

	updated := m.editingWorkout
	updated.Name = name
	m.editingWorkout = data.Workout{}
	for i := range m.workouts {
		if m.workouts[i].WorkoutId == updated.WorkoutId {
			m.workouts[i] = updated
			m.workoutCursor = i
			break
		}
	}
	if m.activeWorkout.WorkoutId == updated.WorkoutId {
		m.activeWorkout = updated
	}
	m.status = "Updated workout " + updated.Name
}

func (m *model) startEditExerciseInput() {
	exercise, ok := m.selectedExercise()
	if !ok {
		m.status = "select an exercise first"
		return
	}

	m.mode = modeInput
	m.input.Focus()
	m.inputPurpose = inputEditExercise
	m.editingExercise = exercise
	m.input.SetValue(exerciseInputValue(exercise))
	m.input.Placeholder = "exercise name [sets] [reps]"
	m.input.Prompt = "edit exercise #" + strconv.FormatInt(exercise.ExerciseId, 10) + " $ "
	m.status = "Exercise name, sets, and reps"
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
