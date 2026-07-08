package main

import (
	"ruffnut/data"
	"strings"
	"testing"
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
	for _, key := range []string{"a", ":", "?", "enter", "esc", "up/down", "v"} {
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
