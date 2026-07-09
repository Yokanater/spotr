package commands

type KeyBinding struct {
	Key    string
	Action string
}

var KeyBindings = []KeyBinding{
	{
		Key:    "j",
		Action: "move next",
	},
	{
		Key:    "k",
		Action: "move previous",
	},
	{
		Key:    "enter",
		Action: "open selected item",
	},
	{
		Key:    "a",
		Action: "add item",
	},
	{
		Key:    "s",
		Action: "start workout log",
	},
	{
		Key:    "l",
		Action: "log actual sets and reps",
	},
	{
		Key:    "v",
		Action: "view workout logs",
	},
	{
		Key:    "t",
		Action: "browse templates",
	},
	{
		Key:    "f",
		Action: "finish workout log",
	},
	{
		Key:    "e",
		Action: "edit selected item",
	},
	{
		Key:    "d",
		Action: "delete selected item",
	},
	{
		Key:    "b/esc",
		Action: "go back",
	},
	{
		Key:    ":",
		Action: "command mode",
	},
	{
		Key:    "?",
		Action: "help",
	},
	{
		Key:    "q",
		Action: "confirm quit",
	},
}
