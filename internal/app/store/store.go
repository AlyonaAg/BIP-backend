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
	config, err := s.GetConfig()
	if err != nil {
		return err
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
	ur, _ := s.GetUserRepository()
	if ur != nil {
		return ur
	}

	s.SetUserRepository(&UserRepository{
		store: s,
	})

	ur, _ = s.GetUserRepository()
	return ur
}

func (s *Store) Close() {
	db, _ := s.GetDB()
	db.Close()
}

func (s *Store) GetConfig() (*Config, error) {
	if s.config == nil {
		return nil, errors.New("empty store config")
	}
	return s.config, nil
}

func (s *Store) GetUserRepository() (*UserRepository, error) {
	if s.userRepository == nil {
		return nil, errors.New("empty user repository")
	}
	return s.userRepository, nil
}

func (s *Store) GetDB() (*sql.DB, error) {
	if s.db == nil {
		return nil, errors.New("empty DB")
	}
	return s.db, nil
}

func (s *Store) SetUserRepository(userRepository *UserRepository) {
	s.userRepository = userRepository
}

func (s *Store) SetDB(db *sql.DB) {
	s.db = db
}
