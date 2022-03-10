package apiserver

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"

	"BIP_backend/internal/app/cache"
	"BIP_backend/internal/app/store"
	"BIP_backend/middleware"
)

type Server struct {
	config *Config
	router *gin.Engine
	store  *store.Store
	cache  *cache.Cache
}

func NewServer() (*Server, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, err
	}

	return &Server{
		config: config,
		router: gin.Default(),
		store:  store.NewStore(config.Store),
		cache:  cache.NewCache(config.Cache),
	}, nil
}

func (s *Server) Start() error {
	if err := s.openStore(); err != nil {
		return err
	}
	if err := s.openCache(); err != nil {
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
	storeServer, err := s.GetStore()
	if err != nil {
		return errors.New("empty store")
	}

	if err := storeServer.Open(); err != nil {
		return err
	}
	return nil
}

func (s *Server) openCache() error {
	cacheServer, err := s.GetCache()
	if err != nil {
		return errors.New("empty cache")
	}

	if err := cacheServer.Open(); err != nil {
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
		api.POST("/auth2fa", s.handler2Factor())

		// temporarily for testing
		apiTest := api.Group("/test")
		apiTest.Use(middleware.UserIdentity())
		{
			apiTest.GET("/test_auth", s.handleTestAuth())
		}
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

func (s *Server) GetCache() (*cache.Cache, error) {
	if s.cache == nil {
		return nil, errors.New("empty cache")
	}
	return s.cache, nil
}
