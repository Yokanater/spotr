package screens

import (
	"strings"
	"testing"

	"spotr/ui/theme"

	"charm.land/lipgloss/v2"
)

func TestHelpViewIncludesCommandUsage(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 100, 30)
	view := HelpView(styles)

	for _, want := range []string{
		"exercise list | exercise add <name>",
		"exercise select <id|name>",
		"log start | log add [exercise] <sets> <reps>",
		"log finish",
		"log current",
		"template list | template show <name|path>",
	} {
		if !strings.Contains(view, want) {
			t.Fatalf("HelpView() did not include %q; view:\n%s", want, view)
		}
	}
}

func TestHelpViewIncludesTemplateKey(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 100, 30)
	view := HelpView(styles)

	if !strings.Contains(view, "browse templates") {
		t.Fatalf("HelpView() did not include template key; view:\n%s", view)
	}
}

func TestHelpViewResponsiveWidths(t *testing.T) {
	for _, width := range []int{44, 64, 100} {
		styles := theme.NewStyles(theme.Default(), width, 30)
		view := HelpView(styles)
		limit := styles.Box.GetWidth() + 2

		for _, line := range strings.Split(view, "\n") {
			if got := lipgloss.Width(line); got > limit {
				t.Fatalf("HelpView(%d) line width = %d, want <= %d:\n%s", width, got, limit, line)
			}
		}
	}
}
