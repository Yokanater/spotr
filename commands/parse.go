package commands

import (
	"strings"
)

func Parse (msg string) Command {
	parts := strings.Fields(msg)
	cmd := strings.ToLower(parts[0])
	args := parts[1:]
	final := Command{Cmd: cmd, Args: args}
	return final
}