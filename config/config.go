package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load env: %w", err)
	}

	databaseEnvs := []string{
		"MYSQL_USER",
		"MYSQL_PASSWORD",
		"MYSQL_HOST",
		"MYSQL_PORT",
		"MYSQL_DATABASE",
	}

	for _, env := range databaseEnvs {
		if os.Getenv(env) == "" {
			return nil, fmt.Errorf("database env is not set")
		}
	}

	return &Config{
		DBUser:     os.Getenv("MYSQL_USER"),
		DBPassword: os.Getenv("MYSQL_PASSWORD"),
		DBHost:     os.Getenv("MYSQL_HOST"),
		DBPort:     os.Getenv("MYSQL_PORT"),
		DBName:     os.Getenv("MYSQL_DATABASE"),
	}, nil
}
