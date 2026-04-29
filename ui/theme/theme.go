package theme

import (
	"image/color"
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