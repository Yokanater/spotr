package store

import (
	"database/sql"
	"github.com/Yokanater/spotr/data"
	"github.com/Yokanater/spotr/utils"
	"time"
)

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

func (s *Store) UpdateProgram(program data.Program, name string) error {
	res, err := s.db.Exec(`UPDATE programs SET name = ? WHERE id = ?`, name, program.ProgramId)
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

func (s *Store) DeleteProgram(program data.Program) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		DELETE FROM gym_session_entries
		WHERE session_id IN (
			SELECT gs.id
			FROM gym_sessions gs
			JOIN workouts w ON w.id = gs.workout_id
			WHERE w.program_id = ?
		)`,
		program.ProgramId,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		DELETE FROM gym_sessions
		WHERE workout_id IN (SELECT id FROM workouts WHERE program_id = ?)`,
		program.ProgramId,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		DELETE FROM exercises
		WHERE workout_id IN (SELECT id FROM workouts WHERE program_id = ?)`,
		program.ProgramId,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM workouts WHERE program_id = ?`, program.ProgramId); err != nil {
		return err
	}
	res, err := tx.Exec(`DELETE FROM programs WHERE id = ?`, program.ProgramId)
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
