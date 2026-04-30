package theme

import (
	"charm.land/lipgloss/v2"
)

type Styles struct {
	Opener lipgloss.Style
}

func NewStyles (t Theme) Styles {
	newStyles := Styles{
		Opener: lipgloss.NewStyle().Foreground(t.Foreground).Background(t.Background),
	}
	return newStyles
}