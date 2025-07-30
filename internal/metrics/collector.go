package metrics

import (
	"runtime"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

// SystemMetricsCollector collects system-level metrics
type SystemMetricsCollector struct {
	stopChan chan struct{}
}

// NewSystemMetricsCollector creates a new system metrics collector
func NewSystemMetricsCollector() *SystemMetricsCollector {
	return &SystemMetricsCollector{
		stopChan: make(chan struct{}),
	}
}

// Start begins collecting system metrics
func (c *SystemMetricsCollector) Start() {
	logger.Info("Starting system metrics collector")
	go c.collect()
}

// Stop stops collecting system metrics
func (c *SystemMetricsCollector) Stop() {
	logger.Info("Stopping system metrics collector")
	close(c.stopChan)
}

// collect periodically collects system metrics
func (c *SystemMetricsCollector) collect() {
	ticker := time.NewTicker(30 * time.Second) // Collect every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.updateMetrics()
		case <-c.stopChan:
			return
		}
	}
}

// updateMetrics updates all system metrics
func (c *SystemMetricsCollector) updateMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Update memory metrics
	MemoryAllocBytes.Set(float64(m.Alloc))
	MemorySysBytes.Set(float64(m.Sys))

	// Update goroutine count
	ActiveGoroutines.Set(float64(runtime.NumGoroutine()))

	// Log metrics for debugging
	logger.Debug("System metrics updated",
		"memory_alloc_mb", m.Alloc/1024/1024,
		"memory_sys_mb", m.Sys/1024/1024,
		"goroutines", runtime.NumGoroutine(),
	)
}
