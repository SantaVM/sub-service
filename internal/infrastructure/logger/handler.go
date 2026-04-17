package logger

import (
	"context"
	"log/slog"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type ContextHandler struct {
	handler slog.Handler
}

func NewContextHandler(h slog.Handler) *ContextHandler {
	return &ContextHandler{handler: h}
}

func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// достаём request_id из context
	if reqID := chimiddleware.GetReqID(ctx); reqID != "" {
		r.AddAttrs(slog.String("request_id", reqID))
	}

	return h.handler.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{handler: h.handler.WithGroup(name)}
}
