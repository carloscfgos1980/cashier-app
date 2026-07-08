package config

import (
	"errors"
	"os"

	"github.com/carloscfgos1980/cashier-app/internal/database"

	"github.com/joho/godotenv"
)

var (
	ErrMissingDatabaseURL = errors.New("missing database URL")
	ErrMissingPort        = errors.New("missing port")
)

type Config struct {
	DB     *database.Queries
	DB_URL string
	Port   string
}

func LoadConfig() (*Config, error) {
	// Try common .env locations (project root and cmd/ execution path).
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load("../.env.local")

	DB_URL := os.Getenv("DB_URL")
	if DB_URL == "" {
		DB_URL = os.Getenv("DATABASE_URL")
	}
	if DB_URL == "" {
		return nil, ErrMissingDatabaseURL
	}

	Port := os.Getenv("PORT")
	if Port == "" {
		return nil, ErrMissingPort
	}

	// Return the configuration struct with the loaded values
	return &Config{
		DB_URL: DB_URL,
		Port:   Port,
	}, nil
}
