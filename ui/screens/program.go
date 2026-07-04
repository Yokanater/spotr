package screens

import (
	"fmt"
	"ruffnut/data"
	"ruffnut/ui/theme"
	"strings"

	"charm.land/lipgloss/v2"
)

func ProgramView(styles theme.Styles, programs []data.Program, workouts []data.Workout, exercises []data.Exercise, activeProgram data.Program, activeWorkout data.Workout, activeExercise data.Exercise, programCursor int, workoutCursor int, exerciseCursor int) string {
	title := styles.ProgramTitle.Render("training desk")

	var context []string
	if activeProgram.ProgramId == 0 {
		context = append(context, "program none")
	} else {
		context = append(context, "program "+activeProgram.ProgramName)
	}
	if activeWorkout.WorkoutId == 0 {
		context = append(context, "workout none")
	} else {
		context = append(context, "workout "+activeWorkout.Name)
	}
	if activeExercise.ExerciseId == 0 {
		context = append(context, "exercise none")
	} else {
		context = append(context, "exercise "+exerciseLabel(activeExercise))
	}

	subtitle := styles.ProgramSubtitle.Render(strings.Join(context, " / "))
	programPanel := renderProgramSection(styles, "01 programs", programNames(programs), "program list", programCursor)
	workoutPanel := renderProgramSection(styles, "02 workouts", workoutNames(workouts), "workout list", workoutCursor)
	exercisePanel := renderExerciseSection(styles, "03 exercises", exercises, "exercise list", exerciseCursor)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, programPanel, workoutPanel, exercisePanel)
	if styles.ProgramTitle.GetWidth() < 82 {
		panels = lipgloss.JoinVertical(lipgloss.Left, orderedPanels(activeProgram, activeWorkout, programPanel, workoutPanel, exercisePanel)...)
	}
	footer := styles.ProgramSubtitle.Render("program add <name>  /  workout add <name>  /  exercise add <name> [sets] [reps]")
	return lipgloss.JoinVertical(lipgloss.Left, RenderHeader(styles, "program"), "", title, subtitle, "", panels, "", footer)
}

func orderedPanels(activeProgram data.Program, activeWorkout data.Workout, programPanel string, workoutPanel string, exercisePanel string) []string {
	if activeWorkout.WorkoutId != 0 {
		return []string{exercisePanel, workoutPanel, programPanel}
	}
	if activeProgram.ProgramId != 0 {
		return []string{workoutPanel, programPanel, exercisePanel}
	}
	return []string{programPanel, workoutPanel, exercisePanel}
}

func renderExerciseSection(styles theme.Styles, title string, exercises []data.Exercise, emptyHint string, cursor int) string {
	var lines []string
	lines = append(lines, styles.ProgramPanelTitle.Render(title))
	lines = append(lines, "")
	if len(exercises) == 0 {
		lines = append(lines, styles.ProgramEmpty.Render("run: "+emptyHint))
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
			Render(rowStyle.Render(marker+" ") + exercise.Name)
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
	lines = append(lines, "")
	if len(values) == 0 {
		lines = append(lines, styles.ProgramEmpty.Render("run: "+emptyHint))
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
		names = append(names, program.ProgramName)
	}
	return names
}

func workoutNames(workouts []data.Workout) []string {
	names := make([]string, 0, len(workouts))
	for _, workout := range workouts {
		names = append(names, workout.Name)
	}
	return names
}

func exerciseLabel(exercise data.Exercise) string {
	if exercise.Sets > 0 || exercise.Reps > 0 {
		return fmt.Sprintf("%s  %dx%d", exercise.Name, exercise.Sets, exercise.Reps)
	}
	return exercise.Name
}

func exerciseTarget(exercise data.Exercise) string {
	if exercise.Sets > 0 || exercise.Reps > 0 {
		return fmt.Sprintf("%dx%d", exercise.Sets, exercise.Reps)
	}
	return ""
}
