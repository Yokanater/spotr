package screens

import (
	"fmt"
	"ruffnut/data"
	"ruffnut/ui/theme"
	"strings"

	"charm.land/lipgloss/v2"
)

func ProgramView(styles theme.Styles, programs []data.Program, workouts []data.Workout, exercises []data.Exercise, activeProgram data.Program, activeWorkout data.Workout, activeExercise data.Exercise) string {
	title := styles.ProgramTitle.Render("PROGRAMS")

	var context []string
	if activeProgram.ProgramId == 0 {
		context = append(context, "program: none selected")
	} else {
		context = append(context, "program: "+activeProgram.ProgramName)
	}
	if activeWorkout.WorkoutId == 0 {
		context = append(context, "workout: none selected")
	} else {
		context = append(context, "workout: "+activeWorkout.Name)
	}
	if activeExercise.ExerciseId == 0 {
		context = append(context, "exercise: none selected")
	} else {
		context = append(context, "exercise: "+exerciseLabel(activeExercise))
	}

	subtitle := styles.ProgramSubtitle.Render(strings.Join(context, "  |  "))
	programPanel := renderProgramSection(styles, "programs", programNames(programs), "program list")
	workoutPanel := renderProgramSection(styles, "workouts", workoutNames(workouts), "workout list")
	exercisePanel := renderProgramSection(styles, "exercises", exerciseNames(exercises), "exercise list")

	panels := lipgloss.JoinHorizontal(lipgloss.Top, programPanel, "  ", workoutPanel, "  ", exercisePanel)
	return lipgloss.JoinVertical(lipgloss.Center, title, subtitle, "", panels)
}

func renderProgramSection(styles theme.Styles, title string, values []string, emptyHint string) string {
	var lines []string
	lines = append(lines, styles.ProgramPanelTitle.Render(title))
	lines = append(lines, "")
	if len(values) == 0 {
		lines = append(lines, styles.ProgramEmpty.Render(emptyHint))
		return styles.ProgramPanel.Render(strings.Join(lines, "\n"))
	}

	for _, value := range values {
		lines = append(lines, styles.ProgramItem.Render(value))
	}

	return styles.ProgramPanel.Render(strings.Join(lines, "\n"))
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

func exerciseNames(exercises []data.Exercise) []string {
	names := make([]string, 0, len(exercises))
	for _, exercise := range exercises {
		names = append(names, exerciseLabel(exercise))
	}
	return names
}

func exerciseLabel(exercise data.Exercise) string {
	if exercise.Sets > 0 || exercise.Reps > 0 {
		return fmt.Sprintf("%s  %dx%d", exercise.Name, exercise.Sets, exercise.Reps)
	}
	return exercise.Name
}
