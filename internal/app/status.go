package app

import (
	"charm.land/lipgloss/v2"
	"spotr/ui/theme"
	"strings"
)

func (m model) normalHelp() string {
	if m.screen == screenHome {
		return helperMessage("enter training", "t templates", ": command", "? help")
	}
	if m.screen == screenTemplates {
		if len(m.templateFiles) == 0 {
			return helperMessage("b back", ": command")
		}
		return helperMessage("j/k move", "enter import", "b back", ": command")
	}
	if m.screen == screenHistory {
		if m.historyEntries != nil {
			return helperMessage("j/k scroll", "enter open", "e edit", "d delete", "b back")
		}
		return helperMessage("j/k scroll", "enter open", "b back")
	}

	switch m.currentLevel() {
	case screenPrograms:
		if len(m.programs) == 0 {
			return helperMessage("a add program", "t templates", ": command", "? help")
		}
		return helperMessage("j/k move", "enter open", "e edit", "d delete", "a add")
	case screenWorkouts:
		if len(m.workouts) == 0 {
			return helperMessage("a add your first workout", "b programs", ": command")
		}
		return helperMessage("j/k move", "enter open", "e edit", "d delete", "a add", "b programs")
	case screenExercises:
		if len(m.exercises) == 0 {
			return helperMessage("a add your first exercise", "b workouts", ": command")
		}
		return helperMessage("j/k move", "l log", "v logs", "e edit", "d delete", "b workouts")
	default:
		return helperMessage("j/k move", "enter open", "b back", "a add", ": command")
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
	case ":", "?", "a", "b", "d", "e", "enter", "esc", "f", "j/k", "l", "n", "s", "t", "v", "y":
		return true
	default:
		return false
	}
}
