package app

import (
	"log"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/server"
	aleo "github.com/zkportal/aleo-utils-go"
)

func Run() error {
	wrapper, closeFn, err := aleo.NewWrapper()
	if err != nil {
		return err
	}
	defer closeFn()

	s, err := wrapper.NewSession()
	if err != nil {
		return err
	}

	privKey, address, err := s.NewPrivateKey()
	if err != nil {
		return err
	}

	// Prefer passing config, not mutating shared state
	configs.PrivateKey = privKey
	configs.PublicKey = address

	log.Printf("Public key %v", address)

	return server.StartServer(s)
}
