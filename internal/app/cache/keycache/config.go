package keycache

import (
	"errors"
	"os"

	"github.com/pelletier/go-toml"
)

var (
	noKeyEnvironmentVariables = errors.New("no key in environment variables")
)

type Config struct {
	Port           string
	Password       string
	ExpireDuration int64
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

	return &Config{
		Port:           config.Get("key_cache.port").(string),
		Password:       config.Get("key_cache.password").(string),
		ExpireDuration: config.Get("key_cache.expire_duration").(int64),
	}, nil
}
