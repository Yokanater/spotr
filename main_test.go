package main

import "testing"

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
