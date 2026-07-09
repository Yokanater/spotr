package store

import (
	"database/sql"
	"fmt"
	"strings"
)

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
			reps_detail TEXT NOT NULL DEFAULT '',
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
	if err := s.addGymSessionEntryRepsDetailColumn(); err != nil {
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

func (s *Store) addGymSessionEntryRepsDetailColumn() error {
	hasRepsDetail, err := s.columnExists("gym_session_entries", "reps_detail")
	if err != nil {
		return err
	}
	if hasRepsDetail {
		return nil
	}
	_, err = s.db.Exec(`ALTER TABLE gym_session_entries ADD COLUMN reps_detail TEXT NOT NULL DEFAULT ''`)
	return err
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
