package data

type GymSession struct {
	SessionId int64
	WorkoutId int64
	StartedAt string
	EndedAt   string
	Notes     string
}

type GymSessionEntry struct {
	EntryId    int64
	SessionId  int64
	ExerciseId int64
	Exercise   string
	Workout    string
	StartedAt  string
	Sets       int
	Reps       int
	RepsDetail string
	Weight     float64
	Notes      string
}
