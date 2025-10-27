package middleware

import (
	"net/http"
)

// Middleware represents a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain applies multiple middleware to a handler in order
func Chain(handler http.Handler, middleware ...Middleware) http.Handler {
	// Defensive checks to avoid nil-function panics.
	if handler == nil {
		// Return a safe default handler instead of panicking.
		return http.NotFoundHandler()
	}
	// Apply middleware in reverse order (last middleware wraps first)
	for i := len(middleware) - 1; i >= 0; i-- {
		if middleware[i] == nil {
			// Skip nil middleware functions.
			continue
		}
		handler = middleware[i](handler)
	}
	return handler
}

// ChainFunc applies multiple middleware to a handler function
func ChainFunc(handlerFunc http.HandlerFunc, middleware ...Middleware) http.Handler {
	return Chain(handlerFunc, middleware...)
}
