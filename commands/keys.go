package commands

type KeyBinding struct {
	Key string
	Action string
}

var KeyBindings = []KeyBinding{
    {
        Key: "j/down",
        Action: "move down",
    },
    {
        Key: "k/up",
        Action: "move up",
    },
    {
        Key: "enter",
        Action: "open selected item",
    },
    {
        Key: "a",
        Action: "add item",
    },
    {
        Key: "e",
        Action: "edit selected item",
    },
    {
        Key: "d",
        Action: "delete selected item",
    },
    {
        Key: "b/esc",
        Action: "go back",
    },
    {
        Key: ":",
        Action: "command mode",
    },
    {
        Key: "?",
        Action: "help",
    },
    {
        Key: "q",
        Action: "quit",
    },
}