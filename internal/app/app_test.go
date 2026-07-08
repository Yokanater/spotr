package app

import (
	"ruffnut/data"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestParseLoggedExerciseValue(t *testing.T) {
	sets, reps, repsDetail, weight, notes, err := parseLoggedExerciseValue("4 12 42.5 last set hard")
	if err != nil {
		t.Fatalf("parseLoggedExerciseValue() error = %v", err)
	}
	if sets != 4 || reps != 12 || repsDetail != "" || weight != 42.5 || notes != "last set hard" {
		t.Fatalf("parseLoggedExerciseValue() = %d, %d, %q, %.1f, %q; want 4, 12, empty detail, 42.5, notes", sets, reps, repsDetail, weight, notes)
	}
}

func TestParseLoggedExerciseValueRejectsMissingReps(t *testing.T) {
	_, _, _, _, _, err := parseLoggedExerciseValue("4")
	if err == nil {
		t.Fatal("parseLoggedExerciseValue() error = nil; want usage error")
	}
}

func TestParseLoggedExerciseValueSupportsPerSetReps(t *testing.T) {
	sets, reps, repsDetail, weight, notes, err := parseLoggedExerciseValue("6/4 135 second set cooked")
	if err != nil {
		t.Fatalf("parseLoggedExerciseValue() error = %v", err)
	}
	if sets != 2 || reps != 4 || repsDetail != "6/4" || weight != 135 || notes != "second set cooked" {
		t.Fatalf("parseLoggedExerciseValue() = %d, %d, %q, %.1f, %q; want 2, 4, 6/4, 135, notes", sets, reps, repsDetail, weight, notes)
	}
}

func TestFormatSessionEntryShowsPerSetReps(t *testing.T) {
	entry := data.GymSessionEntry{Exercise: "bench", Sets: 2, Reps: 4, RepsDetail: "6/4", Weight: 135}
	got := formatSessionEntry(entry)
	if !strings.Contains(got, "2x6/4") {
		t.Fatalf("formatSessionEntry() = %q; want per-set reps", got)
	}
}

func TestLogEntryInputValuePrefillsEditableLog(t *testing.T) {
	entry := data.GymSessionEntry{Sets: 2, Reps: 4, RepsDetail: "6/4", Weight: 135, Notes: "second set cooked"}
	got := logEntryInputValue(entry)
	if got != "6/4 135 second set cooked" {
		t.Fatalf("logEntryInputValue() = %q; want editable per-set log value", got)
	}
}

func TestExerciseInputValuePrefillsEditableExercise(t *testing.T) {
	exercise := data.Exercise{Name: "Bench Press", Sets: 3, Reps: 8}
	got := exerciseInputValue(exercise)
	if got != "Bench Press 3 8" {
		t.Fatalf("exerciseInputValue() = %q; want editable exercise value", got)
	}
}

func TestParseExerciseValueSupportsNameWithDefaults(t *testing.T) {
	name, sets, reps, err := parseExerciseValue(strings.Fields("Incline Bench 4 10"))
	if err != nil {
		t.Fatalf("parseExerciseValue() error = %v", err)
	}
	if name != "Incline Bench" || sets != 4 || reps != 10 {
		t.Fatalf("parseExerciseValue() = %q, %d, %d; want Incline Bench, 4, 10", name, sets, reps)
	}
}

func TestHelperMessageUsesDotSeparator(t *testing.T) {
	got := helperMessage("up/down move", "enter open program", "a add program")
	if strings.Contains(got, ",") {
		t.Fatalf("helperMessage() = %q; want no commas", got)
	}
	if !strings.Contains(got, " · ") {
		t.Fatalf("helperMessage() = %q; want dot separators", got)
	}
}

func TestIsHelperKeyRecognizesOnlyKeyTokens(t *testing.T) {
	for _, key := range []string{"a", ":", "?", "enter", "esc", "j/k", "up/down", "v"} {
		if !isHelperKey(key) {
			t.Fatalf("isHelperKey(%q) = false; want true", key)
		}
	}

	for _, word := range []string{"type", "suggested", "edit", "command"} {
		if isHelperKey(word) {
			t.Fatalf("isHelperKey(%q) = true; want false", word)
		}
	}
}

func TestNormalKeySupportsVimHistoryScroll(t *testing.T) {
	m := model{
		mode:   modeNormal,
		screen: screenHistory,
		historyEntries: []data.GymSessionEntry{
			{EntryId: 1},
			{EntryId: 2},
		},
	}

	updated, _ := m.handleNormalKey(tea.KeyPressMsg{Code: 'j'})
	got := updated.(model)
	if got.historyCursor != 1 {
		t.Fatalf("historyCursor after j = %d; want 1", got.historyCursor)
	}

	updated, _ = got.handleNormalKey(tea.KeyPressMsg{Code: 'k'})
	got = updated.(model)
	if got.historyCursor != 0 {
		t.Fatalf("historyCursor after k = %d; want 0", got.historyCursor)
	}
}

func TestQuitConfirmationUsesHelperStatus(t *testing.T) {
	m := model{mode: modeNormal}
	m.requestQuit()
	if m.status != helperMessage("quit spotr?", "y confirm", "n cancel") {
		t.Fatalf("requestQuit() status = %q; want helper confirmation", m.status)
	}
}
