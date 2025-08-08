package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort               string `json:"SERVER_PORT"`
	PaymentProcessorDefault  string `json:"PAYMENT_PROCESSOR_DEFAULT"`
	PaymentProcessorFallback string `json:"PAYMENT_PROCESSOR_FALLBACK"`
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
		ServerPort:               os.Getenv("SERVER_PORT"),
		PaymentProcessorDefault:  os.Getenv("PAYMENT_PROCESSOR_DEFAULT"),
		PaymentProcessorFallback: os.Getenv("PAYMENT_PROCESSOR_FALLBACK"),
	}

	if err := validate(config); err != nil {
		log.Fatalf("error validating config: %v", err)
	}

	return config
}
