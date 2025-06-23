package main

import (
	"log"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/app"
)

// main is the main function for the server.
func main() {

	// Run the application.
	if err := app.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}

}
