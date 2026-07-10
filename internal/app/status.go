package app

import (
	"charm.land/lipgloss/v2"
	"spotr/ui/theme"
	"strings"
)

func (m model) normalHelp() string {
	if m.screen == screenHome {
		return helperMessage("enter workouts", "p programs", "t templates", "? all keys", "q quit")
	}
	if m.screen == screenPrograms {
		if len(m.programs) == 0 {
			return helperMessage("a new program", "t use template", "b back", "? all keys")
		}
		return helperMessage("↑/↓ choose", "enter use program", "a new", "e rename", "d delete", "b back")
	}
	if m.screen == screenTemplates {
		if len(m.templateFiles) == 0 {
			return helperMessage("b back", ": command")
		}
		return helperMessage("↑/↓ choose", "enter import", "b back", "? all keys")
	}
	if m.screen == screenHistory {
		if m.historyEntries != nil {
			return helperMessage("↑/↓ scroll", "enter open", "e edit", "d delete", "b back")
		}
		return helperMessage("↑/↓ scroll", "enter open", "b back", "? all keys")
	}

	switch m.currentLevel() {
	case screenPrograms:
		if len(m.programs) == 0 {
			return helperMessage("a new program", "t use template", "b home", "? all keys")
		}
		return helperMessage("↑/↓ choose", "enter use program", "a new", "b home")
	case screenWorkouts:
		if len(m.workouts) == 0 {
			return helperMessage("a add first workout", "p switch program", "b home", "? all keys")
		}
		return helperMessage("↑/↓ choose", "enter exercises", "a add", "e rename", "d delete", "p programs", "b home")
	case screenExercises:
		if len(m.exercises) == 0 {
			return helperMessage("a add first exercise", "b workouts", "p programs", "? all keys")
		}
		return helperMessage("↑/↓ choose", "l log sets", "v progress", "e edit", "d delete", "b workouts")
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
	parts := strings.Split(status, " · ")
	if len(parts) == 1 {
		return styles.Status.Render(status)
	}

	rendered := make([]string, 0, len(parts)*2-1)
	for i, part := range parts {
		if i > 0 {
			rendered = append(rendered, styles.HelperSeparator.Render(" · "))
		}
		rendered = append(rendered, renderHelperPart(styles, part))
	}

	return styles.Status.Render(lipgloss.JoinHorizontal(lipgloss.Top, rendered...))
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
