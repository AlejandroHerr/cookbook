package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type RequestIDContextKey struct{}

func RequestIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			ww.Header().Set("X-Request-ID", requestID)

			ctx := r.Context()
			ctx = context.WithValue(ctx, RequestIDContextKey{}, requestID)

			next.ServeHTTP(ww, r.WithContext(ctx))
		})
	}
}

func RequestLoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status and size
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			requestID := r.Context().Value(RequestIDContextKey{}).(string)

			reqLogger := logger.With(
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_ip", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("referer", r.Referer()),
				slog.String("host", r.Host),
			)

			if r.URL.RawQuery != "" {
				reqLogger = logger.With(slog.Any("query", r.URL.Query()))
			}

			// Log the route pattern if available from Chi
			if routePattern := chi.RouteContext(r.Context()).RoutePattern(); routePattern != "" {
				reqLogger = logger.With(slog.String("route_pattern", routePattern))
			}

			// Debug level for request start
			reqLogger.Debug("request started")

			next.ServeHTTP(ww, r)

			// Log request completion with outcome information
			status := ww.Status()
			duration := float64(time.Since(start).Microseconds()) / 1000

			// Choose log level based on status code
			attrs := []any{
				"status", status,
				"bytes", ww.BytesWritten(),
				"duration_ms", duration,
			}

			// Add content type if present
			if ct := ww.Header().Get("Content-Type"); ct != "" {
				attrs = append(attrs, "content_type", ct)
			}

			switch {
			case status >= 500:
				reqLogger.Error("server error", attrs...)
			case status >= 400:
				reqLogger.Warn("client error", attrs...)
			case status >= 300:
				reqLogger.Info("redirect", attrs...)
			default:
				reqLogger.Info("request completed", attrs...)
			}
		})
	}
}
