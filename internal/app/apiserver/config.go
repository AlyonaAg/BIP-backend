package apiserver

import (
	"errors"
	"os"

	"github.com/pelletier/go-toml"

	"BIP_backend/internal/app/cache"
	"BIP_backend/internal/app/store"
)

var (
	noKeyEnvironmentVariables = errors.New("no key in environment variables")
)

type Config struct {
	BindAddr string
	Store    *store.Config
	Cache    *cache.Config
}

func NewConfig() (*Config, error) {
	configPath, ok := os.LookupEnv("PATH_CONFIG")
	if !ok {
		return nil, noKeyEnvironmentVariables
	}

	config, err := toml.LoadFile(configPath)
	if err != nil {
		return nil, err
	}

	storeConfig, err := store.NewConfig()
	if err != nil {
		return nil, err
	}

	cacheConfig, err := cache.NewConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		BindAddr: config.Get("server.bind_addr").(string),
		Store:    storeConfig,
		Cache:    cacheConfig,
	}, nil
}
