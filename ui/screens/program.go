package screens

import (
	"ruffnut/data"
	"ruffnut/ui/theme"
	"strings"
)

func ProgramView(styles theme.Styles, programs []data.Program, workouts []data.Workout, exercises []data.Exercise, activeProgram data.Program, activeWorkout data.Workout) string {
	var lines []string
	if activeProgram.ProgramId == 0 {
		lines = append(lines, "No active program selected.")
	} else {
		lines = append(lines, "Program: "+activeProgram.ProgramName)
	}
	if activeWorkout.WorkoutId != 0 {
		lines = append(lines, "Workout: "+activeWorkout.Name)
	}

	lines = append(lines, "")
	lines = append(lines, "Programs:")
	for i := range programs {
		lines = append(lines, "- "+programs[i].ProgramName)
	}

	lines = append(lines, "")
	lines = append(lines, "Workouts:")
	if len(workouts) == 0 {
		lines = append(lines, "- none")
	} else {
		for i := range workouts {
			lines = append(lines, "- "+workouts[i].Name)
		}
	}

	lines = append(lines, "")
	lines = append(lines, "Exercises:")
	if len(exercises) == 0 {
		lines = append(lines, "- none")
	} else {
		for i := range exercises {
			lines = append(lines, "- "+exercises[i].Name)
		}
	}

	return styles.Help.Render(strings.Join(lines, "\n"))
}
