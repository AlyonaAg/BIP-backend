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

	router := s.GetRouter()
	if router == nil {
		return errors.New("empty router")
	}
	return router.Run()
}

func (s *Server) openStore() error {
	store := s.GetStore()
	if store == nil {
		return errors.New("empty store")
	}

	if err := store.Open(); err != nil {
		return err
	}
	return nil
}

func (s *Server) configureRouter() error {
	router := s.GetRouter()
	if router == nil {
		return errors.New("empty router")
	}

	api := router.Group("/api")
	{
		api.POST("/registration", s.handleUserCreate())
	}
	return nil
}

func (s *Server) GetConfig() *Config {
	return s.config
}

func (s *Server) GetRouter() *gin.Engine {
	return s.router
}

func (s *Server) GetStore() *store.Store {
	return s.store
}
