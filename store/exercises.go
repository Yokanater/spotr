package store

import (
	"database/sql"
	"ruffnut/data"
	"ruffnut/utils"
	"time"
)

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

func (s *Store) UpdateExercise(exercise data.Exercise, name string, sets int, reps int) error {
	res, err := s.db.Exec(
		`UPDATE exercises SET name = ?, sets = ?, reps = ? WHERE id = ? AND workout_id = ?`,
		name,
		sets,
		reps,
		exercise.ExerciseId,
		exercise.WorkoutId,
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

func (s *Store) DeleteExercise(exercise data.Exercise) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM gym_session_entries WHERE exercise_id = ?`, exercise.ExerciseId); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM exercises WHERE id = ? AND workout_id = ?`, exercise.ExerciseId, exercise.WorkoutId)
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
	return tx.Commit()
}
