package store

import (
	"database/sql"
	"fmt"
	"ruffnut/data"
	"time"
)

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

func (s *Store) AddGymSessionEntry(session data.GymSession, exercise data.Exercise, sets int, reps int, repsDetail string, weight float64, notes string) error {
	createdAt := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO gym_session_entries (session_id, exercise_id, sets, reps, reps_detail, weight, notes, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		session.SessionId,
		exercise.ExerciseId,
		sets,
		reps,
		repsDetail,
		weight,
		notes,
		createdAt,
	)
	return err
}

func (s *Store) UpdateGymSessionEntry(entry data.GymSessionEntry, sets int, reps int, repsDetail string, weight float64, notes string) error {
	res, err := s.db.Exec(
		`UPDATE gym_session_entries
		SET sets = ?, reps = ?, reps_detail = ?, weight = ?, notes = ?
		WHERE id = ?`,
		sets,
		reps,
		repsDetail,
		weight,
		notes,
		entry.EntryId,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) DeleteGymSessionEntry(entry data.GymSessionEntry) error {
	res, err := s.db.Exec(`DELETE FROM gym_session_entries WHERE id = ?`, entry.EntryId)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
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

func (s *Store) SelectGymSessionByID(id int64) (data.GymSession, error) {
	var session data.GymSession
	var endedAt sql.NullString
	err := s.db.QueryRow(
		`SELECT id, workout_id, started_at, ended_at, notes
		FROM gym_sessions
		WHERE id = ?`,
		id,
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
		`SELECT se.id, se.session_id, se.exercise_id, e.name, se.sets, se.reps, se.reps_detail, se.weight, se.notes
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
		if err := rows.Scan(&entry.EntryId, &entry.SessionId, &entry.ExerciseId, &entry.Exercise, &entry.Sets, &entry.Reps, &entry.RepsDetail, &entry.Weight, &entry.Notes); err != nil {
			return entries, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return entries, err
	}
	return entries, nil
}

func (s *Store) SelectGymSessionEntry(id string, workout data.Workout) (data.GymSessionEntry, error) {
	var entry data.GymSessionEntry
	err := s.db.QueryRow(
		`SELECT se.id, se.session_id, se.exercise_id, e.name, w.name, gs.started_at, se.sets, se.reps, se.reps_detail, se.weight, se.notes
		FROM gym_session_entries se
		JOIN gym_sessions gs ON gs.id = se.session_id
		JOIN exercises e ON e.id = se.exercise_id
		JOIN workouts w ON w.id = gs.workout_id
		WHERE se.id = ? AND gs.workout_id = ?`,
		id,
		workout.WorkoutId,
	).Scan(&entry.EntryId, &entry.SessionId, &entry.ExerciseId, &entry.Exercise, &entry.Workout, &entry.StartedAt, &entry.Sets, &entry.Reps, &entry.RepsDetail, &entry.Weight, &entry.Notes)
	if err != nil {
		return data.GymSessionEntry{}, err
	}
	return entry, nil
}

func (s *Store) ListExerciseLogEntries(program data.Program, exerciseName string, limit int) ([]data.GymSessionEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	entries := []data.GymSessionEntry{}
	rows, err := s.db.Query(
		`SELECT se.id, se.session_id, se.exercise_id, e.name, w.name, gs.started_at, se.sets, se.reps, se.reps_detail, se.weight, se.notes
		FROM gym_session_entries se
		JOIN gym_sessions gs ON gs.id = se.session_id
		JOIN exercises e ON e.id = se.exercise_id
		JOIN workouts w ON w.id = gs.workout_id
		WHERE w.program_id = ? AND lower(e.name) = lower(?)
		ORDER BY gs.started_at DESC, se.id DESC
		LIMIT ?`,
		program.ProgramId,
		exerciseName,
		limit,
	)
	if err != nil {
		return entries, err
	}
	defer rows.Close()

	for rows.Next() {
		var entry data.GymSessionEntry
		if err := rows.Scan(&entry.EntryId, &entry.SessionId, &entry.ExerciseId, &entry.Exercise, &entry.Workout, &entry.StartedAt, &entry.Sets, &entry.Reps, &entry.RepsDetail, &entry.Weight, &entry.Notes); err != nil {
			return entries, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return entries, err
	}
	return entries, nil
}
