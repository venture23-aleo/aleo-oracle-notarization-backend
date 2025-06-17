package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/api"
	aleo "github.com/zkportal/aleo-utils-go"
)

const (
	IdleTimeout      = 30
	ReadWriteTimeout = 60
)

func StartServer(s aleo.Session) error {
	mux := http.NewServeMux()

	api.RegisterRoutes(mux,s)

	port := os.Getenv("PORT")

	if port == "" {
    	port = "8080" // default port
	}

	bindAddr := fmt.Sprintf("0.0.0.0:%s", port)

	server := &http.Server{
		IdleTimeout:       time.Second * IdleTimeout,
		ReadHeaderTimeout: time.Second * ReadWriteTimeout,
		WriteTimeout:      time.Second * ReadWriteTimeout,
		Addr:              bindAddr,
		Handler:           mux,
	}

	log.Printf("Server is running on %v",bindAddr)

	return server.ListenAndServe()
}
