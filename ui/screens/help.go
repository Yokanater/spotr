package screens

import (
	"fmt"
	"ruffnut/commands"
	"ruffnut/ui/theme"
)


func HelpView (styles theme.Styles) string {
	help := ""
	registry := commands.Registry
	for _ , v := range registry {
		str := fmt.Sprintf("%v: %v \n", v.Name, v.Summary)
		help += str
	}
	s := styles.Help.Render(help)
	return s
}