package apiserver

import (
	"github.com/gin-gonic/gin"
	"log"
)

type Server struct {
	config *Config
	router *gin.Engine
}

func New(config *Config) *Server {
	return &Server{
		config: config,
		router: gin.Default(),
	}
}

func (s *Server) Start() error {
	s.configureRouter()
	log.Print("Starting API-server.")
	return s.router.Run()
}

func (s *Server) configureRouter() {
	s.router.GET("/api/app/list", s.handleHello())
	s.router.POST("/api/registration", s.handleRegistration())
}
