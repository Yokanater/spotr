package commands

type KeyBinding struct {
	Key    string
	Action string
}

var KeyBindings = []KeyBinding{
	{
		Key:    "↑/↓ k/j",
		Action: "move",
	},
	{
		Key:    "enter",
		Action: "open",
	},
	{
		Key:    "a",
		Action: "add",
	},
	{
		Key:    "s",
		Action: "start",
	},
	{
		Key:    "l",
		Action: "log sets",
	},
	{
		Key:    "v",
		Action: "graph/logs",
	},
	{
		Key:    "t",
		Action: "templates",
	},
	{
		Key:    "p",
		Action: "programs",
	},
	{
		Key:    "f",
		Action: "finish",
	},
	{
		Key:    "e",
		Action: "edit",
	},
	{
		Key:    "d",
		Action: "delete",
	},
	{
		Key:    "b/esc",
		Action: "back",
	},
	{
		Key:    ":",
		Action: "commands",
	},
	{
		Key:    "?",
		Action: "help",
	},
	{
		Key:    "q",
		Action: "quit",
	},
}
