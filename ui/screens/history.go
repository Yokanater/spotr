package screens

import (
	"fmt"
	"ruffnut/data"
	"ruffnut/ui/theme"
	"sort"
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

	rows := max(1, styles.ProgramListRows-2)
	start, end := visibleRange(len(entries), cursor, rows)
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

	progressLines := renderProgressSummary(styles, entries)
	lines = append(lines, progressLines...)
	lines = append(lines, "")

	rows := max(1, styles.ProgramListRows-len(progressLines)-1)
	start, end := visibleRange(len(entries), cursor, rows)
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
		lines = append(lines, rowStyle.Render(movementHistoryLine(styles, marker, entry)))
		detail := movementHistoryDetail(entry)
		if detail != "" {
			lines = append(lines, styles.ProgramEmpty.Render("  "+detail))
		}
	}
	if end < len(entries) {
		lines = append(lines, styles.ProgramEmpty.Render("  ..."))
	}
	return styles.ProgramPanel.Width(styles.Box.GetWidth()).Render(strings.Join(lines, "\n"))
}

func renderProgressSummary(styles theme.Styles, entries []data.GymSessionEntry) []string {
	points := weightedChartEntries(entries)
	if len(points) == 0 {
		return []string{styles.ProgramEmpty.Render("no weighted logs to graph yet")}
	}
	if len(points) > 8 {
		points = points[len(points)-8:]
	}

	contentW := max(24, styles.Box.GetWidth()-8)
	_, maxWeight := weightRange(points)
	first := points[0]
	last := points[len(points)-1]
	delta := last.Weight - first.Weight
	summary := fmt.Sprintf("best %.1f   latest %.1f   change %+0.1f", maxWeight, last.Weight, delta)
	if contentW < 48 {
		summary = fmt.Sprintf("best %.1f  latest %.1f", maxWeight, last.Weight)
	}

	lines := []string{styles.ProgramPanelTitle.Render("progress")}
	lines = append(lines, styles.ProgramEmpty.Render(summary))
	lines = append(lines, styles.ProgramItem.Render(weightSparkline(points)))
	lines = append(lines, styles.ProgramEmpty.Render(progressChips(points, contentW)))
	return lines
}

func weightedChartEntries(entries []data.GymSessionEntry) []data.GymSessionEntry {
	points := make([]data.GymSessionEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.Weight > 0 {
			points = append(points, entry)
		}
	}
	sort.SliceStable(points, func(i int, j int) bool {
		if points[i].StartedAt == points[j].StartedAt {
			return points[i].EntryId < points[j].EntryId
		}
		return points[i].StartedAt < points[j].StartedAt
	})
	return points
}

func weightRange(points []data.GymSessionEntry) (float64, float64) {
	minWeight := points[0].Weight
	maxWeight := points[0].Weight
	for _, point := range points[1:] {
		if point.Weight < minWeight {
			minWeight = point.Weight
		}
		if point.Weight > maxWeight {
			maxWeight = point.Weight
		}
	}
	return minWeight, maxWeight
}

func weightSparkline(points []data.GymSessionEntry) string {
	minWeight, maxWeight := weightRange(points)
	levels := []rune("▁▂▃▄▅▆▇█")
	values := make([]rune, 0, len(points))
	for _, point := range points {
		level := 0
		if maxWeight > minWeight {
			level = int((point.Weight - minWeight) / (maxWeight - minWeight) * float64(len(levels)-1))
		}
		if level < 0 {
			level = 0
		}
		if level >= len(levels) {
			level = len(levels) - 1
		}
		values = append(values, levels[level])
	}
	return strings.Join(strings.Split(string(values), ""), " ")
}

func progressChips(points []data.GymSessionEntry, width int) string {
	chips := make([]string, 0, len(points))
	for _, point := range points {
		chips = append(chips, fmt.Sprintf("%s %.1f %s", shortDate(point.StartedAt), point.Weight, historySetRepLabel(point)))
	}
	row := strings.Join(chips, "  ·  ")
	if lipgloss.Width(row) <= width {
		return row
	}
	first := points[0]
	last := points[len(points)-1]
	return fmt.Sprintf("%s %.1f %s  →  %s %.1f %s", shortDate(first.StartedAt), first.Weight, historySetRepLabel(first), shortDate(last.StartedAt), last.Weight, historySetRepLabel(last))
}

func shortDate(value string) string {
	if len(value) >= 10 {
		return value[5:10]
	}
	return value
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

func movementHistoryLine(styles theme.Styles, marker string, entry data.GymSessionEntry) string {
	id := styles.HelperKey.Render(fmt.Sprintf("#%d", entry.SessionId))
	return fmt.Sprintf("%s %s  %s  %s", marker, id, shortDate(entry.StartedAt), historyEntryLine(entry))
}

func movementHistoryDetail(entry data.GymSessionEntry) string {
	parts := []string{}
	if entry.Workout != "" {
		parts = append(parts, entry.Workout)
	}
	if entry.Notes != "" {
		parts = append(parts, entry.Notes)
	}
	return strings.Join(parts, "  /  ")
}

func historySetRepLabel(entry data.GymSessionEntry) string {
	if entry.RepsDetail != "" {
		return fmt.Sprintf("%dx%s", entry.Sets, entry.RepsDetail)
	}
	return fmt.Sprintf("%dx%d", entry.Sets, entry.Reps)
}
