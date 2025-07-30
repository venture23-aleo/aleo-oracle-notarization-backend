package middleware

import (
	"net/http"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/common"
	httpUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/httputil"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/metrics"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging middleware logs the incoming request and the response
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate a unique request ID
		requestID := common.GenerateShortRequestID()

		// Add request ID to response headers for client reference
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to request context
		ctx := logger.ContextWithRequestID(r.Context(), requestID)
		r = r.WithContext(ctx)

		// Get logger with request ID context
		reqLogger := logger.FromContext(ctx)

		clientIP := httpUtil.GetClientIP(r)
		userAgent := r.UserAgent()

		// Log the incoming request
		reqLogger.Debug("Incoming request",
			"method", r.Method,
			"path", r.URL.Path,
			"remote_addr", clientIP,
			"user_agent", userAgent,
		)

		// Create a custom response writer to capture status code
		responseWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(responseWriter, r)

		// Calculate duration once for both logging and metrics
		duration := time.Since(start)
		durationSeconds := duration.Seconds()

		// Record Prometheus metrics
		metrics.RecordHttpRequest(r.Method, r.URL.Path, responseWriter.statusCode, durationSeconds)

		// Log the request completion with duration and status
		reqLogger.Info("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status_code", responseWriter.statusCode,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", clientIP,
			"user_agent", userAgent,
		)
	})
}
