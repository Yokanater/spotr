package screens

import (
	"strings"
	"testing"

	"spotr/data"
	"spotr/ui/theme"

	"charm.land/lipgloss/v2"
)

func TestHomeViewShowsProgramLauncher(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 88, 24)
	view := HomeView(styles, []data.Program{{ProgramId: 1, ProgramName: "PPL"}}, 0, data.Program{ProgramId: 1})

	for _, want := range []string{"programs", "PPL", "current", "Choose a program"} {
		if !strings.Contains(view, want) {
			t.Fatalf("HomeView() missing %q; view:\n%s", want, view)
		}
	}
}

func TestHomeViewEmptyStateOffersCreationAndTemplates(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 88, 24)
	view := HomeView(styles, nil, 0, data.Program{})

	for _, want := range []string{"spotr", "your first workout starts here", "start from scratch", "use a template"} {
		if !strings.Contains(view, want) {
			t.Fatalf("HomeView() missing %q; view:\n%s", want, view)
		}
	}
}

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
