package store

type Config struct {
	DatabaseURL   string `toml:"database_url"`
	PathMigration string `toml:"path_migration"`
}

func NewConfig() *Config {
	return &Config{
		DatabaseURL:   "",
		PathMigration: "",
	}
}
