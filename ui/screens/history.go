package screens

import (
	"fmt"
	"math"
	"sort"
	"spotr/data"
	"spotr/ui/theme"
	"strconv"
	"strings"
	"time"

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
	if len(points) > 12 {
		points = points[len(points)-12:]
	}

	contentW := max(24, styles.Box.GetWidth()-8)
	first := points[0]
	last := points[len(points)-1]
	firstStrength := estimatedOneRepMax(first)
	lastStrength := estimatedOneRepMax(last)
	bestStrength := firstStrength
	bestVolume := exerciseVolume(first)
	for _, point := range points[1:] {
		bestStrength = max(bestStrength, estimatedOneRepMax(point))
		bestVolume = max(bestVolume, exerciseVolume(point))
	}
	strengthSummary := fmt.Sprintf("latest %.1f   best %.1f   change %+0.1f", lastStrength, bestStrength, lastStrength-firstStrength)
	volumeSummary := fmt.Sprintf("volume latest %.0f   best %.0f", exerciseVolume(last), bestVolume)

	lines := []string{styles.ProgramPanelTitle.Render("strength · estimated 1RM")}
	lines = append(lines, styles.ProgramEmpty.Render(strengthSummary))
	lines = append(lines, styles.ProgramItem.Render(strings.Join(strengthChart(points, contentW), "\n")))
	lines = append(lines, styles.ProgramEmpty.Render(volumeSummary))
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

func estimatedOneRepMax(entry data.GymSessionEntry) float64 {
	return entry.Weight * (1 + float64(maxEffortReps(entry))/30)
}

func maxEffortReps(entry data.GymSessionEntry) int {
	maximum := entry.Reps
	for _, value := range strings.Split(entry.RepsDetail, "/") {
		reps, err := strconv.Atoi(value)
		if err == nil && reps > maximum {
			maximum = reps
		}
	}
	return maximum
}

func exerciseVolume(entry data.GymSessionEntry) float64 {
	totalReps := entry.Sets * entry.Reps
	if entry.RepsDetail != "" {
		totalReps = 0
		for _, value := range strings.Split(entry.RepsDetail, "/") {
			reps, err := strconv.Atoi(value)
			if err == nil {
				totalReps += reps
			}
		}
	}
	return entry.Weight * float64(totalReps)
}

func strengthChart(points []data.GymSessionEntry, width int) []string {
	if len(points) == 1 {
		value := estimatedOneRepMax(points[0])
		return []string{
			fmt.Sprintf("%6.1f │█", value),
			"       " + shortDate(points[0].StartedAt),
		}
	}
	const height = 3
	plotWidth := min(36, max(10, width-9))
	columns := make([]float64, plotWidth)
	for i := range columns {
		columns[i] = math.NaN()
	}
	positions := chartPositions(points, plotWidth)
	minimum := estimatedOneRepMax(points[0])
	maximum := minimum
	for i, point := range points {
		value := estimatedOneRepMax(point)
		minimum = min(minimum, value)
		maximum = max(maximum, value)
		position := positions[i]
		if math.IsNaN(columns[position]) || value > columns[position] {
			columns[position] = value
		}
	}

	levels := []rune(" ▁▂▃▄▅▆▇█")
	lines := make([]string, 0, height+1)
	for row := 0; row < height; row++ {
		label := "       │"
		if row == 0 {
			label = fmt.Sprintf("%6.1f │", maximum)
		} else if row == height-1 {
			label = fmt.Sprintf("%6.1f │", minimum)
		}
		var bars strings.Builder
		for _, value := range columns {
			if math.IsNaN(value) {
				bars.WriteByte(' ')
				continue
			}
			scaled := 1
			if maximum > minimum {
				scaled += int((value - minimum) / (maximum - minimum) * float64(height*8-1))
			} else {
				scaled = height * 4
			}
			remaining := scaled - (height-1-row)*8
			remaining = min(8, max(0, remaining))
			bars.WriteRune(levels[remaining])
		}
		lines = append(lines, label+bars.String())
	}
	firstDate := shortDate(points[0].StartedAt)
	lastDate := shortDate(points[len(points)-1].StartedAt)
	dateGap := max(1, plotWidth-len(firstDate)-len(lastDate))
	lines = append(lines, "       "+firstDate+strings.Repeat(" ", dateGap)+lastDate)
	return lines
}

func chartPositions(points []data.GymSessionEntry, width int) []int {
	positions := make([]int, len(points))
	if len(points) <= 1 {
		return positions
	}
	first, firstErr := time.Parse(time.RFC3339, points[0].StartedAt)
	last, lastErr := time.Parse(time.RFC3339, points[len(points)-1].StartedAt)
	span := last.Sub(first)
	for i, point := range points {
		position := i * (width - 1) / (len(points) - 1)
		if firstErr == nil && lastErr == nil && span > 0 {
			if timestamp, err := time.Parse(time.RFC3339, point.StartedAt); err == nil {
				position = int(float64(timestamp.Sub(first)) / float64(span) * float64(width-1))
			}
		}
		positions[i] = min(width-1, max(0, position))
	}
	return positions
}

func progressChips(points []data.GymSessionEntry, width int) string {
	chips := make([]string, 0, len(points))
	for _, point := range points {
		chips = append(chips, fmt.Sprintf("%s load %.1f %s", shortDate(point.StartedAt), point.Weight, historySetRepLabel(point)))
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
