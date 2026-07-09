package screens

import (
	"strings"
	"testing"

	"spotr/ui/theme"

	"charm.land/lipgloss/v2"
)

func TestTemplatesViewShowsTemplates(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 88, 28)
	view := TemplatesView(styles, []TemplateListItem{{
		Name:        "Push Pull Legs",
		Description: "A simple split.",
		Workouts:    3,
		Exercises:   12,
		Details: []TemplateWorkoutItem{{
			Name: "Push",
			Exercises: []TemplateExerciseItem{
				{Name: "Bench Press", Sets: 3, Reps: 8},
			},
		}},
	}}, 0)

	for _, want := range []string{"templates", "Push Pull Legs", "3 workouts / 12 exercises", "Push", "Bench Press  3x8"} {
		if !strings.Contains(view, want) {
			t.Fatalf("TemplatesView() missing %q:\n%s", want, view)
		}
	}
}

func TestTemplatesViewShowsSelectedTemplatePreview(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 88, 28)
	view := TemplatesView(styles, []TemplateListItem{
		{
			Name:      "First",
			Workouts:  1,
			Exercises: 1,
			Details: []TemplateWorkoutItem{{
				Name: "Workout A",
				Exercises: []TemplateExerciseItem{
					{Name: "Exercise A", Sets: 1, Reps: 1},
				},
			}},
		},
		{
			Name:      "Second",
			Workouts:  1,
			Exercises: 1,
			Details: []TemplateWorkoutItem{{
				Name: "Workout B",
				Exercises: []TemplateExerciseItem{
					{Name: "Exercise B", Sets: 2, Reps: 2},
				},
			}},
		},
	}, 1)

	for _, want := range []string{"Second", "Workout B", "Exercise B  2x2"} {
		if !strings.Contains(view, want) {
			t.Fatalf("TemplatesView() missing selected detail %q:\n%s", want, view)
		}
	}
	if strings.Contains(view, "Workout A") || strings.Contains(view, "Exercise A") {
		t.Fatalf("TemplatesView() rendered unselected detail:\n%s", view)
	}
}

func TestTemplatesViewResponsiveWidth(t *testing.T) {
	for _, width := range []int{44, 88} {
		styles := theme.NewStyles(theme.Default(), width, 20)
		view := TemplatesView(styles, []TemplateListItem{{
			Name:        "Push Pull Legs",
			Description: "A simple split.",
			Workouts:    3,
			Exercises:   12,
			Details: []TemplateWorkoutItem{{
				Name: "Push",
				Exercises: []TemplateExerciseItem{
					{Name: "Incline Dumbbell Press", Sets: 3, Reps: 10},
				},
			}},
		}}, 0)
		limit := styles.Box.GetWidth() + 2

		for _, line := range strings.Split(view, "\n") {
			if got := lipgloss.Width(line); got > limit {
				t.Fatalf("TemplatesView(%d) line width = %d, want <= %d:\n%s", width, got, limit, line)
			}
		}
	}
}
