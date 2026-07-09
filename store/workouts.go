package store

import (
	"database/sql"
	"spotr/data"
	"spotr/utils"
	"time"
)

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

func (s *Store) UpdateWorkout(workout data.Workout, name string) error {
	res, err := s.db.Exec(
		`UPDATE workouts SET name = ? WHERE id = ? AND program_id = ?`,
		name,
		workout.WorkoutId,
		workout.ProgramId,
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

func (s *Store) DeleteWorkout(workout data.Workout) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		DELETE FROM gym_session_entries
		WHERE session_id IN (SELECT id FROM gym_sessions WHERE workout_id = ?)`,
		workout.WorkoutId,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM gym_sessions WHERE workout_id = ?`, workout.WorkoutId); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM exercises WHERE workout_id = ?`, workout.WorkoutId); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM workouts WHERE id = ? AND program_id = ?`, workout.WorkoutId, workout.ProgramId)
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
