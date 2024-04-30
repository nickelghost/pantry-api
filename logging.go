package main

import (
	"context"
	"log/slog"
	"os"
)

type CloudLoggingHandler struct{ handler slog.Handler }

func NewCloudLoggingHandler() *CloudLoggingHandler {
	return &CloudLoggingHandler{handler: slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.MessageKey:
				a.Key = "message"
			case slog.SourceKey:
				a.Key = "logging.googleapis.com/sourceLocation"
			case slog.LevelKey:
				a.Key = "severity"

				level, _ := a.Value.Any().(slog.Level)
				if level == slog.Level(12) { //nolint:gomnd
					a.Value = slog.StringValue("CRITICAL")
				}
			case "trace":
				a.Key = "logging.googleapis.com/trace"
			}

			return a
		},
	})}
}

func (h *CloudLoggingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *CloudLoggingHandler) Handle(ctx context.Context, rec slog.Record) error {
	return h.handler.Handle(ctx, rec) //nolint:wrapcheck
}

func (h *CloudLoggingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CloudLoggingHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *CloudLoggingHandler) WithGroup(name string) slog.Handler {
	return &CloudLoggingHandler{handler: h.handler.WithGroup(name)}
}
