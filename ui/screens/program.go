package screens

import (
	"fmt"
	"ruffnut/data"
	"ruffnut/ui/theme"
	"strings"

	"charm.land/lipgloss/v2"
)

func ProgramView(styles theme.Styles, programs []data.Program, workouts []data.Workout, exercises []data.Exercise, activeProgram data.Program, activeWorkout data.Workout, activeExercise data.Exercise) string {
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
	programPanel := renderProgramSection(styles, "01 programs", programNames(programs), "program list")
	workoutPanel := renderProgramSection(styles, "02 workouts", workoutNames(workouts), "workout list")
	exercisePanel := renderProgramSection(styles, "03 exercises", exerciseNames(exercises), "exercise list")

	panels := lipgloss.JoinHorizontal(lipgloss.Top, programPanel, workoutPanel, exercisePanel)
	if styles.ProgramTitle.GetWidth() < 82 {
		panels = lipgloss.JoinVertical(lipgloss.Left, programPanel, workoutPanel, exercisePanel)
	}
	footer := styles.ProgramSubtitle.Render("program add <name>  /  workout add <name>  /  exercise add <name> [sets] [reps]")
	return lipgloss.JoinVertical(lipgloss.Left, RenderHeader(styles, "program"), "", title, subtitle, "", panels, "", footer)
}

func renderProgramSection(styles theme.Styles, title string, values []string, emptyHint string) string {
	var lines []string
	lines = append(lines, styles.ProgramPanelTitle.Render(title))
	lines = append(lines, "")
	if len(values) == 0 {
		lines = append(lines, styles.ProgramEmpty.Render("run: "+emptyHint))
		return styles.ProgramPanel.Render(strings.Join(lines, "\n"))
	}

	for i, value := range values {
		prefix := "  "
		if i == 0 {
			prefix = "> "
		}
		lines = append(lines, styles.ProgramItem.Render(prefix+value))
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
