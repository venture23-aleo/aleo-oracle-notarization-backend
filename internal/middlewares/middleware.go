package middlewares

import (
	"net/http"
)

// Middleware represents a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain applies multiple middleware to a handler in order
func Chain(handler http.Handler, middleware ...Middleware) http.Handler {
	// Apply middleware in reverse order (last middleware wraps first)
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}

// ChainFunc applies multiple middleware to a handler function
func ChainFunc(handlerFunc http.HandlerFunc, middleware ...Middleware) http.Handler {
	return Chain(handlerFunc, middleware...)
}
