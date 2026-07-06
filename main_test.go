package main

import (
	"strings"
	"testing"
)

func TestParseLoggedExerciseValue(t *testing.T) {
	sets, reps, weight, notes, err := parseLoggedExerciseValue("4 12 42.5 last set hard")
	if err != nil {
		t.Fatalf("parseLoggedExerciseValue() error = %v", err)
	}
	if sets != 4 || reps != 12 || weight != 42.5 || notes != "last set hard" {
		t.Fatalf("parseLoggedExerciseValue() = %d, %d, %.1f, %q; want 4, 12, 42.5, notes", sets, reps, weight, notes)
	}
}

func TestParseLoggedExerciseValueRejectsMissingReps(t *testing.T) {
	_, _, _, _, err := parseLoggedExerciseValue("4")
	if err == nil {
		t.Fatal("parseLoggedExerciseValue() error = nil; want usage error")
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
