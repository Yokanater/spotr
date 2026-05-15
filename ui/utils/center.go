package utils

import "charm.land/lipgloss/v2"

func CenterPlace(w int, h int, i string) string {
	box := lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, i)
	return box
}
