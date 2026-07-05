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
		data.Program{ProgramId: 1, ProgramName: "ppl"},
		data.Workout{WorkoutId: 1, ProgramId: 1, Name: "push"},
		data.Exercise{},
		0,
		0,
		0,
	)

	if !strings.Contains(view, "3x10") {
		t.Fatalf("ProgramView() did not render exercise target; view:\n%s", view)
	}
}

func TestProgramViewShowsProgressivePanels(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 100, 30)

	view := ProgramView(
		styles,
		[]data.Program{{ProgramId: 1, ProgramName: "ppl"}},
		[]data.Workout{{WorkoutId: 1, ProgramId: 1, Name: "push"}},
		[]data.Exercise{{ExerciseId: 1, WorkoutId: 1, Name: "bench", Sets: 3, Reps: 10}},
		data.Program{},
		data.Workout{},
		data.Exercise{},
		0,
		0,
		0,
	)
	if !strings.Contains(view, "01 programs") {
		t.Fatalf("ProgramView() did not render programs panel; view:\n%s", view)
	}
	if strings.Contains(view, "02 workouts") || strings.Contains(view, "03 exercises") {
		t.Fatalf("ProgramView() rendered panels before selection; view:\n%s", view)
	}

	view = ProgramView(
		styles,
		[]data.Program{{ProgramId: 1, ProgramName: "ppl"}},
		[]data.Workout{{WorkoutId: 1, ProgramId: 1, Name: "push"}},
		[]data.Exercise{{ExerciseId: 1, WorkoutId: 1, Name: "bench", Sets: 3, Reps: 10}},
		data.Program{ProgramId: 1, ProgramName: "ppl"},
		data.Workout{},
		data.Exercise{},
		0,
		0,
		0,
	)
	if !strings.Contains(view, "01 programs") || !strings.Contains(view, "02 workouts") {
		t.Fatalf("ProgramView() did not render program and workout panels; view:\n%s", view)
	}
	if strings.Contains(view, "03 exercises") {
		t.Fatalf("ProgramView() rendered exercises before workout selection; view:\n%s", view)
	}
}

func TestRenderHeaderShowsLog(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 100, 30)
	view := RenderHeader(styles, "program")

	if !strings.Contains(view, "log") {
		t.Fatalf("RenderHeader() did not render log nav item; view:\n%s", view)
	}
}
