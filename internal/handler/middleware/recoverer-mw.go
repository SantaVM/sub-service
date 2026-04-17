package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {

					// логируем с контекстом (request_id подтянется автоматически)
					log.ErrorContext(r.Context(), "panic recovered",
						"error", rec,
						"stacktrace", string(debug.Stack()),
					)

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
