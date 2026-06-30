package store

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestNewSQLiteMigratesLegacyExerciseColumns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ruffnut.db")
	db := openLegacyDB(t, path)
	execSQL(t, db, `INSERT INTO programs (name, created_at) VALUES ('ppl', '2026-06-30T00:00:00Z')`)
	execSQL(t, db, `INSERT INTO workouts (program_id, name, created_at) VALUES (1, 'push', '2026-06-30T00:00:00Z')`)
	execSQL(t, db, `INSERT INTO exercises (workout_id, name, created_at) VALUES (1, 'bench', '2026-06-30T00:00:00Z')`)
	closeDB(t, db)

	st, err := NewSQLite(path)
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer st.Close()

	var sets int
	var reps int
	err = st.db.QueryRow(`SELECT sets, reps FROM exercises WHERE name = 'bench'`).Scan(&sets, &reps)
	if err != nil {
		t.Fatalf("query migrated exercise defaults: %v", err)
	}
	if sets != 0 || reps != 0 {
		t.Fatalf("sets, reps = %d, %d; want 0, 0", sets, reps)
	}
}

func TestNewSQLiteMigratesLegacyWorkoutUniqueConstraint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ruffnut.db")
	db := openLegacyDB(t, path)
	closeDB(t, db)

	st, err := NewSQLite(path)
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer st.Close()

	var schema string
	err = st.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'workouts'`).Scan(&schema)
	if err != nil {
		t.Fatalf("query workouts schema: %v", err)
	}
	if !strings.Contains(schema, "UNIQUE(program_id, name)") {
		t.Fatalf("workouts schema = %q; want per-program unique constraint", schema)
	}

	execStoreSQL(t, st, `INSERT INTO programs (name, created_at) VALUES ('ppl', '2026-06-30T00:00:00Z')`)
	execStoreSQL(t, st, `INSERT INTO programs (name, created_at) VALUES ('upper lower', '2026-06-30T00:00:00Z')`)
	execStoreSQL(t, st, `INSERT INTO workouts (program_id, name, created_at) VALUES (1, 'push', '2026-06-30T00:00:00Z')`)
	execStoreSQL(t, st, `INSERT INTO workouts (program_id, name, created_at) VALUES (2, 'push', '2026-06-30T00:00:00Z')`)
}

func openLegacyDB(t *testing.T, path string) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}

	execSQL(t, db, `
		CREATE TABLE schema_version (
			version INTEGER NOT NULL
		);
	`)
	execSQL(t, db, `
		CREATE TABLE programs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL
		);
	`)
	execSQL(t, db, `
		CREATE TABLE workouts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			program_id REFERENCES programs(id),
			name TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL
		);
	`)
	execSQL(t, db, `
		CREATE TABLE exercises (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workout_id REFERENCES workouts(id),
			name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			UNIQUE(workout_id, name)
		);
	`)

	return db
}

func execSQL(t *testing.T, db *sql.DB, query string) {
	t.Helper()

	if _, err := db.Exec(query); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

func closeDB(t *testing.T, db *sql.DB) {
	t.Helper()

	if err := db.Close(); err != nil {
		t.Fatalf("close db: %v", err)
	}
}

func execStoreSQL(t *testing.T, st *Store, query string) {
	t.Helper()

	if _, err := st.db.Exec(query); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}
