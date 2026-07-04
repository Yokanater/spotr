package screens

import (
	"strings"
	"testing"

	"ruffnut/data"
	"ruffnut/ui/theme"
)

func TestProgramViewShowsExerciseTargetWithLongName(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 88, 30)
	view := ProgramView(
		styles,
		nil,
		nil,
		[]data.Exercise{
			{
				Name: "single arm cable rear delt fly with pause",
				Sets: 3,
				Reps: 10,
			},
		},
		data.Program{},
		data.Workout{},
		data.Exercise{},
		0,
		0,
		0,
	)

	if !strings.Contains(view, "3x10") {
		t.Fatalf("ProgramView() did not render exercise target; view:\n%s", view)
	}
}
