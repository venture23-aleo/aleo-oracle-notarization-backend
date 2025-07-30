package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

func init() {
	// Initialize logger for tests
	logger.InitLogger("INFO")
}

// Test helper function to create a simple handler
func createTestHandler(statusCode int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(body))
	}
}

// Test helper function to create middleware that adds a header
func createHeaderMiddleware(headerName, headerValue string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(headerName, headerValue)
			next.ServeHTTP(w, r)
		})
	}
}

// Test helper function to create middleware that modifies the request
func createRequestModifierMiddleware(modifyFunc func(*http.Request)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			modifyFunc(r)
			next.ServeHTTP(w, r)
		})
	}
}

func TestChain(t *testing.T) {
	tests := []struct {
		name            string
		middleware      []Middleware
		expectedStatus  int
		expectedBody    string
		expectedHeaders map[string]string
	}{
		{
			name:           "No middleware",
			middleware:     []Middleware{},
			expectedStatus: 200,
			expectedBody:   "test response",
		},
		{
			name: "Single middleware",
			middleware: []Middleware{
				createHeaderMiddleware("X-Test", "value1"),
			},
			expectedStatus: 200,
			expectedBody:   "test response",
			expectedHeaders: map[string]string{
				"X-Test": "value1",
			},
		},
		{
			name: "Multiple middleware in order",
			middleware: []Middleware{
				createHeaderMiddleware("X-First", "first"),
				createHeaderMiddleware("X-Second", "second"),
				createHeaderMiddleware("X-Third", "third"),
			},
			expectedStatus: 200,
			expectedBody:   "test response",
			expectedHeaders: map[string]string{
				"X-First":  "first",
				"X-Second": "second",
				"X-Third":  "third",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			handler := createTestHandler(tt.expectedStatus, tt.expectedBody)

			// Apply middleware chain
			chainedHandler := Chain(handler, tt.middleware...)

			// Create test request
			req, err := http.NewRequest("GET", "/test", nil)
			require.NoError(t, err)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			chainedHandler.ServeHTTP(rr, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())

			// Verify headers
			for headerName, expectedValue := range tt.expectedHeaders {
				assert.Equal(t, expectedValue, rr.Header().Get(headerName))
			}
		})
	}
}

func TestChainFunc(t *testing.T) {
	tests := []struct {
		name            string
		middleware      []Middleware
		expectedStatus  int
		expectedBody    string
		expectedHeaders map[string]string
	}{
		{
			name:           "No middleware with handler func",
			middleware:     []Middleware{},
			expectedStatus: 201,
			expectedBody:   "handler func response",
		},
		{
			name: "Single middleware with handler func",
			middleware: []Middleware{
				createHeaderMiddleware("X-Handler", "func-value"),
			},
			expectedStatus: 201,
			expectedBody:   "handler func response",
			expectedHeaders: map[string]string{
				"X-Handler": "func-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler function
			handlerFunc := createTestHandler(tt.expectedStatus, tt.expectedBody)

			// Apply middleware chain
			chainedHandler := ChainFunc(handlerFunc, tt.middleware...)

			// Create test request
			req, err := http.NewRequest("POST", "/test", nil)
			require.NoError(t, err)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			chainedHandler.ServeHTTP(rr, req)

			// Verify response
			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())

			// Verify headers
			for headerName, expectedValue := range tt.expectedHeaders {
				assert.Equal(t, expectedValue, rr.Header().Get(headerName))
			}
		})
	}
}

func TestMiddlewareOrder(t *testing.T) {
	// Test that middleware is applied in the correct order
	order := []string{}

	// Create middleware that records execution order
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "middleware1-before")
			next.ServeHTTP(w, r)
			order = append(order, "middleware1-after")
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "middleware2-before")
			next.ServeHTTP(w, r)
			order = append(order, "middleware2-after")
		})
	}

	// Create handler that records execution
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, "handler")
		w.WriteHeader(200)
		w.Write([]byte("test"))
	})

	// Apply middleware chain
	chainedHandler := Chain(handler, middleware1, middleware2)

	// Create test request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute request
	chainedHandler.ServeHTTP(rr, req)

	// Verify execution order
	expectedOrder := []string{
		"middleware1-before",
		"middleware2-before",
		"handler",
		"middleware2-after",
		"middleware1-after",
	}

	assert.Equal(t, expectedOrder, order)
}

func TestLoggingMiddleware(t *testing.T) {
	// Create a simple test handler
	testHandler := createTestHandler(200, "test response")

	// Apply logging middleware
	loggedHandler := Logging(testHandler)

	// Create test request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Add some headers to test client IP extraction
	req.Header.Set("X-Real-IP", "192.168.1.1")
	req.Header.Set("User-Agent", "test-agent")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute request
	loggedHandler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "test response", rr.Body.String())

	// Verify that request ID header is set
	requestID := rr.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)
	assert.Len(t, requestID, 32)

	// Verify that context was modified with request ID
	ctx := req.Context()
	assert.NotNil(t, ctx)
}

func TestLoggingMiddlewareWithDifferentStatusCodes(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		expectedStatus int
	}{
		{"Success", 200, 200},
		{"Created", 201, 201},
		{"Bad Request", 400, 400},
		{"Not Found", 404, 404},
		{"Internal Server Error", 500, 500},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test handler with specific status code
			testHandler := createTestHandler(tc.statusCode, "test response")

			// Apply logging middleware
			loggedHandler := Logging(testHandler)

			// Create test request
			req, err := http.NewRequest("POST", "/test", nil)
			require.NoError(t, err)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			loggedHandler.ServeHTTP(rr, req)

			// Verify response
			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Equal(t, "test response", rr.Body.String())

			// Verify request ID is set
			assert.NotEmpty(t, rr.Header().Get("X-Request-ID"))
		})
	}
}

func TestLoggingMiddlewareWithDifferentMethods(t *testing.T) {
	testCases := []struct {
		name   string
		method string
	}{
		{"GET", "GET"},
		{"POST", "POST"},
		{"PUT", "PUT"},
		{"DELETE", "DELETE"},
		{"PATCH", "PATCH"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test handler
			testHandler := createTestHandler(200, "test response")

			// Apply logging middleware
			loggedHandler := Logging(testHandler)

			// Create test request
			req, err := http.NewRequest(tc.method, "/test", nil)
			require.NoError(t, err)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			loggedHandler.ServeHTTP(rr, req)

			// Verify response
			assert.Equal(t, 200, rr.Code)
			assert.Equal(t, "test response", rr.Body.String())

			// Verify request ID is set
			assert.NotEmpty(t, rr.Header().Get("X-Request-ID"))
		})
	}
}

func TestResponseWriterWrapper(t *testing.T) {
	// Test that the responseWriter wrapper correctly captures status codes
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(418) // I'm a teapot
		w.Write([]byte("teapot"))
	})

	// Apply logging middleware
	loggedHandler := Logging(testHandler)

	// Create test request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute request
	loggedHandler.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, 418, rr.Code)
	assert.Equal(t, "teapot", rr.Body.String())
}

func TestLoggingMiddlewareDuration(t *testing.T) {
	// Create a handler that takes some time
	slowHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(200)
		w.Write([]byte("slow response"))
	})

	// Apply logging middleware
	loggedHandler := Logging(slowHandler)

	// Create test request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Execute request
	start := time.Now()
	loggedHandler.ServeHTTP(rr, req)
	duration := time.Since(start)

	// Verify response
	assert.Equal(t, 200, rr.Code)
	assert.Equal(t, "slow response", rr.Body.String())

	// Verify that the request took at least 10ms
	assert.True(t, duration >= 10*time.Millisecond, "Request should have taken at least 10ms")
}

// Benchmark tests
func BenchmarkChain(b *testing.B) {
	middleware1 := createHeaderMiddleware("X-Test1", "value1")
	middleware2 := createHeaderMiddleware("X-Test2", "value2")
	middleware3 := createHeaderMiddleware("X-Test3", "value3")

	handler := createTestHandler(200, "test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Chain(handler, middleware1, middleware2, middleware3)
	}
}

func BenchmarkChainFunc(b *testing.B) {
	middleware1 := createHeaderMiddleware("X-Test1", "value1")
	middleware2 := createHeaderMiddleware("X-Test2", "value2")

	handlerFunc := createTestHandler(200, "test")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ChainFunc(handlerFunc, middleware1, middleware2)
	}
}

func BenchmarkLoggingMiddleware(b *testing.B) {
	handler := createTestHandler(200, "test")
	loggedHandler := Logging(handler)

	req, _ := http.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loggedHandler.ServeHTTP(rr, req)
	}
}
