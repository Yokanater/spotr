package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Theme struct {
	// surfaces
	Background color.Color
	Foreground color.Color

	// text
	Text      color.Color
	TextMuted color.Color
	TextFaint color.Color

	// accents
	Accent color.Color

	// frames
	Border    color.Color
	Divider   color.Color
	Highlight color.Color
	// layout
	Radius   int
	PadX     int
	PadY     int
	InputMax int
}

func Default() Theme {
	defaultTheme := Theme{
		Background: lipgloss.Color("#151515"),
		Foreground: lipgloss.Color("#F1E8C8"),
		Accent:     lipgloss.Color("#E8D889"),
		Text:       lipgloss.Color("#E7E1D2"),
		TextMuted:  lipgloss.Color("#A8A397"),
		TextFaint:  lipgloss.Color("#6F6B63"),
		Border:     lipgloss.Color("#3A3833"),
		Divider:    lipgloss.Color("#4B4840"),
		Highlight:  lipgloss.Color("#D8C77A"),
		InputMax:   88,
		PadX:       8,
	}
	return defaultTheme
}
