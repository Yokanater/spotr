package app

import (
	"database/sql"
	"fmt"
	"spotr/data"
	"strconv"
	"strings"
)

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

	case "edit":
		if len(args) < 4 {
			m.status = "usage: log edit <entry-id> <sets> <reps> [weight] [notes]"
			return
		}
		entry, err := m.store.SelectGymSessionEntry(args[1], m.activeWorkout)
		if err != nil {
			m.status = err.Error()
			return
		}
		sets, reps, repsDetail, weight, notes, err := parseLoggedExerciseValue(strings.Join(args[2:], " "))
		if err != nil {
			m.status = err.Error()
			return
		}
		if err := m.store.UpdateGymSessionEntry(entry, sets, reps, repsDetail, weight, notes); err != nil {
			m.status = err.Error()
			return
		}
		entry.Sets = sets
		entry.Reps = reps
		entry.RepsDetail = repsDetail
		entry.Weight = weight
		entry.Notes = notes
		m.refreshHistoryAfterEntryUpdate(entry)
		m.status = "Updated " + formatSessionEntry(entry)

	case "delete":
		if len(args) < 2 {
			m.status = "usage: log delete <entry-id>"
			return
		}
		entry, err := m.store.SelectGymSessionEntry(args[1], m.activeWorkout)
		if err != nil {
			m.status = err.Error()
			return
		}
		if err := m.store.DeleteGymSessionEntry(entry); err != nil {
			m.status = err.Error()
			return
		}
		m.refreshHistoryAfterEntryDelete(entry)
		m.status = "Deleted " + formatSessionEntry(entry)

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
		m.status = "Recent sessions"

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
		m.status = "Session details"

	default:
		m.status = fmt.Sprintf("unknown history command: %s", args[0])
	}
}
