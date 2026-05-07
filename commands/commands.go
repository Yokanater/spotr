package commands

type Command struct {
	Cmd  string
	Args []string
}

type Spec struct {
	// Command name (shown in help)
	Name string
	// Alt names
	Aliases []string
	// Usage (example)
	Usage string
	// One line description
	Summary string
}

var Registry = map[string]Spec{
	"help": {
		Name:    "help",
		Aliases: []string{"h", "?"},
		Usage:   "help",
		Summary: "show available commands",
	},
	"quit": {
		Name:    "quit",
		Aliases: []string{"q", "exit"},
		Usage:   "quit",
		Summary: "exit ruffnut",
	},
	"program": {
		Name:    "program",
		Aliases: []string{"prog"},
		Usage:   "program <add|list|use> ...",
		Summary: "manage workout programs",
	},
	"workout": {
		Name:    "workout",
		Aliases: []string{"w"},
		Usage:   "workout <add|list> ...",
		Summary: "manage workout templates",
	},
	"log": {
		Name:    "log",
		Aliases: []string{"l"},
		Usage:   "log <start|save|list> ...",
		Summary: "record and manage workout sessions",
	},
	"history": {
		Name:    "history",
		Aliases: []string{"hist"},
		Usage:   "history <list|show> ...",
		Summary: "browse saved workout sesssions",
	},
}

var AliasToCanonical = map[string]string{
	"h":    "help",
	"?":    "help",
	"q":    "quit",
	"exit": "quit",
	"prog": "program",
	"w":    "workout",
	"l":    "log",
	"hist": "history",
}
