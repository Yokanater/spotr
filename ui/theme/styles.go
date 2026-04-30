package theme

import (
	"charm.land/lipgloss/v2"
)

type Styles struct {
	Opener lipgloss.Style
	Box lipgloss.Style
}

func NewStyles (t Theme, h int, w int) Styles {
	newStyles := Styles{
		Opener: lipgloss.NewStyle().
			Foreground(t.Foreground).
			Background(t.Background),
		Box: lipgloss.NewStyle().
			Width(w).
			Height(h),
	}
	return newStyles
}