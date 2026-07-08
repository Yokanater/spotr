package app

import (
	tea "charm.land/bubbletea/v2"
	"database/sql"
	"fmt"
	"ruffnut/commands"
	"ruffnut/data"
	"strconv"
	"strings"
)

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
