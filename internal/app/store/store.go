package store

import (
	"database/sql"
	"errors"
	"log"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/lib/pq"
)

type Store struct {
	config         *Config
	db             *sql.DB
	userRepository *UserRepository
}

func NewStore(config *Config) *Store {
	return &Store{
		config: config,
	}
}

func (s *Store) Open() error {
	config := s.GetConfig()
	if config == nil {
		return errors.New("empty store config")
	}
	databaseURL := config.DatabaseURL

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	pathMigration := config.PathMigration
	m, err := migrate.NewWithDatabaseInstance(pathMigration, "postgres", driver)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	s.SetDB(db)

	log.Print("Store OK.")
	return nil
}

func (s *Store) User() *UserRepository {
	if s.GetUserRepository() != nil {
		return s.GetUserRepository()
	}

	s.SetUserRepository(&UserRepository{
		store: s,
	})
	return s.GetUserRepository()
}

func (s *Store) Close() {
	s.GetDB().Close()
}

func (s *Store) GetConfig() *Config {
	return s.config
}

func (s *Store) GetUserRepository() *UserRepository {
	return s.userRepository
}

func (s *Store) GetDB() *sql.DB {
	return s.db
}

func (s *Store) SetUserRepository(userRepository *UserRepository) {
	s.userRepository = userRepository
}

func (s *Store) SetDB(db *sql.DB) {
	s.db = db
}
