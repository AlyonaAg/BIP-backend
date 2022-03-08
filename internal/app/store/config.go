package store

import (
	"os"

	"github.com/pelletier/go-toml"
)

type Config struct {
	DatabaseURL   string
	PathMigration string
}

func NewConfig() (*Config, error) {
	configPath, _ := os.LookupEnv("PATH_CONFIG")
	config, err := toml.LoadFile(configPath)
	if err != nil {
		return nil, err
	}

	return &Config{
		DatabaseURL:   config.Get("store.database_url").(string),
		PathMigration: config.Get("store.path_migration").(string),
	}, nil
}
