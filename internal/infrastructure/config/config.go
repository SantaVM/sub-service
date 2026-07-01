package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	BasePath    string

	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	Env string

	SwaggerHost    string
	RequestTimeout int
}

func Load() (*Config, error) {
	env := getEnv("GO_ENV", "local")

	if err := loadEnvFile(env); err != nil {
		return nil, err
	}

	timeoutStr := getEnv("REQUEST_TIMEOUT", "4")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REQUEST_TIMEOUT: %w", err)
	}

	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		BasePath:       getEnv("BASE_PATH", "/api/v1"),
		DBHost:         getEnv("DB_HOST", "localhost"),
		DBPort:         getEnv("DB_PORT", "5432"),
		DBName:         getEnv("DB_NAME", "subservice"),
		DBUser:         getEnv("DB_USER", "postgres"),
		DBPassword:     getEnv("DB_PASSWORD", "postgres"),
		SwaggerHost:    getEnv("SWAGGER_HOST", "localhost"),
		RequestTimeout: timeout,
	}

	if err := validatePort(cfg.Port); err != nil {
		return nil, err
	}

	if err := validatePort(cfg.DBPort); err != nil {
		return nil, err
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

func validatePort(port string) error {
	if port == "" {
		return fmt.Errorf("PORT is not set")
	}

	portValue, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid PORT: %w", err)
	}

	if portValue < 1 || portValue > 65535 {
		return fmt.Errorf("PORT must be between 1 and 65535")
	}

	return nil
}
