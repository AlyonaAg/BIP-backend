package main

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"

	"BIP_backend/internal/app/apiserver"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	config := apiserver.NewConfig()
	configPath, ok := os.LookupEnv("PATH_CONFIG")
	if !ok {
		log.Fatal("Error env (missing PATH_CONFIG).")
	}

	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	s := apiserver.NewServer(config)
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
