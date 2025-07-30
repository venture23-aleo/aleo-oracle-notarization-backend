package logger

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test helper function to capture log output
func captureLogOutput(t *testing.T, testFunc func()) string {
	// Create a buffer to capture output
	var buf bytes.Buffer

	// Create a new logger that writes to our buffer
	testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Save original logger
	originalLogger := Logger

	// Replace logger temporarily
	Logger = testLogger

	// Run the test function
	testFunc()

	// Restore original logger
	Logger = originalLogger

	return buf.String()
}

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		expected slog.Level
	}{
		{"Empty level", "", slog.LevelInfo},
		{"DEBUG level", "DEBUG", slog.LevelDebug},
		{"WARN level", "WARN", slog.LevelWarn},
		{"ERROR level", "ERROR", slog.LevelError},
		{"Unknown level", "UNKNOWN", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize logger
			InitLogger(tt.logLevel)

			// Verify logger is created
			assert.NotNil(t, Logger)

			// Test that logger can be used
			output := captureLogOutput(t, func() {
				Logger.Info("test message")
			})

			assert.Contains(t, output, "test message")
		})
	}
}

func TestGetFunctionName(t *testing.T) {
	// Test that getFunctionName returns a valid function name
	functionName := getFunctionName()

	// Should not be empty
	assert.NotEmpty(t, functionName)

	// Should not be "unknown" for this test
	assert.NotEqual(t, "unknown", functionName)

	// Should be a valid function name (not empty and not "unknown")
	assert.True(t, len(functionName) > 0 && functionName != "unknown")
}

func TestLoggingFunctions(t *testing.T) {
	// Initialize logger
	InitLogger("DEBUG")

	tests := []struct {
		name     string
		logFunc  func()
		expected string
	}{
		{
			name: "Info logging",
			logFunc: func() {
				Info("test info message", "key", "value")
			},
			expected: "test info message",
		},
		{
			name: "Error logging",
			logFunc: func() {
				Error("test error message", "key", "value")
			},
			expected: "test error message",
		},
		{
			name: "Warn logging",
			logFunc: func() {
				Warn("test warn message", "key", "value")
			},
			expected: "test warn message",
		},
		{
			name: "Debug logging",
			logFunc: func() {
				Debug("test debug message", "key", "value")
			},
			expected: "test debug message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureLogOutput(t, tt.logFunc)

			// Verify message is logged
			assert.Contains(t, output, tt.expected)

			// Verify function name is included
			assert.Contains(t, output, "function_name")

			// Verify key-value pairs are included
			assert.Contains(t, output, "key")
			assert.Contains(t, output, "value")
		})
	}
}

func TestLoggingWithNilLogger(t *testing.T) {
	// Set logger to nil
	originalLogger := Logger
	Logger = nil
	defer func() { Logger = originalLogger }()

	// These should not panic
	assert.NotPanics(t, func() {
		Info("test message")
		Error("test message")
		Warn("test message")
		Debug("test message")
	})
}

func TestWithContext(t *testing.T) {
	// Initialize logger
	InitLogger("INFO")

	// Test WithContext with valid logger
	loggerWithContext := WithContext("key1", "value1", "key2", "value2")
	assert.NotNil(t, loggerWithContext)

	// Test WithContext with nil logger
	originalLogger := Logger
	Logger = nil
	defer func() { Logger = originalLogger }()

	nilLoggerWithContext := WithContext("key", "value")
	assert.Nil(t, nilLoggerWithContext)
}

func TestWithRequestID(t *testing.T) {
	// Initialize logger
	InitLogger("INFO")

	// Test WithRequestID with valid logger
	requestID := "test-request-id"
	loggerWithRequestID := WithRequestID(requestID)
	assert.NotNil(t, loggerWithRequestID)

	// Test WithRequestID with nil logger
	originalLogger := Logger
	Logger = nil
	defer func() { Logger = originalLogger }()

	nilLoggerWithRequestID := WithRequestID(requestID)
	assert.Nil(t, nilLoggerWithRequestID)
}

func TestContextWithRequestID(t *testing.T) {
	// Create a context
	ctx := context.Background()
	requestID := "test-request-id"

	// Add request ID to context
	ctxWithRequestID := ContextWithRequestID(ctx, requestID)

	// Verify request ID is in context
	retrievedRequestID, ok := ctxWithRequestID.Value(RequestIDKey{}).(string)
	assert.True(t, ok)
	assert.Equal(t, requestID, retrievedRequestID)

	// Verify original context is not modified
	_, ok = ctx.Value(RequestIDKey{}).(string)
	assert.False(t, ok)
}

func TestFromContext(t *testing.T) {
	// Initialize logger
	InitLogger("INFO")

	tests := []struct {
		name      string
		ctx       context.Context
		expectNil bool
	}{
		{
			name:      "Context with request ID",
			ctx:       ContextWithRequestID(context.Background(), "test-id"),
			expectNil: false,
		},
		{
			name:      "Context without request ID",
			ctx:       context.Background(),
			expectNil: false, // Should return base logger
		},
		{
			name:      "Context with empty request ID",
			ctx:       ContextWithRequestID(context.Background(), ""),
			expectNil: false, // Should return base logger
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := FromContext(tt.ctx)

			if tt.expectNil {
				assert.Nil(t, logger)
			} else {
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestFromContextWithNilLogger(t *testing.T) {
	// Set logger to nil
	originalLogger := Logger
	Logger = nil
	defer func() { Logger = originalLogger }()

	// Test FromContext with nil logger
	ctx := ContextWithRequestID(context.Background(), "test-id")
	logger := FromContext(ctx)
	assert.Nil(t, logger)
}

func TestErrorWithContext(t *testing.T) {
	// Initialize logger
	InitLogger("INFO")

	// Create context with request ID
	ctx := ContextWithRequestID(context.Background(), "test-request-id")

	// Test ErrorWithContext
	output := captureLogOutput(t, func() {
		ErrorWithContext(ctx, "test error with context", "key", "value")
	})

	// Verify message is logged
	assert.Contains(t, output, "test error with context")
	assert.Contains(t, output, "function_name")
	assert.Contains(t, output, "key")
	assert.Contains(t, output, "value")
}

func TestLoggingWithRequestID(t *testing.T) {
	// Initialize logger
	InitLogger("INFO")

	requestID := "test-request-id"

	tests := []struct {
		name     string
		logFunc  func()
		expected string
	}{
		{
			name: "InfoWithRequestID",
			logFunc: func() {
				InfoWithRequestID(requestID, "test info with request ID", "key", "value")
			},
			expected: "test info with request ID",
		},
		{
			name: "ErrorWithRequestID",
			logFunc: func() {
				ErrorWithRequestID(requestID, "test error with request ID", "key", "value")
			},
			expected: "test error with request ID",
		},
		{
			name: "WarnWithRequestID",
			logFunc: func() {
				WarnWithRequestID(requestID, "test warn with request ID", "key", "value")
			},
			expected: "test warn with request ID",
		},
		{
			name: "DebugWithRequestID",
			logFunc: func() {
				DebugWithRequestID(requestID, "test debug with request ID", "key", "value")
			},
			expected: "test debug with request ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureLogOutput(t, tt.logFunc)

			// Verify message is logged
			assert.Contains(t, output, tt.expected)

			// Verify request ID is included
			assert.Contains(t, output, requestID)

			// Verify function name is included
			assert.Contains(t, output, "function_name")

			// Verify key-value pairs are included
			assert.Contains(t, output, "key")
			assert.Contains(t, output, "value")
		})
	}
}

func TestLoggingWithRequestIDNilLogger(t *testing.T) {
	// Set logger to nil
	originalLogger := Logger
	Logger = nil
	defer func() { Logger = originalLogger }()

	requestID := "test-request-id"

	// These should not panic
	assert.NotPanics(t, func() {
		InfoWithRequestID(requestID, "test message")
		ErrorWithRequestID(requestID, "test message")
		WarnWithRequestID(requestID, "test message")
		DebugWithRequestID(requestID, "test message")
	})
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name        string
		logLevel    string
		debugMsg    string
		infoMsg     string
		expectDebug bool
		expectInfo  bool
	}{
		{
			name:        "DEBUG level - should show all",
			logLevel:    "DEBUG",
			debugMsg:    "debug message",
			infoMsg:     "info message",
			expectDebug: true,
			expectInfo:  true,
		},
		{
			name:        "INFO level - should hide debug",
			logLevel:    "INFO",
			debugMsg:    "debug message",
			infoMsg:     "info message",
			expectDebug: false,
			expectInfo:  true,
		},
		{
			name:        "WARN level - should hide debug and info",
			logLevel:    "WARN",
			debugMsg:    "debug message",
			infoMsg:     "info message",
			expectDebug: false,
			expectInfo:  false,
		},
		{
			name:        "ERROR level - should hide debug, info, and warn",
			logLevel:    "ERROR",
			debugMsg:    "debug message",
			infoMsg:     "info message",
			expectDebug: false,
			expectInfo:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			var buf bytes.Buffer

			// Determine the expected log level
			var expectedLevel slog.Level
			switch tt.logLevel {
			case "DEBUG":
				expectedLevel = slog.LevelDebug
			case "INFO":
				expectedLevel = slog.LevelInfo
			case "WARN":
				expectedLevel = slog.LevelWarn
			case "ERROR":
				expectedLevel = slog.LevelError
			default:
				expectedLevel = slog.LevelInfo
			}

			// Create a new logger that writes to our buffer with the correct level
			testLogger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
				Level: expectedLevel,
			}))

			// Save original logger
			originalLogger := Logger

			// Replace logger temporarily
			Logger = testLogger

			// Test debug message
			Debug(tt.debugMsg)

			// Test info message
			Info(tt.infoMsg)

			// Restore original logger
			Logger = originalLogger

			output := buf.String()

			// Check debug message
			if tt.expectDebug {
				assert.Contains(t, output, tt.debugMsg)
			} else {
				assert.NotContains(t, output, tt.debugMsg)
			}

			// Check info message
			if tt.expectInfo {
				assert.Contains(t, output, tt.infoMsg)
			} else {
				assert.NotContains(t, output, tt.infoMsg)
			}
		})
	}
}

func TestRequestIDKeyType(t *testing.T) {
	// Test that RequestIDKey is a unique type
	key1 := RequestIDKey{}
	key2 := RequestIDKey{}

	// They should be the same type
	assert.Equal(t, key1, key2)

	// Test that it can be used as a context key
	ctx := context.Background()
	ctx = context.WithValue(ctx, key1, "test-value")

	value, ok := ctx.Value(key2).(string)
	assert.True(t, ok)
	assert.Equal(t, "test-value", value)
}

// Benchmark tests
func BenchmarkInitLogger(b *testing.B) {
	for i := 0; i < b.N; i++ {
		InitLogger("INFO")
	}
}

func BenchmarkInfo(b *testing.B) {
	InitLogger("INFO")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("benchmark message", "key", "value")
	}
}

func BenchmarkError(b *testing.B) {
	InitLogger("INFO")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Error("benchmark error", "key", "value")
	}
}

func BenchmarkWithRequestID(b *testing.B) {
	InitLogger("INFO")
	requestID := "benchmark-request-id"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WithRequestID(requestID)
	}
}

func BenchmarkContextWithRequestID(b *testing.B) {
	ctx := context.Background()
	requestID := "benchmark-request-id"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ContextWithRequestID(ctx, requestID)
	}
}

func BenchmarkFromContext(b *testing.B) {
	InitLogger("INFO")
	ctx := ContextWithRequestID(context.Background(), "benchmark-request-id")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FromContext(ctx)
	}
}
