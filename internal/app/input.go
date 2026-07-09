package app

import (
	"ruffnut/data"
	"strings"
)

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
	case inputEditWorkout:
		m.submitEditedWorkout(value)
	case inputEditExercise:
		m.submitEditedExercise(value)
	default:
		m.status = "nothing to submit"
	}
}

func inputCancelledStatus(purpose inputPurpose) string {
	switch purpose {
	case inputEditLog, inputEditProgram, inputEditWorkout, inputEditExercise:
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
