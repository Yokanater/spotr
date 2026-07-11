package store

import (
	"github.com/Yokanater/spotr/data"
)

func (s *Store) ActiveProgram() (data.Program, error) {
	var program data.Program
	err := s.db.QueryRow(`
		SELECT p.id, p.name
		FROM app_preferences pref
		JOIN programs p ON p.id = pref.active_program_id
		WHERE pref.id = 1
	`).Scan(&program.ProgramId, &program.ProgramName)
	return program, err
}

func (s *Store) SetActiveProgram(program data.Program) error {
	_, err := s.db.Exec(`UPDATE app_preferences SET active_program_id = ? WHERE id = 1`, program.ProgramId)
	return err
}

func (s *Store) ClearActiveProgram() error {
	_, err := s.db.Exec(`UPDATE app_preferences SET active_program_id = NULL WHERE id = 1`)
	return err
}
