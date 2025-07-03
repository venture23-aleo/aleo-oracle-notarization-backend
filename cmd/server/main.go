package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/server"
	aleoContext "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/aleo_context"
)

func main() {
	// Validate configuration at startup
	log.Println("Validating configuration...")
	if err := configs.ValidateConfigs(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}
	log.Println("✅ Configuration validation passed")

	// Test that configs are loaded correctly
	exchangeConfigs := configs.GetExchangeConfigs()
	symbolExchanges := configs.GetSymbolExchanges()
	
	log.Printf("📊 Loaded %d exchange configurations", len(exchangeConfigs))
	log.Printf("📈 Loaded %d symbol mappings", len(symbolExchanges))
	
	// Log available exchanges and symbols
	log.Println("🏢 Available exchanges:")
	for key, config := range exchangeConfigs {
		log.Printf("  - %s: %s (%s)", key, config.Name, config.BaseURL)
	}
	
	log.Println("💱 Available symbols:")
	for symbol, exchanges := range symbolExchanges {
		log.Printf("  - %s: %v", symbol, exchanges)
	}

	// 1. Initialize Aleo context
	if err := aleoContext.InitAleoContext(); err != nil {
		log.Fatalf("Failed to initialize Aleo context: %v", err)
	}

	// 2. Create server
	srv := server.NewServer()

	// 3. Start server
	serverErr := make(chan error, 1)
	go func() {
		// Log the server is running on the bind address.
		log.Printf("🚀 Server is listening on %v", srv.Addr)
		serverErr <- srv.ListenAndServe()
	}()

	// 4. Listen for shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
		log.Println("Received shutdown signal")
	case err := <-serverErr:
		log.Printf("Server error: %v", err)
	}

	// 5. Graceful shutdown
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	// 6. Shutdown Aleo context
	if err := aleoContext.ShutdownAleoContext(); err != nil {
		log.Printf("Error shutting down Aleo context: %v", err)
	} else {
		log.Println("Aleo context shutdown successfully")
	}

	log.Println("Server exited cleanly")
}