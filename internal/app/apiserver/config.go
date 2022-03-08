package apiserver

import (
	"BIP_backend/internal/app/store"
	"github.com/pelletier/go-toml"
	"os"
)

type Config struct {
	BindAddr string
	Store    *store.Config
}

func NewConfig() (*Config, error) {
	configPath, _ := os.LookupEnv("PATH_CONFIG")
	config, err := toml.LoadFile(configPath)
	if err != nil {
		return nil, err
	}

	storeConfig, err := store.NewConfig()
	if err != nil {
		return nil, err
	}

	return &Config{
		BindAddr: config.Get("server.bind_addr").(string),
		Store:    storeConfig,
	}, nil
}
