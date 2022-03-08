package apiserver

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"

	"BIP_backend/internal/app/store"
)

type Server struct {
	config *Config
	router *gin.Engine
	store  *store.Store
}

func NewServer(config *Config) *Server {
	return &Server{
		config: config,
		router: gin.Default(),
		store:  store.NewStore(config.Store),
	}
}

func (s *Server) Start() error {
	if err := s.openStore(); err != nil {
		return err
	}
	if err := s.configureRouter(); err != nil {
		return err
	}

	log.Print("Starting API-server.")

	router, err := s.GetRouter()
	if err != nil {
		return err
	}

	return router.Run()
}

func (s *Server) openStore() error {
	store, err := s.GetStore()
	if err != nil {
		return errors.New("empty store")
	}

	if err := store.Open(); err != nil {
		return err
	}
	return nil
}

func (s *Server) configureRouter() error {
	router, err := s.GetRouter()
	if err != nil {
		return err
	}

	api := router.Group("/api")
	{
		api.POST("/registration", s.handleUserCreate())
		api.POST("/auth", s.handleSessionsCreate())
	}
	return nil
}

func (s *Server) GetConfig() (*Config, error) {
	if s.config == nil {
		return nil, errors.New("empty config")
	}
	return s.config, nil
}

func (s *Server) GetRouter() (*gin.Engine, error) {
	if s.router == nil {
		return nil, errors.New("empty router")
	}
	return s.router, nil
}

func (s *Server) GetStore() (*store.Store, error) {
	if s.store == nil {
		return nil, errors.New("empty store")
	}
	return s.store, nil
}
