package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
	JWTSecret  string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	envs := []string{
		"MYSQL_USER",
		"MYSQL_PASSWORD",
		"MYSQL_HOST",
		"MYSQL_PORT",
		"MYSQL_DATABASE",
	}

	for _, env := range envs {
		if os.Getenv(env) == "" {
			return nil, err
		}
	}

	return &Config{
		DBUser:     os.Getenv("MYSQL_USER"),
		DBPassword: os.Getenv("MYSQL_PASSWORD"),
		DBHost:     os.Getenv("MYSQL_HOST"),
		DBPort:     os.Getenv("MYSQL_PORT"),
		DBName:     os.Getenv("MYSQL_DATABASE"),
		JWTSecret:  os.Getenv("JWT_SECRET"),
	}, nil
}
