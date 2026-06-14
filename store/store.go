package store

import (
	"database/sql"
	"fmt"
	"ruffnut/data"
	"ruffnut/utils"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func NewSQLite(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

func (s *Store) init() error {
	if s.db == nil {
		return fmt.Errorf("store: nil db")
	}
	if _, err := s.db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return err
	}

	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER NOT NULL
		);
	`)

	if _, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS programs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL
		);`); err != nil {
		return err
	}

	if _, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS workouts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			program_id REFERENCES programs(id),
			name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			UNIQUE(program_id, name)
		);`); err != nil {
		return err
	}

	if _, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS exercises (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workout_id REFERENCES workouts(id),
			name TEXT NOT NULL,
			sets INTEGER NOT NULL DEFAULT 0,
			reps INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			UNIQUE(workout_id, name)
		);`); err != nil {
		return err
	}

	return err
}

func (s *Store) CreateProgram(name string) (int64, error) {
	date := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec("INSERT INTO programs (name, created_at) VALUES (?, ?)", name, date)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return id, err
	}
	return id, nil
}

func (s *Store) ListPrograms() ([]data.Program, error) {
	programs := []data.Program{}

	rows, err := s.db.Query("SELECT id, name FROM programs ORDER BY name")

	if err != nil {
		return programs, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var id int64

		err := rows.Scan(&id, &name)

		if err != nil {
			return programs, err
		}
		programs = append(programs, data.Program{ProgramId: id, ProgramName: name})
	}

	err = rows.Err()
	if err != nil {
		return programs, err
	}
	return programs, nil
}

func (s *Store) SelectProgram(arg string) (data.Program, error) {
	var progId int64 = 0
	var progName string = ""
	if utils.DigitCheck.MatchString(arg) {
		err := s.db.QueryRow(`SELECT id, name FROM programs WHERE id = ?`, arg).Scan(&progId, &progName)
		if err != nil {
			return data.Program{}, err
		}
		program := data.Program{ProgramId: progId, ProgramName: progName}
		return program, err
	}
	err := s.db.QueryRow(`SELECT id, name FROM programs WHERE name = ?`, arg).Scan(&progId, &progName)

	if err != nil {
		return data.Program{}, err
	}

	program := data.Program{ProgramId: progId, ProgramName: progName}
	return program, err
}

func (s *Store) CreateWorkout(name string, program data.Program) error {
	date := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec("INSERT INTO workouts (program_id, name, created_at) VALUES (?, ?, ?)", program.ProgramId, name, date)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) SelectWorkout(arg string, program data.Program) (data.Workout, error) {
	var workoutId int64 = 0
	var workoutName string = ""
	if utils.DigitCheck.MatchString(arg) {
		err := s.db.QueryRow(
			`SELECT id, name FROM workouts WHERE id = ? AND program_id = ?`,
			arg,
			program.ProgramId,
		).Scan(&workoutId, &workoutName)
		if err != nil {
			return data.Workout{}, err
		}
		return data.Workout{WorkoutId: workoutId, ProgramId: program.ProgramId, Name: workoutName}, nil
	}

	err := s.db.QueryRow(
		`SELECT id, name FROM workouts WHERE name = ? AND program_id = ?`,
		arg,
		program.ProgramId,
	).Scan(&workoutId, &workoutName)
	if err != nil {
		return data.Workout{}, err
	}

	return data.Workout{WorkoutId: workoutId, ProgramId: program.ProgramId, Name: workoutName}, nil
}

func (s *Store) ListWorkouts(program data.Program) ([]data.Workout, error) {
	workouts := []data.Workout{}

	rows, err := s.db.Query("SELECT id, name FROM workouts WHERE program_id=? ORDER BY name", program.ProgramId)

	if err != nil {
		return workouts, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string

		err := rows.Scan(&id, &name)

		if err != nil {
			return workouts, err
		}
		workouts = append(workouts, data.Workout{
			WorkoutId: id,
			ProgramId: program.ProgramId,
			Name:      name,
		})
	}

	err = rows.Err()
	if err != nil {
		return workouts, err
	}
	return workouts, nil
}

func (s *Store) CreateExercise(name string, sets int, reps int, workout data.Workout) error {
	date := time.Now().UTC().Format(time.RFC3339)

	_, err := s.db.Exec(
		"INSERT INTO exercises (workout_id, name, sets, reps, created_at) VALUES (?, ?, ?, ?, ?)",
		workout.WorkoutId,
		name,
		sets,
		reps,
		date,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ListExercises(workout data.Workout) ([]data.Exercise, error) {
	exercises := []data.Exercise{}

	rows, err := s.db.Query("SELECT id, name, sets, reps FROM exercises WHERE workout_id=? ORDER BY name", workout.WorkoutId)
	if err != nil {
		return exercises, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string
		var sets int
		var reps int

		err := rows.Scan(&id, &name, &sets, &reps)
		if err != nil {
			return exercises, err
		}
		exercises = append(exercises, data.Exercise{
			ExerciseId: id,
			WorkoutId:  workout.WorkoutId,
			Name:       name,
			Sets:       sets,
			Reps:       reps,
		})
	}

	err = rows.Err()
	if err != nil {
		return exercises, err
	}
	return exercises, nil
}

func (s *Store) SelectExercise(arg string, workout data.Workout) (data.Exercise, error) {
	var exercise data.Exercise
	if utils.DigitCheck.MatchString(arg) {
		err := s.db.QueryRow(
			`SELECT id, name, sets, reps FROM exercises WHERE id = ? AND workout_id = ?`,
			arg,
			workout.WorkoutId,
		).Scan(&exercise.ExerciseId, &exercise.Name, &exercise.Sets, &exercise.Reps)
		if err != nil {
			return data.Exercise{}, err
		}
		exercise.WorkoutId = workout.WorkoutId
		return exercise, nil
	}

	err := s.db.QueryRow(
		`SELECT id, name, sets, reps FROM exercises WHERE name = ? AND workout_id = ?`,
		arg,
		workout.WorkoutId,
	).Scan(&exercise.ExerciseId, &exercise.Name, &exercise.Sets, &exercise.Reps)
	if err != nil {
		return data.Exercise{}, err
	}
	exercise.WorkoutId = workout.WorkoutId
	return exercise, nil
}

func (s *Store) UpdateExerciseDefaults(exercise data.Exercise, sets int, reps int) error {
	_, err := s.db.Exec(
		`UPDATE exercises SET sets = ?, reps = ? WHERE id = ? AND workout_id = ?`,
		sets,
		reps,
		exercise.ExerciseId,
		exercise.WorkoutId,
	)
	return err
}
