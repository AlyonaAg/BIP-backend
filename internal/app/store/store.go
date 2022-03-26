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
	config                 *Config
	db                     *sql.DB
	userRepository         *UserRepository
	orderRepository        *OrderRepository
	photographerRepository *PhotographerRepository
	commentRepository      *CommentRepository
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

func (s *Store) Order() *OrderRepository {
	or, _ := s.GetOrderRepository()
	if or != nil {
		return or
	}

	s.SetOrderRepository(&OrderRepository{
		store: s,
	})

	or, _ = s.GetOrderRepository()
	return or
}

func (s *Store) Photographer() *PhotographerRepository {
	phr, _ := s.GetPhotographerRepository()
	if phr != nil {
		return phr
	}

	s.SetPhotographerRepository(&PhotographerRepository{
		store: s,
	})

	phr, _ = s.GetPhotographerRepository()
	return phr
}

func (s *Store) Comment() *CommentRepository {
	phr, _ := s.GetCommentRepository()
	if phr != nil {
		return phr
	}

	s.SetCommentRepository(&CommentRepository{
		store: s,
	})

	phr, _ = s.GetCommentRepository()
	return phr
}

func (s *Store) Close() error {
	db, err := s.GetDB()
	if err != nil {
		return err
	}
	db.Close()
	return nil
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

func (s *Store) GetOrderRepository() (*OrderRepository, error) {
	if s.orderRepository == nil {
		return nil, errors.New("empty order repository")
	}
	return s.orderRepository, nil
}

func (s *Store) GetPhotographerRepository() (*PhotographerRepository, error) {
	if s.photographerRepository == nil {
		return nil, errors.New("empty photographer repository")
	}
	return s.photographerRepository, nil
}

func (s *Store) GetCommentRepository() (*CommentRepository, error) {
	if s.commentRepository == nil {
		return nil, errors.New("empty comment repository")
	}
	return s.commentRepository, nil
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

func (s *Store) SetOrderRepository(orderRepository *OrderRepository) {
	s.orderRepository = orderRepository
}

func (s *Store) SetPhotographerRepository(photographerRepository *PhotographerRepository) {
	s.photographerRepository = photographerRepository
}

func (s *Store) SetCommentRepository(commentRepository *CommentRepository) {
	s.commentRepository = commentRepository
}

func (s *Store) SetDB(db *sql.DB) {
	s.db = db
}
