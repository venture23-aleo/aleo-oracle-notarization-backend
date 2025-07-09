package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/metrics"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/server"
	aleoContext "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/aleo_context"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
)

func main() {
	// 1. Initialize logger
	logger.InitLogger()

	// 2. Validate configuration at startup
	logger.Debug("Validating configuration...")
	if err := configs.ValidateConfigs(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// 3. Initialize Aleo context
	if err := aleoContext.InitAleoContext(); err != nil {
		log.Fatalf("Failed to initialize Aleo context: %v", err)
	}

	// 4. Start system metrics collector
	systemMetricsCollector := metrics.NewSystemMetricsCollector()
	systemMetricsCollector.Start()
	defer systemMetricsCollector.Stop()

	// 5. Create HTTP server
	srv := server.NewServer()

	// 6. Start server
	serverErr := make(chan error, 1)
	go func() {
		// Log the server is running on the bind address.
		logger.Info("Server started", "address", srv.Addr)
		serverErr <- srv.ListenAndServe()
	}()

	// 7. Listen for shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Error("Received shutdown signal")
	case err := <-serverErr:
		logger.Error("Server error", "error", err)
	}

	// 8. Graceful shutdown
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", "error", err)
	}

	// 9. Shutdown Aleo context
	if err := aleoContext.ShutdownAleoContext(); err != nil {
		logger.Error("Error shutting down Aleo context", "error", err)
	} else {
		logger.Info("Aleo context shutdown successfully")
	}

	logger.Debug("Server exited cleanly")
}
