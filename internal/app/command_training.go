package app

import (
	"fmt"
	"spotr/data"
	"strings"
)

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

	case "edit":
		if m.activeProgram.ProgramId == 0 {
			m.status = "select a program first: program select <id|name>"
			return
		}
		if len(args) < 2 {
			m.status = "usage: program edit <name>"
			return
		}
		name := strings.Join(args[1:], " ")
		if err := m.store.UpdateProgram(m.activeProgram, name); err != nil {
			m.status = err.Error()
			return
		}
		m.activeProgram.ProgramName = name
		if err := m.loadPrograms(); err != nil {
			m.status = err.Error()
			return
		}
		for i := range m.programs {
			if m.programs[i].ProgramId == m.activeProgram.ProgramId {
				m.programCursor = i
				break
			}
		}
		m.status = "Updated program " + name

	case "delete":
		if len(args) < 2 {
			m.status = "usage: program delete <id|name>"
			return
		}
		program, err := m.store.SelectProgram(strings.Join(args[1:], " "))
		if err != nil {
			m.status = err.Error()
			return
		}
		if err := m.store.DeleteProgram(program); err != nil {
			m.status = err.Error()
			return
		}
		if err := m.loadPrograms(); err != nil {
			m.status = err.Error()
			return
		}
		if m.activeProgram.ProgramId == program.ProgramId {
			m.activeProgram = data.Program{}
			m.activeWorkout = data.Workout{}
			m.activeExercise = data.Exercise{}
			m.workouts = nil
			m.exercises = nil
			m.workoutCursor = 0
			m.exerciseCursor = 0
		}
		m.status = "Deleted program " + program.ProgramName

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

	case "edit":
		if m.activeWorkout.WorkoutId == 0 {
			m.status = "select a workout first: workout select <id|name>"
			return
		}
		if len(args) < 2 {
			m.status = "usage: workout edit <name>"
			return
		}
		name := strings.Join(args[1:], " ")
		if err := m.store.UpdateWorkout(m.activeWorkout, name); err != nil {
			m.status = err.Error()
			return
		}
		m.activeWorkout.Name = name
		if err := m.loadWorkouts(m.activeProgram); err != nil {
			m.status = err.Error()
			return
		}
		for i := range m.workouts {
			if m.workouts[i].WorkoutId == m.activeWorkout.WorkoutId {
				m.workoutCursor = i
				break
			}
		}
		m.status = "Updated workout " + name

	case "delete":
		if len(args) < 2 {
			m.status = "usage: workout delete <id|name>"
			return
		}
		workout, err := m.store.SelectWorkout(strings.Join(args[1:], " "), m.activeProgram)
		if err != nil {
			m.status = err.Error()
			return
		}
		if err := m.store.DeleteWorkout(workout); err != nil {
			m.status = err.Error()
			return
		}
		if err := m.loadWorkouts(m.activeProgram); err != nil {
			m.status = err.Error()
			return
		}
		if m.activeWorkout.WorkoutId == workout.WorkoutId {
			m.activeWorkout = data.Workout{}
			m.activeExercise = data.Exercise{}
			m.exercises = nil
			m.exerciseCursor = 0
		}
		m.status = "Deleted workout " + workout.Name
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

	case "edit":
		if m.activeExercise.ExerciseId == 0 {
			m.status = "select an exercise first: exercise select <id|name>"
			return
		}
		if len(args) < 2 {
			m.status = "usage: exercise edit <name> [sets] [reps]"
			return
		}
		name, sets, reps, err := parseExerciseValue(args[1:])
		if err != nil {
			m.status = err.Error()
			return
		}
		if err := m.store.UpdateExercise(m.activeExercise, name, sets, reps); err != nil {
			m.status = err.Error()
			return
		}
		m.activeExercise.Name = name
		m.activeExercise.Sets = sets
		m.activeExercise.Reps = reps
		if err := m.loadExercises(m.activeWorkout); err != nil {
			m.status = err.Error()
			return
		}
		for i := range m.exercises {
			if m.exercises[i].ExerciseId == m.activeExercise.ExerciseId {
				m.exerciseCursor = i
				break
			}
		}
		m.status = "Updated exercise " + exerciseLabelForStatus(m.activeExercise)

	case "delete":
		if len(args) < 2 {
			m.status = "usage: exercise delete <id|name>"
			return
		}
		exercise, err := m.store.SelectExercise(strings.Join(args[1:], " "), m.activeWorkout)
		if err != nil {
			m.status = err.Error()
			return
		}
		if err := m.store.DeleteExercise(exercise); err != nil {
			m.status = err.Error()
			return
		}
		if err := m.loadExercises(m.activeWorkout); err != nil {
			m.status = err.Error()
			return
		}
		if m.activeExercise.ExerciseId == exercise.ExerciseId {
			m.activeExercise = data.Exercise{}
		}
		m.status = "Deleted exercise " + exercise.Name

	default:
		m.status = fmt.Sprintf("unknown exercise command: %s", cmd)
	}
}
