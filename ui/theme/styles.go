package theme

import (
	"charm.land/lipgloss/v2"
)

type Styles struct {
	Opener lipgloss.Style
	Box lipgloss.Style
	Input lipgloss.Style
	Status lipgloss.Style
}

func NewStyles (t Theme, w int, h int) Styles {
	newStyles := Styles{

		Opener: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(w).
			Foreground(t.Accent).
			Background(t.Background),

		Box: lipgloss.NewStyle().
			Width(w).
			Height(h),

		Input: lipgloss.NewStyle().
			Width(min(t.InputMax, w - t.PadX)).
			Height(5).
			Align(lipgloss.Center).
			AlignVertical(lipgloss.Center).
			Background(t.Background).
			Border(lipgloss.NormalBorder()).
			BorderForeground(t.Highlight).
			Foreground(t.Background),
		
		Status: lipgloss.NewStyle().
			Height(1).
			Foreground(t.Highlight).
			Width(min(t.InputMax, w - t.PadX)).
			Align(lipgloss.Left),
	}
	return newStyles
}