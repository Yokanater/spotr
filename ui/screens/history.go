package screens

import (
	"fmt"
	"ruffnut/data"
	"ruffnut/ui/theme"
	"strings"

	"charm.land/lipgloss/v2"
)

func HistoryView(styles theme.Styles, workout data.Workout, titleText string, sessions []data.GymSession, cursor int, activeSession data.GymSession, entries []data.GymSessionEntry) string {
	title := styles.ProgramTitle.Render("logs")
	subtitle := styles.ProgramSubtitle.Render(historySubtitle(workout, titleText, activeSession))
	body := renderHistoryBody(styles, sessions, cursor, activeSession, entries)

	return lipgloss.JoinVertical(lipgloss.Left, RenderHeader(styles, "logs"), "", title, subtitle, "", body)
}

func historySubtitle(workout data.Workout, titleText string, activeSession data.GymSession) string {
	if activeSession.SessionId != 0 {
		if titleText != "" {
			return fmt.Sprintf("%s / session ID #%d", titleText, activeSession.SessionId)
		}
		return fmt.Sprintf("%s / session ID #%d", workout.Name, activeSession.SessionId)
	}
	if titleText != "" {
		return titleText
	}
	if workout.WorkoutId != 0 {
		return workout.Name + " / recent sessions"
	}
	return "recent sessions"
}

func renderHistoryBody(styles theme.Styles, sessions []data.GymSession, cursor int, activeSession data.GymSession, entries []data.GymSessionEntry) string {
	if activeSession.SessionId != 0 {
		return renderSessionDetail(styles, activeSession, entries, cursor)
	}
	if entries != nil {
		return renderExerciseHistory(styles, entries, cursor)
	}
	return renderSessionList(styles, sessions, cursor)
}

func renderSessionList(styles theme.Styles, sessions []data.GymSession, cursor int) string {
	lines := []string{styles.ProgramPanelTitle.Render("recent")}
	lines = append(lines, "")
	if len(sessions) == 0 {
		lines = append(lines, styles.ProgramEmpty.Render("no logs yet"))
		return styles.ProgramPanel.Width(styles.Box.GetWidth()).Render(strings.Join(lines, "\n"))
	}

	start, end := visibleRange(len(sessions), cursor, styles.ProgramListRows)
	if start > 0 {
		lines = append(lines, styles.ProgramEmpty.Render("  ..."))
	}
	for i := start; i < end; i++ {
		session := sessions[i]
		state := "active"
		if session.EndedAt != "" {
			state = "done"
		}
		rowStyle := styles.ProgramItem
		marker := " "
		if i == cursor {
			rowStyle = styles.ProgramSelected
			marker = ">"
		}
		id := styles.HelperKey.Render(fmt.Sprintf("ID #%d", session.SessionId))
		lines = append(lines, rowStyle.Render(fmt.Sprintf("%s %s  %s  %s", marker, id, state, session.StartedAt)))
		if session.Notes != "" {
			lines = append(lines, styles.ProgramEmpty.Render("  "+session.Notes))
		}
	}
	if end < len(sessions) {
		lines = append(lines, styles.ProgramEmpty.Render("  ..."))
	}
	return styles.ProgramPanel.Width(styles.Box.GetWidth()).Render(strings.Join(lines, "\n"))
}

func renderSessionDetail(styles theme.Styles, session data.GymSession, entries []data.GymSessionEntry, cursor int) string {
	lines := []string{styles.ProgramPanelTitle.Render("session " + styles.HelperKey.Render(fmt.Sprintf("ID #%d", session.SessionId)))}
	lines = append(lines, styles.ProgramEmpty.Render(sessionStateLine(session)), "")
	if len(entries) == 0 {
		lines = append(lines, styles.ProgramEmpty.Render("no entries yet"))
		return styles.ProgramPanel.Width(styles.Box.GetWidth()).Render(strings.Join(lines, "\n"))
	}

	start, end := visibleRange(len(entries), cursor, styles.ProgramListRows)
	if start > 0 {
		lines = append(lines, styles.ProgramEmpty.Render("  ..."))
	}
	for i := start; i < end; i++ {
		entry := entries[i]
		if i > start {
			lines = append(lines, "")
		}
		rowStyle := styles.ProgramItem
		marker := " "
		if i == cursor {
			rowStyle = styles.ProgramSelected
			marker = ">"
		}
		lines = append(lines, rowStyle.Render(marker+" "+historyEntryLine(entry)))
		if entry.Notes != "" {
			lines = append(lines, styles.ProgramEmpty.Render("  "+entry.Notes))
		}
	}
	if end < len(entries) {
		lines = append(lines, styles.ProgramEmpty.Render("  ..."))
	}
	return styles.ProgramPanel.Width(styles.Box.GetWidth()).Render(strings.Join(lines, "\n"))
}

func renderExerciseHistory(styles theme.Styles, entries []data.GymSessionEntry, cursor int) string {
	lines := []string{styles.ProgramPanelTitle.Render("movement history")}
	lines = append(lines, "")
	if len(entries) == 0 {
		lines = append(lines, styles.ProgramEmpty.Render("no linked logs yet"))
		return styles.ProgramPanel.Width(styles.Box.GetWidth()).Render(strings.Join(lines, "\n"))
	}

	start, end := visibleRange(len(entries), cursor, styles.ProgramListRows)
	if start > 0 {
		lines = append(lines, styles.ProgramEmpty.Render("  ..."))
	}
	for i := start; i < end; i++ {
		entry := entries[i]
		if i > start {
			lines = append(lines, "")
		}
		rowStyle := styles.ProgramItem
		marker := " "
		if i == cursor {
			rowStyle = styles.ProgramSelected
			marker = ">"
		}
		id := styles.HelperKey.Render(fmt.Sprintf("ID #%d", entry.SessionId))
		lines = append(lines, rowStyle.Render(fmt.Sprintf("%s %s  %s  %s", marker, id, entry.StartedAt, historyEntryLine(entry))))
		if entry.Workout != "" {
			lines = append(lines, styles.ProgramEmpty.Render("  "+entry.Workout))
		}
		if entry.Notes != "" {
			lines = append(lines, styles.ProgramEmpty.Render("  "+entry.Notes))
		}
	}
	if end < len(entries) {
		lines = append(lines, styles.ProgramEmpty.Render("  ..."))
	}
	return styles.ProgramPanel.Width(styles.Box.GetWidth()).Render(strings.Join(lines, "\n"))
}

func sessionStateLine(session data.GymSession) string {
	if session.EndedAt == "" {
		return "active / started " + session.StartedAt
	}
	return "done / started " + session.StartedAt
}

func historyEntryLine(entry data.GymSessionEntry) string {
	line := fmt.Sprintf("%s  %s", entry.Exercise, historySetRepLabel(entry))
	if entry.Weight > 0 {
		line += fmt.Sprintf("  @ %.1f", entry.Weight)
	}
	return line
}

func historySetRepLabel(entry data.GymSessionEntry) string {
	if entry.RepsDetail != "" {
		return fmt.Sprintf("%dx%s", entry.Sets, entry.RepsDetail)
	}
	return fmt.Sprintf("%dx%d", entry.Sets, entry.Reps)
}
