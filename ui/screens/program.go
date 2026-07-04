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
	programPanel := renderProgramSection(styles, "01 programs", programNames(programs), "press a to add program", programCursor)
	workoutPanel := renderProgramSection(styles, "02 workouts", workoutNames(workouts), "press a to add workout", workoutCursor)
	exercisePanel := renderExerciseSection(styles, "03 exercises", exercises, "press a to add exercise", exerciseCursor)

	panels := visiblePanels(styles, activeProgram, activeWorkout, programPanel, workoutPanel, exercisePanel)
	panelBlock := lipgloss.JoinHorizontal(lipgloss.Top, panels...)
	if styles.ProgramTitle.GetWidth() < 82 {
		panelBlock = lipgloss.JoinVertical(lipgloss.Left, panels...)
	}
	footer := styles.ProgramSubtitle.Render(actionHint(activeProgram, activeWorkout))
	return lipgloss.JoinVertical(lipgloss.Left, RenderHeader(styles, "program"), "", title, subtitle, "", panelBlock, "", footer)
}

func actionHint(activeProgram data.Program, activeWorkout data.Workout) string {
	if activeWorkout.WorkoutId != 0 {
		return "a add exercise   enter select exercise   b back to workouts   : command"
	}
	if activeProgram.ProgramId != 0 {
		return "a add workout   enter open workout   b back to programs   : command"
	}
	return "a add program   enter open program   : command"
}

func visiblePanels(styles theme.Styles, activeProgram data.Program, activeWorkout data.Workout, programPanel string, workoutPanel string, exercisePanel string) []string {
	if styles.ProgramTitle.GetWidth() < 82 {
		if activeWorkout.WorkoutId != 0 {
			return []string{exercisePanel}
		}
		if activeProgram.ProgramId != 0 {
			return []string{workoutPanel}
		}
		return []string{programPanel}
	}

	if activeWorkout.WorkoutId != 0 {
		return []string{programPanel, workoutPanel, exercisePanel}
	}
	if activeProgram.ProgramId != 0 {
		return []string{programPanel, workoutPanel}
	}
	return []string{programPanel}
}

func renderExerciseSection(styles theme.Styles, title string, exercises []data.Exercise, emptyHint string, cursor int) string {
	var lines []string
	lines = append(lines, styles.ProgramPanelTitle.Render(title))
	lines = append(lines, "")
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
