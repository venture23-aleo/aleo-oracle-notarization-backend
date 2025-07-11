package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"strings"
)

var Logger *slog.Logger

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

	Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	// Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}
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

// Convenience functions for easier access
func Info(msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"function_name", getFunctionName()}...)
		Logger.Info(msg, allArgs...)
	}
}

func Error(msg string, args ...any) {
	// Add function name to error logs
	if Logger != nil {
		allArgs := append(args, []any{"function_name", getFunctionName()}...)
		Logger.Error(msg, allArgs...)
	}
}

func ErrorWithContext(ctx context.Context, msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"function_name", getFunctionName()}...)
		FromContext(ctx).Error(msg, allArgs...)
	}
}

func Warn(msg string, args ...any) {
	if Logger != nil {
		// Add function name to warning logs
		allArgs := append(args, []any{"function_name", getFunctionName()}...)
		Logger.Warn(msg, allArgs...)
	}
}

func Debug(msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"function_name", getFunctionName()}...)
		Logger.Debug(msg, allArgs...)
	}
}

// WithContext returns a logger with additional context
func WithContext(args ...any) *slog.Logger {
	if Logger != nil {
		return Logger.With(args...)
	}
	return nil
}

// WithRequestID returns a logger with request ID context
func WithRequestID(requestID string) *slog.Logger {
	if Logger != nil {
		return Logger.With("request_id", requestID)
	}
	return nil
}

// FromContext creates a logger with request ID from context
func FromContext(ctx context.Context) *slog.Logger {
	if Logger == nil {
		return nil
	}

	if requestID, ok := ctx.Value(RequestIDKey{}).(string); ok && requestID != "" {
		return Logger.With("request_id", requestID)
	}

	return Logger
}

// ContextWithRequestID adds request ID to context
func ContextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey{}, requestID)
}

// InfoWithRequestID logs info with request ID
func InfoWithRequestID(requestID, msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"request_id", requestID, "function_name", getFunctionName()}...)
		Logger.Info(msg, allArgs...)
	}
}

// ErrorWithRequestID logs error with request ID and function name
func ErrorWithRequestID(requestID, msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"request_id", requestID, "function_name", getFunctionName()}...)
		Logger.Error(msg, allArgs...)
	}
}

// WarnWithRequestID logs warning with request ID and function name
func WarnWithRequestID(requestID, msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"request_id", requestID, "function_name", getFunctionName()}...)
		Logger.Warn(msg, allArgs...)
	}
}

// DebugWithRequestID logs debug with request ID
func DebugWithRequestID(requestID, msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"request_id", requestID, "function_name", getFunctionName()}...)
		Logger.Debug(msg, allArgs...)
	}
}
