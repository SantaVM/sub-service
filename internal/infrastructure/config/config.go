package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	Env string
}

func Load() (*Config, error) {
	env := getEnv("GO_ENV", "local")

	if err := loadEnvFile(env); err != nil {
		return nil, err
	}

	cfg := &Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBName:     getEnv("DB_NAME", "subservice"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
	}

	// Формирование Database URL
	cfg.DatabaseURL = buildDSN(cfg)

	return cfg, nil
}

func loadEnvFile(env string) error {
	var file string

	switch env {
	case "local":
		file = ".env.local"
	case "docker":
		return nil // переменные уже переданы через docker-compose
	case "prod":
		return nil // в проде НЕ грузим файлы
	default:
		file = ".env"
	}

	if err := godotenv.Load(file); err != nil {
		return err
	}

	return nil
}

func buildDSN(cfg *Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
