package theme

import (
	"charm.land/lipgloss/v2"
)

type Styles struct {
	Opener            lipgloss.Style
	Box               lipgloss.Style
	Input             lipgloss.Style
	Status            lipgloss.Style
	HelperKey         lipgloss.Style
	HelperSeparator   lipgloss.Style
	Help              lipgloss.Style
	Header            lipgloss.Style
	Brand             lipgloss.Style
	Nav               lipgloss.Style
	Logo              lipgloss.Style
	Tagline           lipgloss.Style
	SectionTitle      lipgloss.Style
	HelpRow           lipgloss.Style
	HelpKey           lipgloss.Style
	HelpText          lipgloss.Style
	ProgramTitle      lipgloss.Style
	ProgramSubtitle   lipgloss.Style
	ProgramPanel      lipgloss.Style
	ProgramPanelTitle lipgloss.Style
	ProgramItem       lipgloss.Style
	ProgramSelected   lipgloss.Style
	ProgramEmpty      lipgloss.Style
	ProgramListRows   int
}

func NewStyles(t Theme, w int, h int) Styles {
	contentW := min(t.InputMax, max(1, w-t.PadX))
	helpPaddingX := 3
	if contentW < 72 {
		helpPaddingX = 1
	}
	helpRowW := min(contentW, max(12, (contentW-8)/2))
	panelW := max(12, (contentW-4)/3)
	programListRows := max(3, h-13)
	if contentW < 82 {
		panelW = contentW
		programListRows = max(3, (h-13)/3)
	}
	newStyles := Styles{

		Opener: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(w).
			Foreground(t.Foreground).
			Background(t.Background),

		Box: lipgloss.NewStyle().
			Width(contentW).
			MaxHeight(h).
			Background(t.Background),

		Input: lipgloss.NewStyle().
			Width(contentW).
			Height(3).
			Align(lipgloss.Left).
			AlignVertical(lipgloss.Center).
			Background(t.Background).
			Border(lipgloss.NormalBorder()).
			BorderForeground(t.Divider).
			Foreground(t.Text).
			Padding(0, 2),

		Status: lipgloss.NewStyle().
			Height(1).
			Foreground(t.TextMuted).
			Width(contentW).
			Align(lipgloss.Left),

		HelperKey: lipgloss.NewStyle().
			Foreground(t.Accent).
			Bold(true),

		HelperSeparator: lipgloss.NewStyle().
			Foreground(t.TextFaint),

		Help: lipgloss.NewStyle().
			Width(contentW).
			Foreground(t.Text).
			Border(lipgloss.NormalBorder()).
			BorderForeground(t.Border).
			Padding(1, helpPaddingX),

		Header: lipgloss.NewStyle().
			Width(contentW).
			Foreground(t.TextMuted),

		Brand: lipgloss.NewStyle().
			Foreground(t.Foreground).
			Bold(true),

		Nav: lipgloss.NewStyle().
			Foreground(t.TextFaint),

		Logo: lipgloss.NewStyle().
			Width(contentW).
			Align(lipgloss.Center).
			Foreground(t.Foreground).
			Bold(true),

		Tagline: lipgloss.NewStyle().
			Width(contentW).
			Align(lipgloss.Center).
			Foreground(t.TextMuted),

		SectionTitle: lipgloss.NewStyle().
			Foreground(t.Highlight).
			Bold(true),

		HelpRow: lipgloss.NewStyle().
			Width(helpRowW),

		HelpKey: lipgloss.NewStyle().
			Width(14).
			Foreground(t.Accent).
			Bold(true),

		HelpText: lipgloss.NewStyle().
			Foreground(t.Text),

		ProgramTitle: lipgloss.NewStyle().
			Width(contentW).
			Align(lipgloss.Left).
			Foreground(t.Foreground).
			Bold(true),

		ProgramSubtitle: lipgloss.NewStyle().
			Width(contentW).
			Align(lipgloss.Left).
			Foreground(t.TextMuted),

		ProgramPanel: lipgloss.NewStyle().
			Width(panelW).
			Border(lipgloss.NormalBorder()).
			BorderForeground(t.Border).
			Foreground(t.Text).
			Padding(1, 2).
			MarginRight(1),

		ProgramPanelTitle: lipgloss.NewStyle().
			Foreground(t.Highlight).
			Bold(true),

		ProgramItem: lipgloss.NewStyle().
			Foreground(t.Text),

		ProgramSelected: lipgloss.NewStyle().
			Foreground(t.Accent).
			Bold(true),

		ProgramEmpty: lipgloss.NewStyle().
			Foreground(t.TextFaint),

		ProgramListRows: programListRows,
	}
	return newStyles
}
