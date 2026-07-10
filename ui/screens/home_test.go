package screens

import (
	"strings"
	"testing"

	"spotr/ui/theme"

	"charm.land/lipgloss/v2"
)

func TestRenderHeaderIncludesTemplates(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 88, 24)
	header := RenderHeader(styles, "templates")

	for _, want := range []string{"home", "workouts", "templates", "logs", "help"} {
		if !strings.Contains(header, want) {
			t.Fatalf("RenderHeader() missing %q: %q", want, header)
		}
	}
}

func TestRenderHeaderResponsiveWidth(t *testing.T) {
	for _, width := range []int{44, 88} {
		styles := theme.NewStyles(theme.Default(), width, 24)
		header := RenderHeader(styles, "home")
		limit := styles.Box.GetWidth() + 2

		for _, line := range strings.Split(header, "\n") {
			if got := lipgloss.Width(line); got > limit {
				t.Fatalf("RenderHeader(%d) line width = %d, want <= %d: %q", width, got, limit, line)
			}
		}
	}
}
