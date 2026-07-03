package commands

import (
	"strings"
)

func Parse(msg string) (Command, bool) {
	msg = strings.TrimSpace(msg)
	if msg == "" {
		return Command{}, false
	}
	parts := strings.Fields(msg)
	cmd := strings.ToLower(parts[0])
	args := parts[1:]
	final := Command{Cmd: cmd, Args: args}
	return final, true
}

func Resolve(cmd Command) (canonical string, ok bool) {
	resolved, found := Registry[cmd.Cmd]
	if !found {
		name, found := AliasToCanonical[cmd.Cmd]
		if found {
			resolved, found := Registry[name]
			return resolved.Name, found
		}
		return cmd.Cmd, false
	}
	return resolved.Name, found
}

func HandleKeys(key string) (cmd string) {
	return ""
}
