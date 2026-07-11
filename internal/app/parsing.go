package app

import (
	"fmt"
	"github.com/Yokanater/spotr/data"
	"strconv"
	"strings"
)

func (m *model) parseLogAddArgs(args []string) (data.Exercise, int, int, string, float64, string, error) {
	if len(args) < 1 {
		return data.Exercise{}, 0, 0, "", 0, "", fmt.Errorf("usage: log add [exercise] <sets> <reps> or <reps/reps> [weight] [notes]")
	}

	exercise := m.activeExercise
	valueStart := 0
	if exercise.ExerciseId == 0 || !isLogValueStart(args, 0) {
		if len(args) < 2 {
			return data.Exercise{}, 0, 0, "", 0, "", fmt.Errorf("usage: log add [exercise] <sets> <reps> or <reps/reps> [weight] [notes]")
		}
		valueStart = findFirstLogValue(args)
		if valueStart <= 0 {
			return data.Exercise{}, 0, 0, "", 0, "", fmt.Errorf("sets and reps must be numbers")
		}
		selected, err := m.store.SelectExercise(strings.Join(args[:valueStart], " "), m.activeWorkout)
		if err != nil {
			return data.Exercise{}, 0, 0, "", 0, "", err
		}
		exercise = selected
	}
	if exercise.ExerciseId == 0 {
		return data.Exercise{}, 0, 0, "", 0, "", fmt.Errorf("select an exercise first or pass one to log add")
	}

	sets, reps, repsDetail, weight, notes, err := parseLoggedExerciseValue(strings.Join(args[valueStart:], " "))
	if err != nil {
		return data.Exercise{}, 0, 0, "", 0, "", err
	}
	return exercise, sets, reps, repsDetail, weight, notes, nil
}

func findFirstLogValue(args []string) int {
	for i := range args {
		if isLogValueStart(args, i) {
			return i
		}
	}
	return -1
}

func isLogValueStart(args []string, index int) bool {
	if index >= len(args) {
		return false
	}
	if isValidRepsDetailToken(args[index]) {
		return true
	}
	return index+1 < len(args) && isInt(args[index]) && isInt(args[index+1])
}

func isInt(value string) bool {
	_, err := strconv.Atoi(value)
	return err == nil
}

func parseLoggedExerciseValue(value string) (int, int, string, float64, string, error) {
	args := strings.Fields(value)
	if len(args) < 1 {
		return 0, 0, "", 0, "", fmt.Errorf("usage: sets reps [weight] [notes] or reps/reps [weight] [notes]")
	}

	sets := 0
	reps := 0
	repsDetail := ""
	notesStart := 0
	if isRepsDetailToken(args[0]) {
		parsedReps, err := parseRepsDetail(args[0])
		if err != nil {
			return 0, 0, "", 0, "", err
		}
		sets = len(parsedReps)
		reps = parsedReps[len(parsedReps)-1]
		repsDetail = strings.Join(intStrings(parsedReps), "/")
		notesStart = 1
	} else {
		if len(args) < 2 {
			return 0, 0, "", 0, "", fmt.Errorf("usage: sets reps [weight] [notes] or reps/reps [weight] [notes]")
		}
		var err error
		sets, reps, err = parseSetReps(args[0], args[1])
		if err != nil {
			return 0, 0, "", 0, "", err
		}
		notesStart = 2
	}

	if sets <= 0 {
		return 0, 0, "", 0, "", fmt.Errorf("sets must be greater than zero")
	}
	if reps <= 0 {
		return 0, 0, "", 0, "", fmt.Errorf("reps must be greater than zero")
	}

	weight := 0.0
	if len(args) > notesStart {
		parsedWeight, err := strconv.ParseFloat(args[notesStart], 64)
		if err == nil {
			weight = parsedWeight
			notesStart++
		}
	}

	notes := ""
	if len(args) > notesStart {
		notes = strings.Join(args[notesStart:], " ")
	}
	return sets, reps, repsDetail, weight, notes, nil
}

func isRepsDetailToken(value string) bool {
	return strings.Contains(value, "/")
}

func isValidRepsDetailToken(value string) bool {
	_, err := parseRepsDetail(value)
	return err == nil
}

func parseRepsDetail(value string) ([]int, error) {
	parts := strings.Split(value, "/")
	if len(parts) < 2 {
		return nil, fmt.Errorf("use reps/reps for per-set reps")
	}
	reps := make([]int, 0, len(parts))
	for _, part := range parts {
		parsed, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("per-set reps must be numbers")
		}
		if parsed <= 0 {
			return nil, fmt.Errorf("per-set reps must be greater than zero")
		}
		reps = append(reps, parsed)
	}
	return reps, nil
}

func intStrings(values []int) []string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.Itoa(value))
	}
	return parts
}

func formatSessionList(sessions []data.GymSession) string {
	if len(sessions) == 0 {
		return "no sessions yet"
	}
	parts := make([]string, 0, len(sessions))
	for _, session := range sessions {
		state := "active"
		if session.EndedAt != "" {
			state = "done"
		}
		parts = append(parts, fmt.Sprintf("#%d %s %s", session.SessionId, state, session.StartedAt))
	}
	return strings.Join(parts, " | ")
}

func formatSessionDetail(session data.GymSession, entries []data.GymSessionEntry) string {
	if len(entries) == 0 {
		return fmt.Sprintf("Session #%d has no entries", session.SessionId)
	}
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		parts = append(parts, formatSessionEntry(entry))
	}
	return fmt.Sprintf("Session #%d: %s", session.SessionId, strings.Join(parts, " | "))
}

func formatSessionEntry(entry data.GymSessionEntry) string {
	value := fmt.Sprintf("%s %s", entry.Exercise, setRepLabel(entry))
	if entry.Weight > 0 {
		value += fmt.Sprintf(" @ %.1f", entry.Weight)
	}
	if entry.Notes != "" {
		value += " (" + entry.Notes + ")"
	}
	return value
}

func setRepLabel(entry data.GymSessionEntry) string {
	if entry.RepsDetail != "" {
		return fmt.Sprintf("%dx%s", entry.Sets, entry.RepsDetail)
	}
	return fmt.Sprintf("%dx%d", entry.Sets, entry.Reps)
}

func logEntryInputValue(entry data.GymSessionEntry) string {
	parts := []string{}
	if entry.RepsDetail != "" {
		parts = append(parts, entry.RepsDetail)
	} else {
		parts = append(parts, strconv.Itoa(entry.Sets), strconv.Itoa(entry.Reps))
	}
	if entry.Weight > 0 {
		parts = append(parts, strconv.FormatFloat(entry.Weight, 'f', -1, 64))
	}
	if entry.Notes != "" {
		parts = append(parts, entry.Notes)
	}
	return strings.Join(parts, " ")
}

func exerciseInputValue(exercise data.Exercise) string {
	parts := []string{exercise.Name}
	if exercise.Sets > 0 || exercise.Reps > 0 {
		parts = append(parts, strconv.Itoa(exercise.Sets), strconv.Itoa(exercise.Reps))
	}
	return strings.Join(parts, " ")
}

func exerciseLabelForStatus(exercise data.Exercise) string {
	if exercise.Sets > 0 || exercise.Reps > 0 {
		return fmt.Sprintf("%s %dx%d", exercise.Name, exercise.Sets, exercise.Reps)
	}
	return exercise.Name
}

func parseSetReps(setsArg string, repsArg string) (int, int, error) {
	sets, err := strconv.Atoi(setsArg)
	if err != nil {
		return 0, 0, fmt.Errorf("sets must be a number")
	}
	reps, err := strconv.Atoi(repsArg)
	if err != nil {
		return 0, 0, fmt.Errorf("reps must be a number")
	}
	return sets, reps, nil
}

func parseExerciseAddArgs(args []string) (string, int, int, error) {
	if len(args) < 2 {
		return "", 0, 0, fmt.Errorf("exercise name is required")
	}
	return parseExerciseValue(args[1:])
}

func parseExerciseValue(args []string) (string, int, int, error) {
	nameEnd := len(args)
	sets := 0
	reps := 0

	if len(args) >= 3 {
		parsedSets, err := strconv.Atoi(args[len(args)-2])
		if err != nil {
			return "", 0, 0, fmt.Errorf("sets must be a number")
		}
		parsedReps, err := strconv.Atoi(args[len(args)-1])
		if err != nil {
			return "", 0, 0, fmt.Errorf("reps must be a number")
		}
		nameEnd = len(args) - 2
		sets = parsedSets
		reps = parsedReps
	}

	name := strings.Join(args[:nameEnd], " ")
	if strings.TrimSpace(name) == "" {
		return "", 0, 0, fmt.Errorf("exercise name is required")
	}

	return name, sets, reps, nil
}
