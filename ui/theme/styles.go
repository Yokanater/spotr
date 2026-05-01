package theme

import (
	"charm.land/lipgloss/v2"
)

type Styles struct {
	Opener lipgloss.Style
	Box lipgloss.Style
	Input lipgloss.Style
}

func NewStyles (t Theme, w int, h int) Styles {
	newStyles := Styles{
		Opener: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(w).
			Foreground(t.Accent).
			Background(t.Background),
		Box: lipgloss.NewStyle().
			Background(t.Background).
			Width(w).
			Height(h),
		Input: lipgloss.NewStyle().
			Width(min(t.InputMax, w - t.PadX)).
			Height(3).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Background(t.Accent).
			Foreground(t.Background),
	}
	return newStyles
}