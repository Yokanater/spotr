package screens

import (
	"ruffnut/data"
	"ruffnut/ui/theme"
)

func ProgramView(styles theme.Styles, programs []data.Program) string {
	raw := ""
	for i := range programs {
		program := programs[i].ProgramName
		raw += program
	}
	s := styles.Help.Render(raw)
	return s
}
