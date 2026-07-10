package screens

import (
	"strings"
	"testing"

	"spotr/data"
	"spotr/ui/theme"
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
		[]data.GymSessionEntry{{SessionId: 7, Exercise: "bench", Workout: "upper body", StartedAt: "2026-07-06T10:00:00Z", Sets: 2, Reps: 4, RepsDetail: "6/4", Weight: 135}},
	)

	for _, want := range []string{"movement history", "estimated 1RM", "best 162.0", "volume latest 1350", "07-06 load 135.0 2x6/4", "#7", "upper body", "2x6/4"} {
		if !strings.Contains(view, want) {
			t.Fatalf("HistoryView() missing %q; view:\n%s", want, view)
		}
	}
}

func TestProgressCalculationsUseEffortAndVolume(t *testing.T) {
	entry := data.GymSessionEntry{Weight: 135, Sets: 2, Reps: 4, RepsDetail: "6/4"}
	if got := estimatedOneRepMax(entry); got != 162 {
		t.Fatalf("estimatedOneRepMax() = %.1f; want 162.0", got)
	}
	if got := exerciseVolume(entry); got != 1350 {
		t.Fatalf("exerciseVolume() = %.1f; want 1350.0", got)
	}
}

func TestChartPositionsReflectElapsedTime(t *testing.T) {
	points := []data.GymSessionEntry{
		{StartedAt: "2026-07-01T10:00:00Z"},
		{StartedAt: "2026-07-02T10:00:00Z"},
		{StartedAt: "2026-07-10T10:00:00Z"},
	}
	positions := chartPositions(points, 28)
	if positions[0] != 0 || positions[2] != 27 || positions[1] >= 10 {
		t.Fatalf("chartPositions() = %v; want date-proportional spacing", positions)
	}
}

func TestWeightProgressionSortsByDate(t *testing.T) {
	entries := []data.GymSessionEntry{
		{EntryId: 2, StartedAt: "2026-07-07T10:00:00Z", Weight: 145, Sets: 3, Reps: 8},
		{EntryId: 1, StartedAt: "2026-07-01T10:00:00Z", Weight: 135, Sets: 3, Reps: 10},
	}
	points := weightedChartEntries(entries)

	if len(points) != 2 {
		t.Fatalf("weightedChartEntries() len = %d; want 2", len(points))
	}
	if points[0].Weight != 135 || points[1].Weight != 145 {
		t.Fatalf("weightedChartEntries() = %+v; want chronological weights", points)
	}
}

func TestHistoryViewKeepsSelectedMovementLogVisible(t *testing.T) {
	styles := theme.NewStyles(theme.Default(), 100, 14)
	view := HistoryView(
		styles,
		data.Workout{WorkoutId: 1, Name: "push"},
		"bench across ppl",
		nil,
		4,
		data.GymSession{},
		[]data.GymSessionEntry{
			{SessionId: 1, Exercise: "bench", Workout: "push", StartedAt: "2026-07-01T10:00:00Z", Sets: 3, Reps: 8, Weight: 135},
			{SessionId: 2, Exercise: "bench", Workout: "push", StartedAt: "2026-07-02T10:00:00Z", Sets: 3, Reps: 8, Weight: 140},
			{SessionId: 3, Exercise: "bench", Workout: "push", StartedAt: "2026-07-03T10:00:00Z", Sets: 3, Reps: 8, Weight: 145},
			{SessionId: 4, Exercise: "bench", Workout: "push", StartedAt: "2026-07-04T10:00:00Z", Sets: 3, Reps: 8, Weight: 147.5},
			{SessionId: 5, Exercise: "bench", Workout: "push", StartedAt: "2026-07-05T10:00:00Z", Sets: 3, Reps: 8, Weight: 150},
		},
	)

	if !strings.Contains(view, "#5") {
		t.Fatalf("HistoryView() did not keep selected movement log visible; view:\n%s", view)
	}
	if strings.Contains(view, "#1") {
		t.Fatalf("HistoryView() rendered the top movement log instead of scrolling; view:\n%s", view)
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
