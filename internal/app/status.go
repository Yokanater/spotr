package app

import (
	"charm.land/lipgloss/v2"
	"spotr/ui/theme"
	"strings"
)

func (m model) normalHelp() string {
	if m.screen == screenHome {
		return helperMessage("enter workouts", "p programs", "? help", "q quit")
	}
	if m.screen == screenPrograms {
		if len(m.programs) == 0 {
			return helperMessage("a add", "t templates", "b back", "? help")
		}
		return helperMessage("↑/↓ move", "enter use", "a add", "b back", "? help")
	}
	if m.screen == screenTemplates {
		if len(m.templateFiles) == 0 {
			return helperMessage("b back", "? help")
		}
		return helperMessage("↑/↓ move", "enter import", "b back", "? help")
	}
	if m.screen == screenHistory {
		if m.historyEntries != nil {
			return helperMessage("↑/↓ move", "enter open", "b back", "? help")
		}
		return helperMessage("↑/↓ move", "enter open", "b back", "? help")
	}

	switch m.currentLevel() {
	case screenPrograms:
		if len(m.programs) == 0 {
			return helperMessage("a add", "t templates", "b home", "? help")
		}
		return helperMessage("↑/↓ move", "enter use", "a add", "b home", "? help")
	case screenWorkouts:
		if len(m.workouts) == 0 {
			return helperMessage("a add", "p programs", "b home", "? help")
		}
		return helperMessage("↑/↓ move", "enter open", "a add", "p programs", "? help")
	case screenExercises:
		if len(m.exercises) == 0 {
			return helperMessage("a add", "b workouts", "? help")
		}
		return helperMessage("↑/↓ move", "l log", "v graph", "b workouts", "? help")
	default:
		return helperMessage("j/k move", "enter open", "b back", "a add", ": command")
	}
}

func (m model) keyHelp() string {
	switch m.mode {
	case modeInput:
		switch m.inputPurpose {
		case inputAddProgram, inputAddWorkout, inputAddExercise:
			return helperMessage("enter create", "esc cancel")
		case inputLogExercise:
			return helperMessage("enter log sets", "esc cancel")
		default:
			return helperMessage("enter save", "esc cancel")
		}
	case modeCmd:
		return helperMessage("enter run command", "esc cancel")
	case modeDelete:
		return helperMessage("enter confirm delete", "esc cancel")
	case modeQuit:
		return helperMessage("enter quit", "esc stay")
	default:
		return m.normalHelp()
	}
}

func helperMessage(parts ...string) string {
	return strings.Join(parts, " · ")
}

func renderStatus(styles theme.Styles, status string) string {
	return renderHelper(styles, styles.Status, status)
}

func renderKeyRail(styles theme.Styles, status string) string {
	return renderHelper(styles, styles.KeyRail, status)
}

func renderHelper(styles theme.Styles, container lipgloss.Style, status string) string {
	parts := strings.Split(status, " · ")
	if len(parts) == 1 {
		return container.Render(status)
	}

	rendered := make([]string, 0, len(parts)*2-1)
	for i, part := range parts {
		if i > 0 {
			rendered = append(rendered, styles.HelperSeparator.Render(" · "))
		}
		rendered = append(rendered, renderHelperPart(styles, part))
	}

	return container.Render(lipgloss.JoinHorizontal(lipgloss.Top, rendered...))
}

func renderHelperPart(styles theme.Styles, part string) string {
	key, rest, ok := strings.Cut(part, " ")
	if !ok || !isHelperKey(key) {
		return part
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, styles.HelperKey.Render(key), " "+rest)
}

func isHelperKey(value string) bool {
	switch value {
	case ":", "?", "↑/↓", "a", "b", "d", "e", "enter", "esc", "f", "j/k", "l", "n", "p", "q", "s", "t", "v", "y":
		return true
	default:
		return false
	}
}
