package main

import (
	"log"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/app"
)

func main() {

	if err := app.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}

}
