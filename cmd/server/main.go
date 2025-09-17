package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	aleoUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/aleoutil"
	configs "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/config"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/metrics"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/server"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/sgx"
)

func main() {

	// 1. Initialize logger
	appConfig, err := configs.GetAppConfigWithError()
	if err != nil {
		log.Fatalf("Failed to get app config: %v", err)
	}
	logger.InitLogger(appConfig.LogLevel)

	// 2. Enforce SGX startup check
	if err := sgx.EnforceSGXStartup(); err != nil {
		log.Fatalf("Failed to enforce SGX startup check: %v", err)
	}

	// 3. Validate configuration at startup
	logger.Debug("Validating configuration...")
	if err := configs.ValidateConfigs(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// 4. Initialize Aleo context
	if err := aleoUtil.InitAleoContext(); err != nil {
		log.Fatalf("Failed to initialize Aleo context: %v", err)
	}

	// 5. Start system metrics collector
	systemMetricsCollector := metrics.NewSystemMetricsCollector()
	systemMetricsCollector.Start()
	defer systemMetricsCollector.Stop()

	// 6. Create HTTP server
	notarizationServer, metricsServer := server.NewServer()

	// Create a channel to listen for server errors
	serverErr := make(chan error, 2)

	// 7. Start notarization server
	go func() {
		logger.Info("Notarization server started", "address", notarizationServer.Addr)
		serverErr <- notarizationServer.ListenAndServe()
	}()

	// 8. Start metrics server
	go func() {
		logger.Info("Metrics server started", "address", metricsServer.Addr)
		serverErr <- metricsServer.ListenAndServe()
	}()

	// 9. Listen for shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Error("Received shutdown signal")
	case err := <-serverErr:
		logger.Error("Server error", "error", err)
	}

	// 10. Graceful shutdown
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 11. Shutdown servers
	if err := notarizationServer.Shutdown(ctx); err != nil {
		logger.Error("Notarization server shutdown error", "error", err)
	}
	if err := metricsServer.Shutdown(ctx); err != nil {
		logger.Error("Metrics server shutdown error", "error", err)
	}

	// 12. Shutdown Aleo context
	if err := aleoUtil.ShutdownAleoContext(); err != nil {
		logger.Error("Error shutting down Aleo context", "error", err)
	} else {
		logger.Info("Aleo context shutdown successfully")
	}

	logger.Debug("Server exited cleanly")
}
