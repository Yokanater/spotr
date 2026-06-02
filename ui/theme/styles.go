package theme

import (
	"charm.land/lipgloss/v2"
)

type Styles struct {
	Opener            lipgloss.Style
	Box               lipgloss.Style
	Input             lipgloss.Style
	Status            lipgloss.Style
	Help              lipgloss.Style
	ProgramTitle      lipgloss.Style
	ProgramSubtitle   lipgloss.Style
	ProgramPanel      lipgloss.Style
	ProgramPanelTitle lipgloss.Style
	ProgramItem       lipgloss.Style
	ProgramEmpty      lipgloss.Style
}

func NewStyles(t Theme, w int, h int) Styles {
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
			Width(min(t.InputMax, w-t.PadX)).
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
			Width(min(t.InputMax, w-t.PadX)).
			Align(lipgloss.Left),

		Help: lipgloss.NewStyle(),

		ProgramTitle: lipgloss.NewStyle().
			Width(min(t.InputMax, w-t.PadX)).
			Align(lipgloss.Center).
			Foreground(t.Accent).
			Bold(true),

		ProgramSubtitle: lipgloss.NewStyle().
			Width(min(t.InputMax, w-t.PadX)).
			Align(lipgloss.Center).
			Foreground(t.TextMuted),

		ProgramPanel: lipgloss.NewStyle().
			Width(max(18, (min(t.InputMax, w-t.PadX)-4)/3)).
			Border(lipgloss.NormalBorder()).
			BorderForeground(t.Border).
			Padding(1, 2),

		ProgramPanelTitle: lipgloss.NewStyle().
			Foreground(t.Accent).
			Bold(true),

		ProgramItem: lipgloss.NewStyle().
			Foreground(t.TextMuted),

		ProgramEmpty: lipgloss.NewStyle().
			Foreground(t.TextFaint),
	}
	return newStyles
}
