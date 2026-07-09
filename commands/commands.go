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
	"home": {
		Name:    "home",
		Aliases: []string{"~"},
		Usage:   "home",
		Summary: "navigate to the home screen",
	},
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
		Summary: "exit spotr",
	},
	"program": {
		Name:    "program",
		Aliases: []string{"prog"},
		Usage:   "program list | program add <name> | program select <id|name> | program edit <name> | program delete <id|name>",
		Summary: "list, add, select, edit, or delete workout programs",
	},
	"workout": {
		Name:    "workout",
		Aliases: []string{"w"},
		Usage:   "workout list | workout add <name> | workout select <id|name> | workout edit <name> | workout delete <id|name>",
		Summary: "list, add, select, edit, or delete workouts in the active program",
	},
	"exercise": {
		Name:    "exercise",
		Aliases: []string{"ex"},
		Usage:   "exercise list | exercise add <name> [sets] [reps] | exercise select <id|name> | exercise edit <name> [sets] [reps] | exercise delete <id|name> | exercise set <sets> <reps>",
		Summary: "list, add, select, edit, or delete exercises in the active workout",
	},
	"log": {
		Name:    "log",
		Aliases: []string{"l"},
		Usage:   "log start | log add [exercise] <sets> <reps> | log edit <entry-id> <sets> <reps> | log delete <entry-id> | log finish [notes] | log current",
		Summary: "record the active workout session",
	},
	"history": {
		Name:    "history",
		Aliases: []string{"hist"},
		Usage:   "history list [limit] | history show <session-id>",
		Summary: "browse saved workout sessions",
	},
	"template": {
		Name:    "template",
		Aliases: []string{"tmpl"},
		Usage:   "template list | template show <name|path> | template import <name|path> | template workout <template> <workout> | template export [program] [path] | template validate [name|path]",
		Summary: "browse, import, export, and validate templates",
	},
}

var AliasToCanonical = map[string]string{
	"h":    "help",
	"?":    "help",
	"q":    "quit",
	"exit": "quit",
	"prog": "program",
	"w":    "workout",
	"ex":   "exercise",
	"l":    "log",
	"hist": "history",
	"tmpl": "template",
	"~":    "home",
}

var CommandsOrder = []string{"help", "home", "program", "workout", "exercise", "log", "history", "template", "quit"}
