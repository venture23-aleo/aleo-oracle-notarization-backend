package main

import (
	"fmt"
	"log"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/app"
)

// These will be injected via -ldflags at build time
var (
	Version   = "dev"
	Commit    = "none"
)

func main() {

	fmt.Printf("Version: %s\nCommit: %s\n", Version, Commit)

	if err := app.Run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}

}
