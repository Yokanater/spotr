package store

import (
	"database/sql"
	"fmt"
	"ruffnut/data"
	"ruffnut/utils"
	"strings"
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
	db.SetMaxOpenConns(1)

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
	if err != nil {
		return err
	}

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

	if _, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS gym_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workout_id INTEGER NOT NULL REFERENCES workouts(id),
			started_at TEXT NOT NULL,
			ended_at TEXT,
			notes TEXT NOT NULL DEFAULT ''
		);`); err != nil {
		return err
	}

	if _, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS gym_session_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER NOT NULL REFERENCES gym_sessions(id) ON DELETE CASCADE,
			exercise_id INTEGER NOT NULL REFERENCES exercises(id),
			sets INTEGER NOT NULL,
			reps INTEGER NOT NULL,
			weight REAL NOT NULL DEFAULT 0,
			notes TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		);`); err != nil {
		return err
	}

	if err := s.migrate(); err != nil {
		return err
	}

	return err
}

func (s *Store) migrate() error {
	if err := s.addExerciseDefaultsColumns(); err != nil {
		return err
	}
	if err := s.migrateWorkoutsUniqueConstraint(); err != nil {
		return err
	}
	return nil
}

func (s *Store) addExerciseDefaultsColumns() error {
	hasSets, err := s.columnExists("exercises", "sets")
	if err != nil {
		return err
	}
	if !hasSets {
		if _, err := s.db.Exec(`ALTER TABLE exercises ADD COLUMN sets INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}

	hasReps, err := s.columnExists("exercises", "reps")
	if err != nil {
		return err
	}
	if !hasReps {
		if _, err := s.db.Exec(`ALTER TABLE exercises ADD COLUMN reps INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) columnExists(table string, column string) (bool, error) {
	rows, err := s.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var typ string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func (s *Store) migrateWorkoutsUniqueConstraint() error {
	var schema string
	err := s.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'workouts'`).Scan(&schema)
	if err != nil {
		return err
	}

	if !strings.Contains(schema, "name TEXT NOT NULL UNIQUE") || strings.Contains(schema, "UNIQUE(program_id, name)") {
		return nil
	}

	if _, err := s.db.Exec(`PRAGMA foreign_keys = OFF;`); err != nil {
		return err
	}
	defer s.db.Exec(`PRAGMA foreign_keys = ON;`)

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		CREATE TABLE workouts_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			program_id INTEGER REFERENCES programs(id),
			name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			UNIQUE(program_id, name)
		);`); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO workouts_new (id, program_id, name, created_at)
		SELECT id, program_id, name, created_at FROM workouts;`); err != nil {
		return err
	}

	if _, err := tx.Exec(`DROP TABLE workouts;`); err != nil {
		return err
	}
	if _, err := tx.Exec(`ALTER TABLE workouts_new RENAME TO workouts;`); err != nil {
		return err
	}

	return tx.Commit()
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

func (s *Store) StartGymSession(workout data.Workout) (data.GymSession, error) {
	if existing, err := s.ActiveGymSession(workout); err == nil {
		return existing, fmt.Errorf("active session already started at %s", existing.StartedAt)
	} else if err != sql.ErrNoRows {
		return data.GymSession{}, err
	}

	startedAt := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(
		`INSERT INTO gym_sessions (workout_id, started_at) VALUES (?, ?)`,
		workout.WorkoutId,
		startedAt,
	)
	if err != nil {
		return data.GymSession{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return data.GymSession{}, err
	}
	return data.GymSession{SessionId: id, WorkoutId: workout.WorkoutId, StartedAt: startedAt}, nil
}

func (s *Store) ActiveGymSession(workout data.Workout) (data.GymSession, error) {
	var session data.GymSession
	var endedAt sql.NullString
	err := s.db.QueryRow(
		`SELECT id, workout_id, started_at, ended_at, notes
		FROM gym_sessions
		WHERE workout_id = ? AND ended_at IS NULL
		ORDER BY started_at DESC
		LIMIT 1`,
		workout.WorkoutId,
	).Scan(&session.SessionId, &session.WorkoutId, &session.StartedAt, &endedAt, &session.Notes)
	if err != nil {
		return data.GymSession{}, err
	}
	if endedAt.Valid {
		session.EndedAt = endedAt.String
	}
	return session, nil
}

func (s *Store) FinishGymSession(session data.GymSession, notes string) error {
	endedAt := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`UPDATE gym_sessions SET ended_at = ?, notes = ? WHERE id = ? AND ended_at IS NULL`,
		endedAt,
		notes,
		session.SessionId,
	)
	return err
}

func (s *Store) AddGymSessionEntry(session data.GymSession, exercise data.Exercise, sets int, reps int, weight float64, notes string) error {
	createdAt := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO gym_session_entries (session_id, exercise_id, sets, reps, weight, notes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		session.SessionId,
		exercise.ExerciseId,
		sets,
		reps,
		weight,
		notes,
		createdAt,
	)
	return err
}

func (s *Store) ListGymSessions(workout data.Workout, limit int) ([]data.GymSession, error) {
	if limit <= 0 {
		limit = 10
	}
	sessions := []data.GymSession{}
	rows, err := s.db.Query(
		`SELECT id, workout_id, started_at, ended_at, notes
		FROM gym_sessions
		WHERE workout_id = ?
		ORDER BY started_at DESC
		LIMIT ?`,
		workout.WorkoutId,
		limit,
	)
	if err != nil {
		return sessions, err
	}
	defer rows.Close()

	for rows.Next() {
		var session data.GymSession
		var endedAt sql.NullString
		if err := rows.Scan(&session.SessionId, &session.WorkoutId, &session.StartedAt, &endedAt, &session.Notes); err != nil {
			return sessions, err
		}
		if endedAt.Valid {
			session.EndedAt = endedAt.String
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return sessions, err
	}
	return sessions, nil
}

func (s *Store) SelectGymSession(id string, workout data.Workout) (data.GymSession, error) {
	var session data.GymSession
	var endedAt sql.NullString
	err := s.db.QueryRow(
		`SELECT id, workout_id, started_at, ended_at, notes
		FROM gym_sessions
		WHERE id = ? AND workout_id = ?`,
		id,
		workout.WorkoutId,
	).Scan(&session.SessionId, &session.WorkoutId, &session.StartedAt, &endedAt, &session.Notes)
	if err != nil {
		return data.GymSession{}, err
	}
	if endedAt.Valid {
		session.EndedAt = endedAt.String
	}
	return session, nil
}

func (s *Store) ListGymSessionEntries(session data.GymSession) ([]data.GymSessionEntry, error) {
	entries := []data.GymSessionEntry{}
	rows, err := s.db.Query(
		`SELECT se.id, se.session_id, se.exercise_id, e.name, se.sets, se.reps, se.weight, se.notes
		FROM gym_session_entries se
		JOIN exercises e ON e.id = se.exercise_id
		WHERE se.session_id = ?
		ORDER BY se.id`,
		session.SessionId,
	)
	if err != nil {
		return entries, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry data.GymSessionEntry
		if err := rows.Scan(&entry.EntryId, &entry.SessionId, &entry.ExerciseId, &entry.Exercise, &entry.Sets, &entry.Reps, &entry.Weight, &entry.Notes); err != nil {
			return entries, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return entries, err
	}
	return entries, nil
}
