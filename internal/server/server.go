package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/api"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services"
)

// IdleTimeout and ReadWriteTimeout are the timeout for the server.
const (
	IdleTimeout      = 30
	ReadWriteTimeout = 60
)

// StartServer starts the server.
func StartServer(aleoCtx services.AleoPublicContext) error {

	// Create a new serve mux.
	mux := http.NewServeMux()

	// Register the routes.
	api.RegisterRoutes(mux, aleoCtx)

	// Get the port from the environment variable.
	port := os.Getenv("PORT")

	// If the port is not set, use the default port.
	if port == "" {
		port = "8080" // default port
	}

	// Create the bind address.
	bindAddr := fmt.Sprintf("0.0.0.0:%s", port)

	// Create the server.
	server := &http.Server{
		IdleTimeout:       time.Second * IdleTimeout,
		ReadHeaderTimeout: time.Second * ReadWriteTimeout,
		WriteTimeout:      time.Second * ReadWriteTimeout,
		Addr:              bindAddr,
		Handler:           mux,
	}

	// Log the server is running on the bind address.
	log.Printf("Server is running on %v", bindAddr)

	// Listen and serve the server.
	return server.ListenAndServe()
}
