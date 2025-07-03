package middlewares

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs request details including method, path, duration, and status code
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a custom response writer to capture status code
		responseWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Call the next handler
		next.ServeHTTP(responseWriter, r)
		
		// Log the request details
		duration := time.Since(start)
		log.Printf(
			"[%s] %s %s - %d - %v - %s",
			r.Method,
			r.URL.Path,
			getClientIP(r),
			responseWriter.statusCode,
			duration,
			r.UserAgent(),
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
} 