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
			Width(w).
			Foreground(t.Accent).
			Background(t.Background).
			Align(lipgloss.Center),
		Box: lipgloss.NewStyle().
			Background(t.Background).
			Width(w).
			Height(h),
	}
	return newStyles
}