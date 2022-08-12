package apiserver

import (
	"BIP_backend/internal/app/cache/keycache"
	"BIP_backend/internal/app/cache/onetimepasscache"
	"errors"
	"os"

	"github.com/pelletier/go-toml"

	"BIP_backend/internal/app/store"
)

var (
	noKeyEnvironmentVariables = errors.New("no key in environment variables")
)

type Config struct {
	BindAddr  string
	Store     *store.Config
	PassCache *onetimepasscache.Config
	KeyCache  *keycache.Config
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

	passCacheConfig, err := onetimepasscache.NewConfig()
	if err != nil {
		return nil, err
	}

	keyCacheConfig, err := keycache.NewConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		BindAddr:  config.Get("server.bind_addr").(string),
		Store:     storeConfig,
		PassCache: passCacheConfig,
		KeyCache:  keyCacheConfig,
	}, nil
}
