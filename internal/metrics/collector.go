package metrics

import (
	"runtime"
	"sync"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

// SystemMetricsCollector collects system-level metrics
type SystemMetricsCollector struct {
	stopChan chan struct{}
	mu       sync.Mutex
	started  bool
}

// NewSystemMetricsCollector creates a new system metrics collector
func NewSystemMetricsCollector() *SystemMetricsCollector {
	return &SystemMetricsCollector{
		stopChan: make(chan struct{}),
	}
}

// Start begins collecting system metrics
func (c *SystemMetricsCollector) Start() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.started {
		logger.Warn("System metrics collector already started")
		return
	}
	logger.Info("Starting system metrics collector")
	// Reinitialize stop channel in case of restart after Stop.
	c.stopChan = make(chan struct{})
	c.started = true
	go c.collect()
}

// Stop stops collecting system metrics
func (c *SystemMetricsCollector) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.started {
		logger.Warn("System metrics collector not running")
		return
	}
	logger.Info("Stopping system metrics collector")
	// Make stop idempotent by closing only once.
	select {
	case <-c.stopChan:
		// already closed
	default:
		close(c.stopChan)
	}
	c.started = false
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
