package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func LoggerMw(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			log.InfoContext(
				r.Context(),
				"http request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_ip", r.RemoteAddr,
				// "user_agent", r.UserAgent(),
			)

			next.ServeHTTP(ww, r)

			log.InfoContext(
				r.Context(),
				"http request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start).String(),
			)
		})
	}
}
