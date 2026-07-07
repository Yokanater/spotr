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

func TestHistoryViewKeepsSelectedSessionVisible(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 100, 16)
	view := HistoryView(
		styles,
		data.Workout{WorkoutId: 1, Name: "push"},
		"push sessions",
		[]data.GymSession{
			{SessionId: 1, WorkoutId: 1, StartedAt: "2026-07-01T10:00:00Z"},
			{SessionId: 2, WorkoutId: 1, StartedAt: "2026-07-02T10:00:00Z"},
			{SessionId: 3, WorkoutId: 1, StartedAt: "2026-07-03T10:00:00Z"},
			{SessionId: 4, WorkoutId: 1, StartedAt: "2026-07-04T10:00:00Z"},
			{SessionId: 5, WorkoutId: 1, StartedAt: "2026-07-05T10:00:00Z"},
		},
		4,
		data.GymSession{},
		nil,
	)

	if !strings.Contains(view, "ID #5") {
		t.Fatalf("HistoryView() did not keep selected session visible; view:\n%s", view)
	}
	if strings.Contains(view, "ID #1") {
		t.Fatalf("HistoryView() rendered the top instead of scrolling; view:\n%s", view)
	}
}
