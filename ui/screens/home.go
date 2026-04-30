package screens

import (
	"ruffnut/ui/theme"
)

var logo string = `
'        ____  __  _______________   ____  ________
'       / __ \/ / / / ____/ ____/ | / / / / /_  __/
'      / /_/ / / / / /_  / /_  /  |/ / / / / / /   
'     / _, _/ /_/ / __/ / __/ / /|  / /_/ / / /    
'    /_/ |_|\____/_/   /_/   /_/ |_/\____/ /_/     
'                                                  
	
	`   

func HomeView (styles theme.Styles) string {
	s := styles.Opener.Render(logo)
	return s
}