package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func mustRead(path string) []byte {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read %s: %v", path, err)
	}
	return b
}

func main() {
	// Default relative paths from repo root; override via env if needed.
	caPath := getenv("CA_CERT", "../../deployment/secrets/mtls/ca.crt")
	crtPath := getenv("CLIENT_CERT", "../../deployment/secrets/mtls/clients/AleoOracleClient/client.crt")
	keyPath := getenv("CLIENT_KEY", "../../deployment/secrets/mtls/clients/AleoOracleClient/client.key")
	url := getenv("SERVER_URL", "https://localhost:8443/health")

	// Normalize for helpful error messages
	caPath, _ = filepath.Abs(caPath)
	crtPath, _ = filepath.Abs(crtPath)
	keyPath, _ = filepath.Abs(keyPath)

	// Load CA
	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(mustRead(caPath)); !ok {
		log.Fatal("append CA failed")
	}

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(crtPath, keyPath)
	if err != nil {
		log.Fatalf("load client cert: %v", err)
	}

	// TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS12,
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	// Request
	resp, err := client.Get(url)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("status=%d\n%s\n", resp.StatusCode, string(body))
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
