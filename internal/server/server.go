package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/api"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/middlewares"
)

// InitializeServer initializes the server.
func NewServer() *http.Server {
	// Get app config
	appConfig := configs.GetAppConfig()
	
	// Initialize DDoS protection with config
	middlewares.InitializeDDoSProtection(&appConfig)
	middlewares.InitializeRateLimit(&appConfig)
	middlewares.SetupWhitelistedIPs()

	// Create a new serve mux.
	mux := http.NewServeMux()

	// Register the routes.
	api.RegisterRoutes(mux)

	// Create middleware stack
	middlewareStack := []middlewares.Middleware{
		middlewares.LoggingMiddleware,     // Log all requests
		middlewares.DDoSProtectionMiddleware, // DDoS protection
	}

	// Apply middleware stack to mux
	handler := middlewares.Chain(mux, middlewareStack...)

	serverConfig := appConfig.Server

	// Parse duration strings
	idleTimeout, _ := time.ParseDuration(serverConfig.IdleTimeout)
	readTimeout, _ := time.ParseDuration(serverConfig.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(serverConfig.WriteTimeout)

	// Create the bind address.
	bindAddr := fmt.Sprintf("%s:%d", serverConfig.Host, serverConfig.Port)

	// Create the server.
	server := &http.Server{
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readTimeout,
		WriteTimeout:      writeTimeout,
		Addr:              bindAddr,
		Handler:           handler, // Use the middleware stack
	}
	
	return server
}
