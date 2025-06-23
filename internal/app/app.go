package app

import (
	"log"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/server"
	aleo "github.com/zkportal/aleo-utils-go"
)

func Run() error {
	// Create a new wrapper.
	wrapper, closeFn, err := aleo.NewWrapper()
	if err != nil {
		return err
	}

	// Close the wrapper.
	defer closeFn()

	// Create a new session.
	s, err := wrapper.NewSession()
	if err != nil {
		return err
	}

	// Generate a new private key.
	privKey, address, err := s.NewPrivateKey()
	if err != nil {
		return err
	}

	// Prefer passing config, not mutating shared state.
	configs.PrivateKey = privKey
	configs.PublicKey = address

	// Log the public key.
	log.Printf("Public key %v", address)

	// Start the server.
	return server.StartServer(s)
}
