package screens

import (
	"strings"
	"testing"

	"ruffnut/data"
	"ruffnut/ui/theme"
)

func TestHistoryViewShowsSessionEntries(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 100, 30)
	view := HistoryView(
		styles,
		data.Workout{WorkoutId: 1, Name: "push"},
		"push sessions",
		nil,
		0,
		data.GymSession{SessionId: 7, WorkoutId: 1, StartedAt: "2026-07-06T10:00:00Z"},
		[]data.GymSessionEntry{{Exercise: "bench", Sets: 2, Reps: 4, RepsDetail: "6/4", Weight: 135}},
	)

	for _, want := range []string{"logs", "ID #7", "bench", "2x6/4"} {
		if !strings.Contains(view, want) {
			t.Fatalf("HistoryView() missing %q; view:\n%s", want, view)
		}
	}
}

func TestHistoryViewShowsExerciseHistory(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 100, 30)
	view := HistoryView(
		styles,
		data.Workout{WorkoutId: 1, Name: "push"},
		"bench across ppl",
		nil,
		0,
		data.GymSession{},
		[]data.GymSessionEntry{{SessionId: 7, Exercise: "bench", Workout: "upper body", StartedAt: "2026-07-06T10:00:00Z", Sets: 2, Reps: 4, RepsDetail: "6/4"}},
	)

	for _, want := range []string{"movement history", "ID #7", "upper body", "2x6/4"} {
		if !strings.Contains(view, want) {
			t.Fatalf("HistoryView() missing %q; view:\n%s", want, view)
		}
	}
}
