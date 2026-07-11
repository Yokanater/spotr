package screens

import (
	"fmt"
	"github.com/Yokanater/spotr/data"
	"github.com/Yokanater/spotr/ui/theme"
	"strings"

	"charm.land/lipgloss/v2"
)

var logo = `
                                    ░██             
                                    ░██             
 ░███████  ░████████   ░███████  ░████████ ░██░████ 
░██        ░██    ░██ ░██    ░██    ░██    ░███     
 ░███████  ░██    ░██ ░██    ░██    ░██    ░██      
       ░██ ░███   ░██ ░██    ░██    ░██    ░██      
 ░███████  ░██░█████   ░███████      ░████ ░██      
           ░██                                      
           ░██                                      
`

var compactLogo = `
  spotr
`

func HomeView(styles theme.Styles, programs []data.Program, cursor int, activeProgram data.Program) string {
	header := RenderHeader(styles, "home")
	if len(programs) == 0 {
		return firstRunHomeView(styles, header)
	}
	title := styles.ProgramTitle.Render("programs")
	subtitleText := "Choose a program to open its workouts."
	emptyHint := "press a to create a program or t to browse templates"
	subtitle := styles.ProgramSubtitle.Render(subtitleText)
	panel := renderProgramSection(styles, "your programs", homeProgramNames(programs, activeProgram), emptyHint, cursor)
	panel = lipgloss.NewStyle().Width(styles.Box.GetWidth()).Align(lipgloss.Center).Render(panel)

	return lipgloss.JoinVertical(lipgloss.Left, header, "", title, subtitle, "", panel)
}

func firstRunHomeView(styles theme.Styles, header string) string {
	wordmark := logo
	if styles.Logo.GetWidth() < 58 {
		wordmark = compactLogo
	}
	title := styles.Logo.Render(wordmark)
	tagline := styles.Tagline.Render("your first workout starts here")
	actions := lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.HelperKey.Render("a"), " start from scratch",
		styles.HelperSeparator.Render("   ·   "),
		styles.HelperKey.Render("t"), " use a template",
	)
	actions = lipgloss.NewStyle().Width(styles.Box.GetWidth()).Align(lipgloss.Center).Render(actions)
	return lipgloss.JoinVertical(lipgloss.Center, header, title, tagline, "", actions)
}

func homeProgramNames(programs []data.Program, activeProgram data.Program) []string {
	names := make([]string, 0, len(programs))
	for _, program := range programs {
		name := fmt.Sprintf("#%d  %s", program.ProgramId, program.ProgramName)
		if program.ProgramId == activeProgram.ProgramId {
			name += "  · current"
		}
		names = append(names, name)
	}
	return names
}

func RenderHeader(styles theme.Styles, active string) string {
	brand := styles.Brand.Render("spotr")
	gap := "   "
	if styles.Header.GetWidth() < 72 {
		gap = " "
	}
	nav := renderNav(styles, active, gap, []string{"home", "workouts", "templates", "logs", "help"})
	return styles.Header.Align(lipgloss.Center).Render(lipgloss.JoinHorizontal(lipgloss.Top, brand, "    ", nav))
}

func renderNav(styles theme.Styles, active string, gap string, items []string) string {
	rendered := make([]string, 0, len(items))
	for _, item := range items {
		if item == active {
			rendered = append(rendered, styles.Brand.Render(item))
			continue
		}
		rendered = append(rendered, styles.Nav.Render(item))
	}
	return strings.Join(rendered, gap)
}
