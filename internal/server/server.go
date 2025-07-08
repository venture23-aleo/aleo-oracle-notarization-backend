package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/api"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/middlewares"
)

const (
	IdleTimeout  = 30
	ReadTimeout  = 5
	WriteTimeout = 10
)

// InitializeServer initializes the server.
func NewServer() *http.Server {
	// Get app config
	appConfig := configs.GetAppConfig()
	// Create a new serve mux.
	mux := http.NewServeMux()

	// Register the routes.
	api.RegisterRoutes(mux)

	// Create middleware stack
	middlewareStack := []middlewares.Middleware{
		middlewares.LoggingMiddleware, // Log all requests with request ID
	}

	// Apply middleware stack to mux
	handler := middlewares.Chain(mux, middlewareStack...)

	port := appConfig.Port

	// Create the bind address.
	bindAddr := fmt.Sprintf(":%d", port)

	// Create the server.
	server := &http.Server{
		IdleTimeout:       time.Second * IdleTimeout,
		ReadHeaderTimeout: time.Second * ReadTimeout,
		WriteTimeout:      time.Second * WriteTimeout,
		Addr:              bindAddr,
		Handler:           handler, // Use the middleware stack
	}

	return server
}
