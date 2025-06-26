package main

import (
	"log"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/server"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services"
)

// main is the main function for the server.
func main() {

	aleoCtx, err := services.NewAleoContext()

	if err != nil {
		log.Fatalf("Failed to create Aleo context: %v", err)
	}	

	defer aleoCtx.Close() // Ensure resources are cleaned up when done.
	
	// Run the application.
	if err := server.StartServer(aleoCtx); err != nil {
		log.Fatalf("Server error: %v", err)
	}

}
