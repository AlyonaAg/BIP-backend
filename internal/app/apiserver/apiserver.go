package apiserver

import (
	"errors"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	_ "BIP_backend/docs"
	"BIP_backend/internal/app/cache/keycache"
	"BIP_backend/internal/app/cache/onetimepasscache"
	"BIP_backend/internal/app/store"
	"BIP_backend/middleware"
)

type Server struct {
	config    *Config
	router    *gin.Engine
	store     *store.Store
	passCache *onetimepasscache.Cache
	keyCache  *keycache.Cache
}

func NewServer() (*Server, error) {
	config, err := NewConfig()
	if err != nil {
		return nil, err
	}

	return &Server{
		config:    config,
		router:    gin.Default(),
		store:     store.NewStore(config.Store),
		passCache: onetimepasscache.NewCache(config.PassCache),
		keyCache:  keycache.NewCache(config.KeyCache),
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
	passCacheServer, err := s.GetPassCache()
	if err != nil {
		return errors.New("empty cache")
	}

	if err := passCacheServer.Open(); err != nil {
		return err
	}

	keyCacheServer, err := s.GetKeyCache()
	if err != nil {
		return errors.New("empty cache")
	}

	if err := keyCacheServer.Open(); err != nil {
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

		api.POST("/auth2fa", middleware.UserIdentityWithUnauthorizedToken(), s.handler2Factor())
		api.GET("/profile", middleware.UserIdentityWithAuthorizedToken(), s.handlerProfile())
		api.GET("/get-money", middleware.UserIdentityWithAuthorizedToken(), s.handlerGetMoney())

		apiOrdinaryUser := api.Group("/client").
			Use(middleware.UserIdentityWithAuthorizedToken(), middleware.OrdinaryUserIdentity())
		{
			apiOrdinaryUser.POST("/create-order", s.handlerCreateOrder())
			apiOrdinaryUser.POST("/finish-order", s.checkOrderForClient(), s.handlerFinishOrder())
			apiOrdinaryUser.POST("/review", s.checkOrderForClient(), s.handlerClientReview())
			apiOrdinaryUser.POST("/cancel", s.checkOrderForClient(), s.handlerCancel())
			apiOrdinaryUser.GET("/offer", s.checkOrderForClient(), s.handlerGetAgreedPhotographer())
			apiOrdinaryUser.GET("/photographers", s.handlerGetAllPhotographer())
			apiOrdinaryUser.GET("/all-orders", s.handlerGetClientOrders())
			apiOrdinaryUser.GET("/get-preview", s.checkOrderForClient(), s.handlerGetPreview())
			apiOrdinaryUser.GET("/get-original", s.checkOrderForClient(), s.handlerGetOriginal())
			apiOrdinaryUser.GET("/qrcode", s.checkOrderForClient(), s.handlerCreateQRCode())
			apiOrdinaryUser.PATCH("/accept", s.checkOrderForClient(), s.handlerAccept())
		}

		apiPhotographer := api.Group("/ph").
			Use(middleware.UserIdentityWithAuthorizedToken(), middleware.PhotographerIdentity())
		{
			apiPhotographer.POST("/upload", s.checkOrderForPhotographer(), s.handlerUpload())
			apiPhotographer.POST("/review", s.checkOrderForPhotographer(), s.handlerPhotographerReview())
			apiPhotographer.GET("/orders", s.handlerGetOrder())
			apiPhotographer.GET("/all-orders", s.handlerGetPhotographerOrders())
			apiPhotographer.PATCH("/select", s.handlerSelect())
			apiPhotographer.PATCH("/confirm-qrcode", s.handlerConfirmQRCode())
		}
	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
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

func (s *Server) GetPassCache() (*onetimepasscache.Cache, error) {
	if s.passCache == nil {
		return nil, errors.New("empty pass cache")
	}
	return s.passCache, nil
}

func (s *Server) GetKeyCache() (*keycache.Cache, error) {
	if s.keyCache == nil {
		return nil, errors.New("empty key cache")
	}
	return s.keyCache, nil
}
