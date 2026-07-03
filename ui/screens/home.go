package screens

import (
	"ruffnut/ui/theme"

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
	tagline := styles.Tagline.Render("lift logging for nerds \n")

	return lipgloss.JoinVertical(lipgloss.Center, header, "", title, tagline)
}

func RenderHeader(styles theme.Styles, active string) string {
	brand := styles.Brand.Render("spotr")
	nav := "home   program   workout   exercise   help"
	if styles.Header.GetWidth() < 72 {
		nav = "home program help"
	}
	if active != "" {
		nav += "   / " + active
	}

	return styles.Header.Align(lipgloss.Center).Render(lipgloss.JoinHorizontal(lipgloss.Top, brand, "    ", styles.Nav.Render(nav)))
}
