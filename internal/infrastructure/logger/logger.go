package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Env string // local | prod | docker
}

func New(cfg Config) *slog.Logger {
	env := strings.ToLower(cfg.Env)

	var baseHandler slog.Handler

	switch env {
	case "prod":
		baseHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	case "docker":
		baseHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	default:
		baseHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	// оборачиваем нашим Context handler-ом
	handler := NewContextHandler(baseHandler)

	logger := slog.New(handler)

	slog.SetDefault(logger)

	return logger
}
