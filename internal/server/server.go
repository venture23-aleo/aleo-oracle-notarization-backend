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
	WriteTimeout = 20
)

// NewServer initializes the notarization and metrics servers.
func NewServer() (*http.Server, *http.Server) {
	// Get app config
	appConfig := configs.GetAppConfig()
	// Create a new serve mux.
	mux := http.NewServeMux()

	// Create a new metrics mux.
	metricsMux := http.NewServeMux()

	// Register the notarization routes.
	api.RegisterRoutes(mux)

	// Register the metrics route.
	api.RegisterMetricsRoute(metricsMux)

	// Create middleware stack
	middlewareStack := []middlewares.Middleware{
		middlewares.LoggingAndMetricsMiddleware, // Log all requests with request ID
	}

	// Apply middleware stack to the mux.
	handler := middlewares.Chain(mux, middlewareStack...)

	// Get the port and metrics port.
	port := appConfig.Port
	metricsPort := appConfig.MetricsPort

	// Create the bind addresses.
	bindAddr := fmt.Sprintf(":%d", port)
	metricsBindAddr := fmt.Sprintf(":%d", metricsPort)

	// Create the notarization server.
	server := &http.Server{
		IdleTimeout:       time.Second * IdleTimeout,
		ReadHeaderTimeout: time.Second * ReadTimeout,
		WriteTimeout:      time.Second * WriteTimeout,
		Addr:              bindAddr,
		Handler:           handler, // Use the middleware stack
	}

	// Create the metrics server.
	metricsServer := &http.Server{
		IdleTimeout:       time.Second * IdleTimeout,
		ReadHeaderTimeout: time.Second * ReadTimeout,
		WriteTimeout:      time.Second * WriteTimeout,
		Addr:              metricsBindAddr,
		Handler:           metricsMux, // Use the middleware stack
	}

	return server, metricsServer
}
