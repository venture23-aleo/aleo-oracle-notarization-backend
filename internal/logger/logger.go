package logger

import (
	"context"
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

func Fatal(msg string, args ...any) {
	if Logger != nil {
		allArgs := append(args, []any{"function_name", getFunctionName()}...)
		Logger.Error(msg, allArgs...)
		os.Exit(1)
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
