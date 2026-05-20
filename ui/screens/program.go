package screens

import (
	"ruffnut/data"
	"ruffnut/ui/theme"
	"strings"
)

func ProgramView(styles theme.Styles, programs []data.Program, workouts []data.Workout, activeProgram data.Program) string {
	var lines []string
	if activeProgram.ProgramId == 0 {
		lines = append(lines, "No active program selected.")
	} else {
		lines = append(lines, "Program: "+activeProgram.ProgramName)
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

	return styles.Help.Render(strings.Join(lines, "\n"))
}
