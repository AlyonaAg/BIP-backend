package store

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

type Store struct {
	config *Config
	db     *sql.DB
}

func New(config *Config) *Store {
	return &Store{
		config: config,
	}
}

func (s *Store) Open() error {
	db, err := sql.Open("postgres", s.config.DatabaseURL)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	log.Print("Store OK.")
	s.db = db
	return nil
}

func (s *Store) Close() {
	s.db.Close()
}
