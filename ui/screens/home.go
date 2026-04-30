package screens

import (
	tea "charm.land/bubbletea/v2"
	"ruffnut/ui/theme"
)


type model struct {

}

func main () {

}

var logo string = `
'        ____  __  _______________   ____  ________
'       / __ \/ / / / ____/ ____/ | / / / / /_  __/
'      / /_/ / / / / /_  / /_  /  |/ / / / / / /   
'     / _, _/ /_/ / __/ / __/ / /|  / /_/ / / /    
'    /_/ |_|\____/_/   /_/   /_/ |_/\____/ /_/     
'                                                  
	
	`   

func HomeView () tea.View {
	styles := theme.NewStyles(theme.Default())
	s := styles.Opener.Render(logo)
	v := tea.NewView(s)
	return v
}