package screens

import (
	"strings"
	"testing"

	"spotr/data"
	"spotr/ui/theme"
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
		false,
	)

	if !strings.Contains(view, "3x10") {
		t.Fatalf("ProgramView() did not render exercise target; view:\n%s", view)
	}
}

func TestProgramViewOpensAtWorkoutsAndThenExercises(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 100, 30)

	view := ProgramView(
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
		false,
	)
	if !strings.Contains(view, "workouts") {
		t.Fatalf("ProgramView() did not render workouts panel; view:\n%s", view)
	}
	if strings.Contains(view, "programs") || strings.Contains(view, "exercises") {
		t.Fatalf("ProgramView() added a program step or rendered exercises before workout selection; view:\n%s", view)
	}

	view = ProgramView(
		styles,
		[]data.Program{{ProgramId: 1, ProgramName: "ppl"}},
		[]data.Workout{{WorkoutId: 1, ProgramId: 1, Name: "push"}},
		[]data.Exercise{{ExerciseId: 1, WorkoutId: 1, Name: "bench", Sets: 3, Reps: 10}},
		data.Program{ProgramId: 1, ProgramName: "ppl"},
		data.Workout{WorkoutId: 1, ProgramId: 1, Name: "push"},
		data.Exercise{},
		0,
		0,
		0,
		false,
	)
	if !strings.Contains(view, "workouts") || !strings.Contains(view, "exercises") {
		t.Fatalf("ProgramView() did not render workout and exercise panels; view:\n%s", view)
	}
	if strings.Contains(view, "01 programs") {
		t.Fatalf("ProgramView() rendered programs in the main training flow; view:\n%s", view)
	}
}

func TestProgramViewEmptyProgramsOffersTemplates(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 88, 30)
	view := ProgramView(styles, nil, nil, nil, data.Program{}, data.Workout{}, data.Exercise{}, 0, 0, 0, true)

	if !strings.Contains(view, "press a to create one or t to use a template") {
		t.Fatalf("ProgramView() did not offer templates from empty state; view:\n%s", view)
	}
}

func TestProgramViewDoesNotRenderActionFooter(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 100, 30)
	view := ProgramView(
		styles,
		[]data.Program{{ProgramId: 1, ProgramName: "ppl"}},
		[]data.Workout{{WorkoutId: 1, ProgramId: 1, Name: "push"}},
		nil,
		data.Program{ProgramId: 1, ProgramName: "ppl"},
		data.Workout{},
		data.Exercise{},
		0,
		0,
		0,
		false,
	)

	if strings.Contains(view, "enter open workout") {
		t.Fatalf("ProgramView() rendered duplicate action footer; view:\n%s", view)
	}
}

func TestProgramViewShowsSelectedExerciseInShortViewport(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 60, 8)
	exercises := []data.Exercise{
		{ExerciseId: 1, Name: "bench", Sets: 3, Reps: 8},
		{ExerciseId: 2, Name: "overhead press", Sets: 3, Reps: 8},
		{ExerciseId: 3, Name: "incline dumbbell press", Sets: 3, Reps: 10},
		{ExerciseId: 4, Name: "dip", Sets: 3, Reps: 10},
		{ExerciseId: 5, Name: "triceps pushdown", Sets: 3, Reps: 12},
	}
	view := ProgramView(
		styles,
		nil,
		nil,
		exercises,
		data.Program{ProgramId: 1, ProgramName: "ppl"},
		data.Workout{WorkoutId: 1, ProgramId: 1, Name: "push"},
		data.Exercise{},
		0,
		0,
		4,
		false,
	)

	if !strings.Contains(view, "#5") || !strings.Contains(view, "triceps pushdown") {
		t.Fatalf("ProgramView() did not keep selected exercise visible in short viewport; view:\n%s", view)
	}
	if strings.Contains(view, "#1") {
		t.Fatalf("ProgramView() rendered the top of the list instead of scrolling; view:\n%s", view)
	}
}

func TestRenderHeaderShowsLog(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 100, 30)
	view := RenderHeader(styles, "program")

	if !strings.Contains(view, "log") {
		t.Fatalf("RenderHeader() did not render log nav item; view:\n%s", view)
	}
	if strings.Contains(view, "/ program") {
		t.Fatalf("RenderHeader() rendered redundant active suffix; view:\n%s", view)
	}
}
