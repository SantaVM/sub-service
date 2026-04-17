package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Env string // local | prod
}

func New(cfg Config) *slog.Logger {
	env := strings.ToLower(cfg.Env)

	var handler slog.Handler

	switch env {
	case "prod":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return slog.New(handler)
}
