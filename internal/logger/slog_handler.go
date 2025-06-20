package logger

import (
	"context"
	"log/slog"
)

type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if requestID, ok := ctx.Value("requestID").(string); ok {
		r.Add(slog.String("reqID", requestID))
	}
	return h.Handler.Handle(ctx, r)
}

func NewContextHandler(h slog.Handler) *ContextHandler {
	return &ContextHandler{h}
}
