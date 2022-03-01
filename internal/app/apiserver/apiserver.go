package apiserver

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io"
	"log"
)

type Server struct {
	config *Config
	logger *logrus.Logger
	router *gin.Engine
}

func New(config *Config) *Server {
	return &Server{
		config: config,
		logger: logrus.New(),
		router: gin.Default(),
	}
}

func (s *Server) Start() error {
	if err := s.configureLogger(); err != nil {
		return err
	}
	s.configureRouter()
	//s.logger.Info("Starting API-server.")
	log.Print("Starting API-server.")

	return s.router.Run()
}

func (s *Server) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)
	if err != nil {
		return err
	}

	s.logger.SetLevel(level)
	return nil
}

func (s *Server) configureRouter() {
	s.router.GET("/api/app/list", s.handleHello())
}

func (s *Server) handleHello() gin.HandlerFunc {
	//for local var
	return func(c *gin.Context) {
		io.WriteString(c.Writer, "Hello")
	}
}
