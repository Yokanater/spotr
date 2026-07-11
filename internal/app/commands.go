package app

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"github.com/Yokanater/spotr/commands"
)

func (m model) runCommandLine(line string) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	command, ok := commands.Parse(line)
	if !ok {
		m.status = "error parsing command"
		return m, cmd
	}
	resolved, status := commands.Resolve(command)

	if !status {
		m.status = fmt.Sprintf("Command not defined: %v", resolved)
		return m, cmd
	}
	switch resolved {
	case "program":
		m.handleProgram(command.Args)

	case "workout":
		m.handleWorkout(command.Args)
		return m, cmd
	case "exercise":
		m.handleExercise(command.Args)
		return m, cmd
	case "log":
		m.handleLog(command.Args)
		return m, cmd
	case "history":
		m.handleHistory(command.Args)
		return m, cmd
	case "template":
		m.handleTemplate(command.Args)
		return m, cmd
	case "help":
		m.helpReturnScreen = m.screen
		m.screen = screenHelp
		return m, cmd

	case "home":
		m.screen = screenHome
		return m, cmd

	case "quit":
		return m, tea.Quit
	}
	return m, cmd
}
