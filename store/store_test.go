package store

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"ruffnut/data"

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

func TestGymSessionLifecycle(t *testing.T) {
	st, err := NewSQLite(filepath.Join(t.TempDir(), "ruffnut.db"))
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer st.Close()

	programID, err := st.CreateProgram("ppl")
	if err != nil {
		t.Fatalf("CreateProgram() error = %v", err)
	}
	program := data.Program{ProgramId: programID, ProgramName: "ppl"}
	if err := st.CreateWorkout("push", program); err != nil {
		t.Fatalf("CreateWorkout() error = %v", err)
	}
	workout, err := st.SelectWorkout("push", program)
	if err != nil {
		t.Fatalf("SelectWorkout() error = %v", err)
	}
	if err := st.CreateExercise("bench", 3, 10, workout); err != nil {
		t.Fatalf("CreateExercise() error = %v", err)
	}
	exercise, err := st.SelectExercise("bench", workout)
	if err != nil {
		t.Fatalf("SelectExercise() error = %v", err)
	}

	session, err := st.StartGymSession(workout)
	if err != nil {
		t.Fatalf("StartGymSession() error = %v", err)
	}
	if session.SessionId == 0 || session.WorkoutId != workout.WorkoutId {
		t.Fatalf("StartGymSession() = %+v; want persisted workout session", session)
	}
	if _, err := st.StartGymSession(workout); err == nil || !strings.Contains(err.Error(), "active session already started") {
		t.Fatalf("StartGymSession() duplicate error = %v; want active session error", err)
	}

	if err := st.AddGymSessionEntry(session, exercise, 3, 8, "", 135, "felt good"); err != nil {
		t.Fatalf("AddGymSessionEntry() error = %v", err)
	}
	if err := st.AddGymSessionEntry(session, exercise, 2, 4, "6/4", 135, "dropoff"); err != nil {
		t.Fatalf("AddGymSessionEntry() per-set error = %v", err)
	}
	entries, err := st.ListGymSessionEntries(session)
	if err != nil {
		t.Fatalf("ListGymSessionEntries() error = %v", err)
	}
	if len(entries) != 2 || entries[0].Exercise != "bench" || entries[0].Sets != 3 || entries[0].Reps != 8 || entries[0].Weight != 135 {
		t.Fatalf("ListGymSessionEntries() = %+v; want bench 3x8 @ 135", entries)
	}
	if entries[1].RepsDetail != "6/4" {
		t.Fatalf("ListGymSessionEntries() reps detail = %q; want 6/4", entries[1].RepsDetail)
	}

	if err := st.UpdateGymSessionEntry(entries[0], 3, 9, "", 140, "moved well"); err != nil {
		t.Fatalf("UpdateGymSessionEntry() error = %v", err)
	}
	entries, err = st.ListGymSessionEntries(session)
	if err != nil {
		t.Fatalf("ListGymSessionEntries() after update error = %v", err)
	}
	if entries[0].Sets != 3 || entries[0].Reps != 9 || entries[0].Weight != 140 || entries[0].Notes != "moved well" {
		t.Fatalf("updated entry = %+v; want 3x9 @ 140 with notes", entries[0])
	}

	if err := st.DeleteGymSessionEntry(entries[1]); err != nil {
		t.Fatalf("DeleteGymSessionEntry() error = %v", err)
	}
	entries, err = st.ListGymSessionEntries(session)
	if err != nil {
		t.Fatalf("ListGymSessionEntries() after delete error = %v", err)
	}
	if len(entries) != 1 || entries[0].RepsDetail != "" {
		t.Fatalf("entries after delete = %+v; want only updated straight-set entry", entries)
	}

	if err := st.FinishGymSession(session, "solid push day"); err != nil {
		t.Fatalf("FinishGymSession() error = %v", err)
	}
	sessions, err := st.ListGymSessions(workout, 10)
	if err != nil {
		t.Fatalf("ListGymSessions() error = %v", err)
	}
	if len(sessions) != 1 || sessions[0].EndedAt == "" || sessions[0].Notes != "solid push day" {
		t.Fatalf("ListGymSessions() = %+v; want finished session with notes", sessions)
	}

	if err := st.CreateWorkout("upper body", program); err != nil {
		t.Fatalf("CreateWorkout() upper body error = %v", err)
	}
	upperWorkout, err := st.SelectWorkout("upper body", program)
	if err != nil {
		t.Fatalf("SelectWorkout() upper body error = %v", err)
	}
	if err := st.CreateExercise("bench", 3, 8, upperWorkout); err != nil {
		t.Fatalf("CreateExercise() upper bench error = %v", err)
	}
	upperBench, err := st.SelectExercise("bench", upperWorkout)
	if err != nil {
		t.Fatalf("SelectExercise() upper bench error = %v", err)
	}
	upperSession, err := st.StartGymSession(upperWorkout)
	if err != nil {
		t.Fatalf("StartGymSession() upper error = %v", err)
	}
	if err := st.AddGymSessionEntry(upperSession, upperBench, 3, 7, "", 140, "upper day"); err != nil {
		t.Fatalf("AddGymSessionEntry() upper error = %v", err)
	}

	linkedEntries, err := st.ListExerciseLogEntries(program, "bench", 10)
	if err != nil {
		t.Fatalf("ListExerciseLogEntries() error = %v", err)
	}
	if len(linkedEntries) != 2 {
		t.Fatalf("ListExerciseLogEntries() = %+v; want two linked bench entries after delete", linkedEntries)
	}
	if linkedEntries[0].Workout != "upper body" || linkedEntries[1].Workout != "push" {
		t.Fatalf("ListExerciseLogEntries() workouts = %+v; want upper body then push", linkedEntries)
	}
}

func TestUpdateAndDeleteProgram(t *testing.T) {
	st, err := NewSQLite(filepath.Join(t.TempDir(), "ruffnut.db"))
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer st.Close()

	programID, err := st.CreateProgram("ppl")
	if err != nil {
		t.Fatalf("CreateProgram() error = %v", err)
	}
	program := data.Program{ProgramId: programID, ProgramName: "ppl"}
	if err := st.CreateWorkout("push", program); err != nil {
		t.Fatalf("CreateWorkout() error = %v", err)
	}
	workout, err := st.SelectWorkout("push", program)
	if err != nil {
		t.Fatalf("SelectWorkout() error = %v", err)
	}
	if err := st.CreateExercise("bench", 3, 8, workout); err != nil {
		t.Fatalf("CreateExercise() error = %v", err)
	}
	exercise, err := st.SelectExercise("bench", workout)
	if err != nil {
		t.Fatalf("SelectExercise() error = %v", err)
	}
	session, err := st.StartGymSession(workout)
	if err != nil {
		t.Fatalf("StartGymSession() error = %v", err)
	}
	if err := st.AddGymSessionEntry(session, exercise, 3, 8, "", 135, ""); err != nil {
		t.Fatalf("AddGymSessionEntry() error = %v", err)
	}

	if err := st.UpdateProgram(program, "powerbuilding"); err != nil {
		t.Fatalf("UpdateProgram() error = %v", err)
	}
	renamed, err := st.SelectProgram("powerbuilding")
	if err != nil {
		t.Fatalf("SelectProgram() renamed error = %v", err)
	}
	if renamed.ProgramId != program.ProgramId {
		t.Fatalf("renamed program id = %d; want %d", renamed.ProgramId, program.ProgramId)
	}

	if err := st.DeleteProgram(renamed); err != nil {
		t.Fatalf("DeleteProgram() error = %v", err)
	}
	programs, err := st.ListPrograms()
	if err != nil {
		t.Fatalf("ListPrograms() error = %v", err)
	}
	if len(programs) != 0 {
		t.Fatalf("ListPrograms() = %+v; want deleted program removed", programs)
	}
	var sessionCount int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM gym_sessions`).Scan(&sessionCount); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if sessionCount != 0 {
		t.Fatalf("session count = %d; want cascade cleanup", sessionCount)
	}
}

func TestUpdateAndDeleteExercise(t *testing.T) {
	st, err := NewSQLite(filepath.Join(t.TempDir(), "ruffnut.db"))
	if err != nil {
		t.Fatalf("NewSQLite() error = %v", err)
	}
	defer st.Close()

	programID, err := st.CreateProgram("ppl")
	if err != nil {
		t.Fatalf("CreateProgram() error = %v", err)
	}
	program := data.Program{ProgramId: programID, ProgramName: "ppl"}
	if err := st.CreateWorkout("push", program); err != nil {
		t.Fatalf("CreateWorkout() error = %v", err)
	}
	workout, err := st.SelectWorkout("push", program)
	if err != nil {
		t.Fatalf("SelectWorkout() error = %v", err)
	}
	if err := st.CreateExercise("bench", 3, 8, workout); err != nil {
		t.Fatalf("CreateExercise() error = %v", err)
	}
	exercise, err := st.SelectExercise("bench", workout)
	if err != nil {
		t.Fatalf("SelectExercise() error = %v", err)
	}
	session, err := st.StartGymSession(workout)
	if err != nil {
		t.Fatalf("StartGymSession() error = %v", err)
	}
	if err := st.AddGymSessionEntry(session, exercise, 3, 8, "", 135, ""); err != nil {
		t.Fatalf("AddGymSessionEntry() error = %v", err)
	}

	if err := st.UpdateExercise(exercise, "barbell bench", 4, 6); err != nil {
		t.Fatalf("UpdateExercise() error = %v", err)
	}
	renamed, err := st.SelectExercise("barbell bench", workout)
	if err != nil {
		t.Fatalf("SelectExercise() renamed error = %v", err)
	}
	if renamed.ExerciseId != exercise.ExerciseId || renamed.Sets != 4 || renamed.Reps != 6 {
		t.Fatalf("renamed exercise = %+v; want same id with 4x6 defaults", renamed)
	}

	if err := st.DeleteExercise(renamed); err != nil {
		t.Fatalf("DeleteExercise() error = %v", err)
	}
	exercises, err := st.ListExercises(workout)
	if err != nil {
		t.Fatalf("ListExercises() error = %v", err)
	}
	if len(exercises) != 0 {
		t.Fatalf("ListExercises() = %+v; want deleted exercise removed", exercises)
	}
	entries, err := st.ListGymSessionEntries(session)
	if err != nil {
		t.Fatalf("ListGymSessionEntries() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("ListGymSessionEntries() = %+v; want deleted exercise logs removed", entries)
	}
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
