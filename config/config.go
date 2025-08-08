package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string `json:"SERVER_PORT"`
}

func New() *Config {
	return setup()
}

func setup() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	config := &Config{
		ServerPort: os.Getenv("SERVER_PORT"),
	}

	if err := validate(config); err != nil {
		log.Fatalf("error to validate config: %v", err)
	}

	return config
}
