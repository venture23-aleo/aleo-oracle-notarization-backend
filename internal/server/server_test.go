package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	configs "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/config"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

func init() {
	// Initialize logger for tests
	logger.InitLogger("INFO")
}

func TestNewServer(t *testing.T) {
	// Test server creation
	server, metricsServer := NewServer()

	// Verify servers are not nil
	assert.NotNil(t, server)
	assert.NotNil(t, metricsServer)

	// Verify server configurations
	assert.Equal(t, time.Duration(IdleTimeout)*time.Second, server.IdleTimeout)
	assert.Equal(t, time.Duration(ReadTimeout)*time.Second, server.ReadHeaderTimeout)
	assert.Equal(t, time.Duration(WriteTimeout)*time.Second, server.WriteTimeout)

	// Verify metrics server configurations
	assert.Equal(t, time.Duration(IdleTimeout)*time.Second, metricsServer.IdleTimeout)
	assert.Equal(t, time.Duration(ReadTimeout)*time.Second, metricsServer.ReadHeaderTimeout)
	assert.Equal(t, time.Duration(WriteTimeout)*time.Second, metricsServer.WriteTimeout)

	// Verify handlers are set
	assert.NotNil(t, server.Handler)
	assert.NotNil(t, metricsServer.Handler)

	// Verify addresses are properly formatted
	appConfig := configs.GetAppConfig()
	expectedAddr := fmt.Sprintf(":%d", appConfig.Port)
	expectedMetricsAddr := fmt.Sprintf(":%d", appConfig.MetricsPort)

	assert.Equal(t, expectedAddr, server.Addr)
	assert.Equal(t, expectedMetricsAddr, metricsServer.Addr)
}

func TestServerConstants(t *testing.T) {
	// Test that constants are properly defined
	assert.Equal(t, 30, IdleTimeout)
	assert.Equal(t, 5, ReadTimeout)
	assert.Equal(t, 20, WriteTimeout)

	// Test that timeouts are reasonable
	assert.True(t, IdleTimeout > 0)
	assert.True(t, ReadTimeout > 0)
	assert.True(t, WriteTimeout > 0)
	assert.True(t, IdleTimeout > ReadTimeout)
	assert.True(t, IdleTimeout > WriteTimeout)
}

func TestServerHandlerIntegration(t *testing.T) {
	// Create servers
	server, metricsServer := NewServer()

	// Test that handlers are properly configured
	require.NotNil(t, server.Handler)
	require.NotNil(t, metricsServer.Handler)

	// Test that we can create a test request
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Test that the handler can process requests
	server.Handler.ServeHTTP(rr, req)

	// Verify that the request was processed (status code should be set)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestMetricsServerHandlerIntegration(t *testing.T) {
	// Create servers
	_, metricsServer := NewServer()

	// Test that metrics handler is properly configured
	require.NotNil(t, metricsServer.Handler)

	// Test that we can create a test request
	req, err := http.NewRequest("GET", "/metrics", nil)
	require.NoError(t, err)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Test that the metrics handler can process requests
	metricsServer.Handler.ServeHTTP(rr, req)

	// Verify that the request was processed (status code should be set)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestServerConfiguration(t *testing.T) {
	// Test server creation multiple times to ensure consistency
	server1, metricsServer1 := NewServer()
	server2, metricsServer2 := NewServer()

	// Verify that servers are created consistently
	assert.Equal(t, server1.IdleTimeout, server2.IdleTimeout)
	assert.Equal(t, server1.ReadHeaderTimeout, server2.ReadHeaderTimeout)
	assert.Equal(t, server1.WriteTimeout, server2.WriteTimeout)
	assert.Equal(t, server1.Addr, server2.Addr)

	assert.Equal(t, metricsServer1.IdleTimeout, metricsServer2.IdleTimeout)
	assert.Equal(t, metricsServer1.ReadHeaderTimeout, metricsServer2.ReadHeaderTimeout)
	assert.Equal(t, metricsServer1.WriteTimeout, metricsServer2.WriteTimeout)
	assert.Equal(t, metricsServer1.Addr, metricsServer2.Addr)
}

func TestServerAddressFormatting(t *testing.T) {
	// Create servers
	server, metricsServer := NewServer()

	// Test that addresses are properly formatted with colons
	assert.Contains(t, server.Addr, ":")
	assert.Contains(t, metricsServer.Addr, ":")

	// Test that addresses are not empty
	assert.NotEmpty(t, server.Addr)
	assert.NotEmpty(t, metricsServer.Addr)

	// Test that addresses don't start with colon (should be ":port")
	assert.True(t, len(server.Addr) > 1)
	assert.True(t, len(metricsServer.Addr) > 1)
}

func TestServerTimeoutConfiguration(t *testing.T) {
	// Create server
	server, metricsServer := NewServer()

	// Test timeout relationships
	assert.True(t, server.IdleTimeout > server.ReadHeaderTimeout,
		"IdleTimeout should be greater than ReadHeaderTimeout")
	assert.True(t, server.IdleTimeout > server.WriteTimeout,
		"IdleTimeout should be greater than WriteTimeout")

	assert.True(t, metricsServer.IdleTimeout > metricsServer.ReadHeaderTimeout,
		"Metrics IdleTimeout should be greater than ReadHeaderTimeout")
	assert.True(t, metricsServer.IdleTimeout > metricsServer.WriteTimeout,
		"Metrics IdleTimeout should be greater than WriteTimeout")

	// Test that timeouts are reasonable values
	assert.True(t, server.IdleTimeout >= 30*time.Second)
	assert.True(t, server.ReadHeaderTimeout >= 5*time.Second)
	assert.True(t, server.WriteTimeout >= 20*time.Second)
}

func TestServerHandlerTypes(t *testing.T) {
	// Create servers
	server, metricsServer := NewServer()

	// Test that handlers are not nil
	assert.NotNil(t, server.Handler)
	assert.NotNil(t, metricsServer.Handler)

	// Test that handlers implement http.Handler interface
	var _ http.Handler = server.Handler
	var _ http.Handler = metricsServer.Handler

	// Test that metrics server handler is a ServeMux (no middleware)
	assert.IsType(t, &http.ServeMux{}, metricsServer.Handler)
}

// Benchmark tests for server creation
func BenchmarkNewServer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewServer()
	}
}

func BenchmarkServerHandler(b *testing.B) {
	server, _ := NewServer()
	req, _ := http.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server.Handler.ServeHTTP(rr, req)
	}
}

func BenchmarkMetricsServerHandler(b *testing.B) {
	_, metricsServer := NewServer()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metricsServer.Handler.ServeHTTP(rr, req)
	}
}
