package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/AlejandroHerr/cookbook/internal/common/api"
	"github.com/golang-cz/devslog"
)

type Config struct {
	Environment string // "development", "production", "staging"
	Level       string // "debug", "info", "warn", "error"
	App         string
	Version     string
}

func New(cfg Config) *slog.Logger {
	level := levelFromString(cfg.Level, slog.LevelDebug)
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.Environment == "development",
	}

	var handler slog.Handler

	switch strings.ToLower(cfg.Environment) {
	case "development":
		// Use pretty text handler for development
		handler = devslog.NewHandler(os.Stdout, &devslog.Options{
			HandlerOptions: opts,
		})
	default:
		// Use JSON for production/staging
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	handler = NewContextHandler(handler, map[any]string{
		api.RequestIDContextKey{}: "request_id",
	})

	// Create base logger with common attributes
	logger := slog.New(handler).With(
		"app", cfg.App,
		"environment", cfg.Environment,
		"version", cfg.Version,
	)

	return logger
}

func levelFromString(level string, defaultLevel slog.Level) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return defaultLevel
	}
}

// NewTestLogger returns a no-op logger that discards all output.
// Perfect for unit tests where you don't want to see log output.
func NewTestLogger() *slog.Logger {
	return slog.New(NewDiscardHandler())
}

// DiscardHandler implements slog.Handler but discards all log records.
type DiscardHandler struct{}

// NewDiscardHandler creates a new handler that discards all logs.
func NewDiscardHandler() *DiscardHandler {
	return &DiscardHandler{}
}

// Enabled always returns false to minimize overhead.
func (h *DiscardHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

// Handle discards the record and does nothing.
func (h *DiscardHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

// WithAttrs returns a new handler with the same behavior.
func (h *DiscardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

// WithGroup returns a new handler with the same behavior.
func (h *DiscardHandler) WithGroup(_ string) slog.Handler {
	return h
}

type ContextHandler struct {
	handler slog.Handler
	keys    map[any]string // Maps context keys to attribute names
}

func NewContextHandler(handler slog.Handler, keys map[any]string) *ContextHandler {
	return &ContextHandler{
		handler: handler,
		keys:    keys,
	}
}

func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	for ctxKey, attrName := range h.keys {
		if value := ctx.Value(ctxKey); value != nil {
			switch v := value.(type) {
			case string:
				if v != "" {
					r.AddAttrs(slog.String(attrName, v))
				}
			case int:
				r.AddAttrs(slog.Int(attrName, v))
			case int64:
				r.AddAttrs(slog.Int64(attrName, v))
			case float64:
				r.AddAttrs(slog.Float64(attrName, v))
			case bool:
				r.AddAttrs(slog.Bool(attrName, v))
			default:
				// For other types, convert to string
				r.AddAttrs(slog.Any(attrName, v))
			}
		}
	}

	return h.handler.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{
		handler: h.handler.WithAttrs(attrs),
		keys:    h.keys,
	}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{
		handler: h.handler.WithGroup(name),
		keys:    h.keys,
	}
}
