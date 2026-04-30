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
	Text color.Color
	TextMuted color.Color
	TextFaint color.Color

	// accents
	Accent color.Color

	// frames
	Border color.Color
	Divider color.Color

	// layout
	Radius int
	PadX int
	PadY int
}

func Default () Theme {
	defaultTheme := Theme{
		Background: lipgloss.Color("#111111"),
		Foreground: lipgloss.Color("#0000ff"),
		Accent: lipgloss.Color("#5fd8d0"),
		Text: lipgloss.Color("#db386e"),
	}
	return defaultTheme
}