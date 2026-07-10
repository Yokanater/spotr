package screens

import (
	"fmt"
	"spotr/data"
	"spotr/ui/theme"
	"strings"

	"charm.land/lipgloss/v2"
)

func ProgramView(styles theme.Styles, programs []data.Program, workouts []data.Workout, exercises []data.Exercise, activeProgram data.Program, activeWorkout data.Workout, activeExercise data.Exercise, programCursor int, workoutCursor int, exerciseCursor int, choosingProgram bool) string {
	if choosingProgram || activeProgram.ProgramId == 0 {
		heading := renderTrainingHeading(styles, "programs", "choose once — Spotr remembers")
		programPanel := renderProgramSection(styles, "programs", programNames(programs), "press a to create one or t to use a template", programCursor)
		return lipgloss.JoinVertical(lipgloss.Left, RenderHeader(styles, "workouts"), "", heading, "", programPanel)
	}

	context := activeProgram.ProgramName
	if activeWorkout.WorkoutId != 0 {
		context += " / " + activeWorkout.Name
	}
	heading := renderTrainingHeading(styles, "workouts", context)
	workoutPanel := renderProgramSection(styles, "workouts", workoutNames(workouts), "press a to add your first workout", workoutCursor)
	exercisePanel := renderExerciseSection(styles, "exercises", exercises, "press a to add your first exercise", exerciseCursor)

	panels := visibleTrainingPanels(styles, activeWorkout, workoutPanel, exercisePanel)
	panelBlock := lipgloss.JoinHorizontal(lipgloss.Top, panels...)
	if styles.ProgramTitle.GetWidth() < 82 {
		panelBlock = lipgloss.JoinVertical(lipgloss.Left, panels...)
	}
	if styles.ProgramListRows <= 1 {
		return lipgloss.JoinVertical(lipgloss.Left, RenderHeader(styles, "workouts"), "", panelBlock)
	}
	return lipgloss.JoinVertical(lipgloss.Left, RenderHeader(styles, "workouts"), "", heading, "", panelBlock)
}

func renderTrainingHeading(styles theme.Styles, title string, context string) string {
	width := styles.ProgramTitle.GetWidth()
	left := styles.ProgramTitle.Width(0).Render(title)
	right := styles.ProgramSubtitle.Width(0).Render(context)
	space := width - lipgloss.Width(left) - lipgloss.Width(right)
	if space < 2 {
		return lipgloss.JoinVertical(lipgloss.Left, left, right)
	}
	return lipgloss.NewStyle().Width(width).Render(lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", space), right))
}

func visibleTrainingPanels(styles theme.Styles, activeWorkout data.Workout, workoutPanel string, exercisePanel string) []string {
	if styles.ProgramTitle.GetWidth() < 82 {
		if activeWorkout.WorkoutId != 0 {
			return []string{exercisePanel}
		}
		return []string{workoutPanel}
	}

	if activeWorkout.WorkoutId != 0 {
		return []string{workoutPanel, exercisePanel}
	}
	return []string{workoutPanel}
}

func renderExerciseSection(styles theme.Styles, title string, exercises []data.Exercise, emptyHint string, cursor int) string {
	var lines []string
	lines = append(lines, styles.ProgramPanelTitle.Render(title))
	if styles.ProgramListRows > 1 {
		lines = append(lines, "")
	}
	if len(exercises) == 0 {
		lines = append(lines, styles.ProgramEmpty.Render(emptyHint))
		return styles.ProgramPanel.Render(strings.Join(lines, "\n"))
	}

	contentW := max(12, styles.ProgramPanel.GetWidth()-6)
	targetW := 6
	nameW := max(4, contentW-targetW-1)
	start, end := visibleRange(len(exercises), cursor, styles.ProgramListRows)
	if start > 0 {
		lines = append(lines, styles.ProgramEmpty.Render("  ..."))
	}

	for i := start; i < end; i++ {
		exercise := exercises[i]
		marker := " "
		rowStyle := styles.ProgramItem
		if i == cursor {
			marker = ">"
			rowStyle = styles.ProgramSelected
		}

		name := lipgloss.NewStyle().
			Width(nameW).
			MaxWidth(nameW).
			Render(rowStyle.Render(marker+" ") + fmt.Sprintf("#%d  %s", exercise.ExerciseId, exercise.Name))
		target := lipgloss.NewStyle().
			Width(targetW).
			Align(lipgloss.Right).
			Render(rowStyle.Render(exerciseTarget(exercise)))

		lines = append(lines, rowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, name, " ", target)))
	}
	if end < len(exercises) {
		lines = append(lines, styles.ProgramEmpty.Render("  ..."))
	}

	return styles.ProgramPanel.Render(strings.Join(lines, "\n"))
}

func renderProgramSection(styles theme.Styles, title string, values []string, emptyHint string, cursor int) string {
	var lines []string
	lines = append(lines, styles.ProgramPanelTitle.Render(title))
	if styles.ProgramListRows > 1 {
		lines = append(lines, "")
	}
	if len(values) == 0 {
		lines = append(lines, styles.ProgramEmpty.Render(emptyHint))
		return styles.ProgramPanel.Render(strings.Join(lines, "\n"))
	}

	start, end := visibleRange(len(values), cursor, styles.ProgramListRows)
	if start > 0 {
		lines = append(lines, styles.ProgramEmpty.Render("  ..."))
	}

	for i := start; i < end; i++ {
		value := values[i]
		marker := " "
		rowStyle := styles.ProgramItem
		if i == cursor {
			marker = ">"
			rowStyle = styles.ProgramSelected
		}
		lines = append(lines, rowStyle.Render(marker+" "+value))
	}
	if end < len(values) {
		lines = append(lines, styles.ProgramEmpty.Render("  ..."))
	}

	return styles.ProgramPanel.Render(strings.Join(lines, "\n"))
}

func visibleRange(length int, cursor int, rows int) (int, int) {
	if length == 0 {
		return 0, 0
	}
	if rows <= 0 || rows >= length {
		return 0, length
	}
	cursor = min(max(0, cursor), length-1)
	start := cursor - rows/2
	if start < 0 {
		start = 0
	}
	if start+rows > length {
		start = length - rows
	}
	return start, start + rows
}

func programNames(programs []data.Program) []string {
	names := make([]string, 0, len(programs))
	for _, program := range programs {
		names = append(names, fmt.Sprintf("#%d  %s", program.ProgramId, program.ProgramName))
	}
	return names
}

func workoutNames(workouts []data.Workout) []string {
	names := make([]string, 0, len(workouts))
	for _, workout := range workouts {
		names = append(names, fmt.Sprintf("#%d  %s", workout.WorkoutId, workout.Name))
	}
	return names
}

func exerciseLabel(exercise data.Exercise) string {
	if exercise.Sets > 0 || exercise.Reps > 0 {
		return fmt.Sprintf("#%d  %s  %dx%d", exercise.ExerciseId, exercise.Name, exercise.Sets, exercise.Reps)
	}
	return fmt.Sprintf("#%d  %s", exercise.ExerciseId, exercise.Name)
}

func exerciseTarget(exercise data.Exercise) string {
	if exercise.Sets > 0 || exercise.Reps > 0 {
		return fmt.Sprintf("%dx%d", exercise.Sets, exercise.Reps)
	}
	return ""
}
