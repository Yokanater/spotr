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
		);`); 
		err != nil {
		return err
	}

	return err
}

func (s *Store) CreateProgram(name string) error {
	date := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec("INSERT INTO programs (name, created_at) VALUES (?, ?)", name, date)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) ListPrograms() ([]string, error) {
	programs := []string{}

	rows, err := s.db.Query("SELECT name FROM programs ORDER BY name")

	if err != nil {
		return programs, err
	}
	defer rows.Close()

	for rows.Next() {
		var name string;

		err := rows.Scan(&name)

		if err != nil {
			return programs, err
		}
		programs = append(programs, name)
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