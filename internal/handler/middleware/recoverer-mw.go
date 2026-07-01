package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
)

func Recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {

					stack := debug.Stack()

					// логируем с контекстом (request_id подтянется автоматически)
					log.ErrorContext(r.Context(), "panic recovered",
						"error", rec,
						"stacktrace", string(stack),
					)

					// для отладки выводим стек в stderr для удобочитаемости
					fmt.Fprintf(os.Stderr, "stacktrace:\n%s\n", stack)

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
