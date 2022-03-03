package apiserver

import (
	"BIP_backend/internal/app/store"
	"github.com/gin-gonic/gin"
	"log"
)

type Server struct {
	config *Config
	router *gin.Engine
	store  *store.Store
}

func New(config *Config) *Server {
	return &Server{
		config: config,
		router: gin.Default(),
	}
}

func (s *Server) Start() error {
	s.configureRouter()
	if err := s.configureStore(); err != nil {
		return err
	}

	log.Print("Starting API-server")
	return s.router.Run()
}

func (s *Server) configureRouter() {
	s.router.POST("/api/registration", s.handleUserCreate())
}

func (s *Server) configureStore() error {
	st := store.New(s.config.Store)
	if err := st.Open(); err != nil {
		return err
	}
	s.store = st
	return nil
}
