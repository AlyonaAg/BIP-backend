package cache

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
		Port:           config.Get("cache.port").(string),
		Password:       config.Get("cache.password").(string),
		ExpireDuration: config.Get("cache.expire_duration").(int64),
	}, nil
}
