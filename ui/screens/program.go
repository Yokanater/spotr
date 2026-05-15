package screens

import (
	"ruffnut/ui/theme"
)

func ProgramView(styles theme.Styles, programs []string) string {
	raw := ""
	for i := range programs {
		program := programs[i]
		raw += program
	}
	s := styles.Help.Render(raw)
	return s
}
