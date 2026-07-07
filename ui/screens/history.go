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

	lines = append(lines, renderWeightProgression(styles, entries)...)
	lines = append(lines, "")

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

func renderWeightProgression(styles theme.Styles, entries []data.GymSessionEntry) []string {
	points := weightedChartEntries(entries)
	if len(points) == 0 {
		return []string{styles.ProgramEmpty.Render("no weighted logs to graph yet")}
	}
	if len(points) > 8 {
		points = points[len(points)-8:]
	}

	contentW := max(24, styles.Box.GetWidth()-8)
	chartW := min(54, max(12, contentW-12))
	chartH := 5
	minWeight, maxWeight := weightRange(points)
	rows := make([][]rune, chartH)
	for i := range rows {
		rows[i] = []rune(strings.Repeat(" ", chartW))
	}

	previousX := 0
	previousY := pointY(points[0].Weight, minWeight, maxWeight, chartH)
	for i, point := range points {
		x := pointX(i, len(points), chartW)
		y := pointY(point.Weight, minWeight, maxWeight, chartH)
		if i > 0 {
			drawSegment(rows, previousX, previousY, x, y)
		}
		rows[y][x] = '●'
		previousX = x
		previousY = y
	}

	lines := []string{styles.ProgramPanelTitle.Render("weight progression")}
	for y, row := range rows {
		label := "       "
		switch y {
		case 0:
			label = fmt.Sprintf("%6.1f", maxWeight)
		case chartH - 1:
			label = fmt.Sprintf("%6.1f", minWeight)
		}
		lines = append(lines, styles.ProgramEmpty.Render(label+" │ ")+styles.ProgramItem.Render(string(row)))
	}
	lines = append(lines, styles.ProgramEmpty.Render("       └ "+strings.Repeat("─", chartW)))
	lines = append(lines, styles.ProgramEmpty.Render("dates  "+compactChartLabels(points, chartW, chartDateLabel)))
	lines = append(lines, styles.ProgramEmpty.Render("reps   "+compactChartLabels(points, chartW, chartRepLabel)))
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

func pointX(index int, length int, width int) int {
	if length <= 1 {
		return width / 2
	}
	return index * (width - 1) / (length - 1)
}

func pointY(weight float64, minWeight float64, maxWeight float64, height int) int {
	if maxWeight == minWeight {
		return height / 2
	}
	ratio := (maxWeight - weight) / (maxWeight - minWeight)
	y := int(ratio * float64(height-1))
	if y < 0 {
		return 0
	}
	if y >= height {
		return height - 1
	}
	return y
}

func drawSegment(rows [][]rune, x1 int, y1 int, x2 int, y2 int) {
	if x2 <= x1 {
		return
	}
	for x := x1 + 1; x < x2; x++ {
		t := float64(x-x1) / float64(x2-x1)
		y := y1 + int(float64(y2-y1)*t)
		if rows[y][x] == ' ' {
			rows[y][x] = '─'
		}
	}
}

func compactChartLabels(points []data.GymSessionEntry, width int, label func(data.GymSessionEntry) string) string {
	if len(points) == 0 {
		return ""
	}
	row := []rune(strings.Repeat(" ", width))
	for i, point := range points {
		value := []rune(label(point))
		x := pointX(i, len(points), width)
		if x+len(value) > width {
			x = max(0, width-len(value))
		}
		for j, char := range value {
			row[x+j] = char
		}
	}
	return strings.TrimRight(string(row), " ")
}

func chartDateLabel(entry data.GymSessionEntry) string {
	if len(entry.StartedAt) >= 10 {
		return entry.StartedAt[5:10]
	}
	return entry.StartedAt
}

func chartRepLabel(entry data.GymSessionEntry) string {
	return historySetRepLabel(entry)
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
