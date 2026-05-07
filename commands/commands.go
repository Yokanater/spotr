package commands

type Command struct{
	Cmd string
	Args []string
}

type Spec struct{
	Name string
	Aliases []string
	Usage string
	Short string
}

var Registry = map[string]Spec{
	"help": {
		Name: "help",
		Aliases: []string{"please", "sorry"},
		Usage: "Prints a list of commands and what to do with them",
		Short: "h",
	},
};