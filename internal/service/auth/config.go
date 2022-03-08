package auth

import (
	"os"

	"github.com/pelletier/go-toml"
)

type Config struct {
	signingKey     string
	expireDuration int64
}

func NewConfig() (*Config, error) {
	configPath, _ := os.LookupEnv("PATH_CONFIG")
	config, err := toml.LoadFile(configPath)
	if err != nil {
		return nil, err
	}

	return &Config{
		signingKey:     config.Get("auth.signing_key").(string),
		expireDuration: config.Get("auth.expire_duration").(int64),
	}, nil
}
