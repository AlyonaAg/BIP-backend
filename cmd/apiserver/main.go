package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"BIP_backend/internal/app/apiserver"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
}

// @title           BIP API
// @version         1.0
// @description     API for photographer search app

// @host       192.168.1.71:8080
// @BasePath  /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	_, ok := os.LookupEnv("PATH_CONFIG")
	if !ok {
		log.Fatal("Error env (missing PATH_CONFIG).")
	}

	s, err := apiserver.NewServer()
	if err != nil {
		log.Fatal(err)
	}

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
