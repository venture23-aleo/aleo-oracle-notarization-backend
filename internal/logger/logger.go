// Package logger provides structured, context-aware logging helpers used across the service.
// It is a thin wrapper around slog with context-aware helpers and safe defaults
// that avoid panics when the global logger is uninitialized.
package logger

import (
	"context"
	"io"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
    Logger   *slog.Logger
    seqCount uint64
    loggerLock       sync.Mutex
)

type LogHandler struct {
	logLevel slog.Level
	out      *os.File
	attrs    []slog.Attr
	groups   []string
}

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

	logHandler := LogHandler{logLevel: level, out: os.Stdout}

	Logger = slog.New(&logHandler)
	// Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}


func (h *LogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.logLevel
}

func (h *LogHandler) Handle(_ context.Context, r slog.Record) error {
	groupName := strings.Join(h.groups, ".")

	loggerLock.Lock()
	defer loggerLock.Unlock()	

	// Update state
	seqCount++

	var parts []string

    parts = append(parts,
        "time="+r.Time.Format(time.RFC3339),
        "level="+r.Level.String(),
        "seq="+fmt.Sprint(seqCount),
        "msg="+fmt.Sprintf("%q", r.Message),
    )


	for _, attr := range h.attrs {
		if groupName != "" {
			parts = append(parts, fmt.Sprintf("%s.%s=%v", groupName, attr.Key, attr.Value.Any()))
		} else {
			parts = append(parts, fmt.Sprintf("%s=%v", attr.Key, attr.Value.Any()))
		}
	}

	r.Attrs(func(a slog.Attr) bool {
		if groupName != "" {
			parts = append(parts, fmt.Sprintf("%s.%s=%v", groupName, a.Key, a.Value.Any()))
		} else {
			parts = append(parts, fmt.Sprintf("%s=%v", a.Key, a.Value.Any()))
		}
		return true
	})

	fmt.Fprintln(h.out, strings.Join(parts, " "))

	return nil
}

func (h *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := *h
	newHandler.attrs = append(newHandler.attrs, attrs...)
	return &newHandler
}

func (h *LogHandler) WithGroup(name string) slog.Handler {
	newHandler := *h
	newHandler.groups = append(newHandler.groups, name)
	return &newHandler
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
// It accepts a message and optional key/value pairs for structured logging.
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

func Fatal(msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"function_name", getFunctionName()}...)
		Logger.Error(msg, allArgs...)
		os.Exit(1)
	}
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
