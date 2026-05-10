package store

import (
	"database/sql"
	"fmt"
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
		CREATE TABLE IF NOT EXISTS program (
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
	cmd := fmt.Sprintf("INSERT INTO program VALUES (%v, %v)", name, date)
	_, err := s.db.Exec(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) ListPrograms() ([]string, error) {
	rows, err := s.db.Query("SELECT * FROM program")
	programs := []string{};
	for rows.Next() {
		var id int;
		var name string;
		var createdAt string;
		err := rows.Scan(&id, &name, &createdAt)
		if err != nil {
			return programs, nil
		}
		programs = append(programs, name)
	}
	if err != nil {
		return programs, nil
	}
	return programs, nil
}