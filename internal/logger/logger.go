// Package logger provides structured, context-aware logging helpers used across the service.
// Package logger provides a thin wrapper around slog with context-aware helpers
// and safe defaults that avoid panics when the global logger is uninitialized.
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

var Logger *slog.Logger

// defaultNoopLogger is used when the global Logger is not initialized yet.
// It discards all logs to avoid panics while keeping call sites simple.
var defaultNoopLogger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))

// baseLogger returns the initialized global logger or a safe no-op logger.
func baseLogger() *slog.Logger {
	if Logger != nil {
		return Logger
	}
	return defaultNoopLogger
}

func InitLogger(logLevel string) {
	level := slog.LevelInfo
	if logLevel != "" {
		switch logLevel {
		case "DEBUG":
			level = slog.LevelDebug
		case "WARN":
			level = slog.LevelWarn
		case "ERROR":
			level = slog.LevelError
		}
	}
	Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	// Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

// RequestIDKey is the context key for request ID
type RequestIDKey struct{}

// getFunctionName gets the calling function name
func getFunctionName() string {
	pc, _, _, ok := runtime.Caller(2) // Skip 2 frames: getFunctionName -> Error/Warn -> actual caller
	if !ok {
		return "unknown"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}

	// Extract just the function name from the full path
	fullName := fn.Name()
	parts := strings.Split(fullName, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1] // Get the last part (function name)
	}
	return fullName
}

// Info logs an informational message, enriching with the calling function name.
// Info logs a message with optional key/value pairs.
func Info(msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"function_name", getFunctionName()}...)
		Logger.Info(msg, allArgs...)
	}
}

// Error logs an error message, enriching with the calling function name.
func Error(msg string, args ...any) {
	// Add function name to error logs
	allArgs := append(args, []any{"function_name", getFunctionName()}...)
	baseLogger().Error(msg, allArgs...)
}

// ErrorWithContext logs an error using a logger derived from the provided context.
func ErrorWithContext(ctx context.Context, msg string, args ...any) {
	allArgs := append(args, []any{"function_name", getFunctionName()}...)
	FromContext(ctx).Error(msg, allArgs...)
}

// Warn logs a warning message, enriching with the calling function name.
func Warn(msg string, args ...any) {
	// Add function name to warning logs
	allArgs := append(args, []any{"function_name", getFunctionName()}...)
	baseLogger().Warn(msg, allArgs...)
}

// Debug logs a debug message, enriching with the calling function name.
func Debug(msg string, args ...any) {
	allArgs := append(args, []any{"function_name", getFunctionName()}...)
	baseLogger().Debug(msg, allArgs...)
}

// WithContext returns a logger with additional structured fields.
func WithContext(args ...any) *slog.Logger {
	return baseLogger().With(args...)
}

// WithRequestID returns a logger annotated with the provided request ID.
func WithRequestID(requestID string) *slog.Logger {
	return baseLogger().With("request_id", requestID)
}

// FromContext returns a logger derived from the context, annotating with request_id when present.
func FromContext(ctx context.Context) *slog.Logger {
	// Guard against nil context and uninitialized logger.
	if ctx == nil {
		return baseLogger()
	}
	if requestID, ok := ctx.Value(RequestIDKey{}).(string); ok && requestID != "" {
		return baseLogger().With("request_id", requestID)
	}
	return baseLogger()
}

// ContextWithRequestID returns a new context containing the provided request ID.
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey{}, requestID)
}

// InfoWithRequestID logs an informational message annotated with the provided request ID.
func InfoWithRequestID(requestID, msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"request_id", requestID, "function_name", getFunctionName()}...)
		Logger.Info(msg, allArgs...)
	}
}

// ErrorWithRequestID logs an error annotated with the provided request ID.
func ErrorWithRequestID(requestID, msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"request_id", requestID, "function_name", getFunctionName()}...)
		Logger.Error(msg, allArgs...)
	}
}

// WarnWithRequestID logs a warning annotated with the provided request ID.
func WarnWithRequestID(requestID, msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"request_id", requestID, "function_name", getFunctionName()}...)
		Logger.Warn(msg, allArgs...)
	}
}

// DebugWithRequestID logs a debug message annotated with the provided request ID.
func DebugWithRequestID(requestID, msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"request_id", requestID, "function_name", getFunctionName()}...)
		Logger.Debug(msg, allArgs...)
	}
}
