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
)

func main() {
	// 1. Initialize logger
	appConfig, err := configs.GetAppConfigWithError()
	if err != nil {
		log.Fatalf("Failed to get app config: %v", err)
	}
	logger.InitLogger(appConfig.LogLevel)

	// 2. Validate configuration at startup
	logger.Debug("Validating configuration...")
	if err := configs.ValidateConfigs(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	// 3. Initialize Aleo context
	if err := aleoUtil.InitAleoContext(); err != nil {
		log.Fatalf("Failed to initialize Aleo context: %v", err)
	}

	// 4. Start system metrics collector
	systemMetricsCollector := metrics.NewSystemMetricsCollector()
	systemMetricsCollector.Start()
	defer systemMetricsCollector.Stop()

	// 5. Create HTTP server
	notarizationServer, metricsServer := server.NewServer()

	// Create a channel to listen for server errors
	serverErr := make(chan error, 2)

	// 6. Start notarization server
	go func() {
		logger.Info("Notarization server started", "address", notarizationServer.Addr)
		serverErr <- notarizationServer.ListenAndServe()
	}()

	// 7. Start metrics server
	go func() {
		logger.Info("Metrics server started", "address", metricsServer.Addr)
		serverErr <- metricsServer.ListenAndServe()
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

	// 9. Shutdown servers
	if err := notarizationServer.Shutdown(ctx); err != nil {
		logger.Error("Notarization server shutdown error", "error", err)
	}
	if err := metricsServer.Shutdown(ctx); err != nil {
		logger.Error("Metrics server shutdown error", "error", err)
	}

	// 9. Shutdown Aleo context
	if err := aleoUtil.ShutdownAleoContext(); err != nil {
		logger.Error("Error shutting down Aleo context", "error", err)
	} else {
		logger.Info("Aleo context shutdown successfully")
	}

	logger.Debug("Server exited cleanly")
}
