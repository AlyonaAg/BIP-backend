package auth

import (
	"errors"
	"os"

	"github.com/pelletier/go-toml"
)

var (
	noKeyEnvironmentVariables = errors.New("no key in environment variables")
)

type Config struct {
	signingKey     string
	expireDuration int64
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
		signingKey:     config.Get("auth.signing_key").(string),
		expireDuration: config.Get("auth.expire_duration").(int64),
	}, nil
}
