package mail

import (
	"errors"
	"os"

	"github.com/pelletier/go-toml"
)

var (
	noKeyEnvironmentVariables = errors.New("no key in environment variables")
)

type Config struct {
	Mail         string
	Password     string
	Host         string
	PathTemplate string
	Port         int64
}

func newConfig() (*Config, error) {
	configPath, ok := os.LookupEnv("PATH_CONFIG")
	if !ok {
		return nil, noKeyEnvironmentVariables
	}

	config, err := toml.LoadFile(configPath)
	if err != nil {
		return nil, err
	}

	return &Config{
		Mail:         config.Get("mail.mail").(string),
		Password:     config.Get("mail.password").(string),
		Host:         config.Get("mail.host").(string),
		PathTemplate: config.Get("mail.path_template").(string),
		Port:         config.Get("mail.port").(int64),
	}, nil
}
