package screens

import (
	"spotr/ui/theme"
	"strings"

	"charm.land/lipgloss/v2"
)

var logo string = `
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

var compactLogo string = `
  spotr
`

func HomeView(styles theme.Styles) string {
	header := RenderHeader(styles, "home")
	wordmark := logo
	if styles.Logo.GetWidth() < 58 {
		wordmark = compactLogo
	}
	title := styles.Logo.Render(wordmark)
	tagline := styles.Tagline.Render("lift logging for nerds")

	return lipgloss.JoinVertical(lipgloss.Center, header, "", title, tagline)
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
