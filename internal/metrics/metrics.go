package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Request Metrics
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Attestation Metrics
	AttestationRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "attestation_requests_total",
			Help: "Total number of attestation requests",
		},
		[]string{"type", "status"},
	)

	AttestationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "attestation_duration_seconds",
			Help:    "Attestation processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	// SGX Quote Generation Metrics
	SgxQuoteGenerationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sgx_quote_generation_total",
			Help: "Total number of SGX quote generation attempts",
		},
		[]string{"status"},
	)

	SgxQuoteGenerationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "sgx_quote_generation_duration_seconds",
			Help:    "SGX quote generation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// Data Extraction Metrics
	DataExtractionRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "data_extraction_requests_total",
			Help: "Total number of data extraction requests",
		},
		[]string{"format", "status"},
	)

	DataExtractionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "data_extraction_duration_seconds",
			Help:    "Data extraction duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"format"},
	)

	// External HTTP Request Metrics
	ExternalHttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "external_http_requests_total",
			Help: "Total number of external HTTP requests",
		},
		[]string{"target", "status_code"},
	)

	ExternalHttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "external_http_request_duration_seconds",
			Help:    "External HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"target"},
	)

	// Price Feed Metrics
	PriceFeedRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "price_feed_requests_total",
			Help: "Total number of price feed requests",
		},
		[]string{"asset", "status"},
	)

	PriceFeedProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "price_feed_processing_duration_seconds",
			Help:    "Price feed processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"asset"},
	)

	// Random Number Generation Metrics
	RandomNumberGenerationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "random_number_generation_total",
			Help: "Total number of random number generation requests",
		},
		[]string{"status"},
	)

	RandomNumberGenerationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "random_number_generation_duration_seconds",
			Help:    "Random number generation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// System Metrics
	ActiveGoroutines = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_goroutines",
			Help: "Number of active goroutines",
		},
	)

	MemoryAllocBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_alloc_bytes",
			Help: "Current memory allocation in bytes",
		},
	)

	MemorySysBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_sys_bytes",
			Help: "Total memory obtained from system in bytes",
		},
	)

	// Enclave Health Metrics
	EnclaveHealthStatus = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "enclave_health_status",
			Help: "Enclave health status (1 = healthy, 0 = unhealthy)",
		},
	)

	// Business Metrics
	AttestationDataSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "attestation_data_size_bytes",
			Help:    "Size of attestation data in bytes",
			Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000},
		},
		[]string{"type"},
	)

	// Error Metrics
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors by type",
		},
		[]string{"type", "component"},
	)

	// Exchange API Metrics
	ExchangeApiErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "exchange_api_errors_total",
			Help: "Total number of exchange API errors",
		},
		[]string{"exchange", "error_code"},
	)

	PriceFeedExchangeCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "price_feed_exchange_count",
			Help: "Number of exchanges in the feed",
		},
		[]string{"feed"},
	)
)

// RecordHttpRequest records HTTP request metrics
func RecordHttpRequest(method, path string, statusCode int, duration float64) {
	HttpRequestsTotal.WithLabelValues(method, path, string(rune(statusCode))).Inc()
	HttpRequestDuration.WithLabelValues(method, path).Observe(duration)
}

// RecordAttestationRequest records attestation request metrics
func RecordAttestationRequest(attestationType, status string, duration float64) {
	AttestationRequestsTotal.WithLabelValues(attestationType, status).Inc()
	AttestationDuration.WithLabelValues(attestationType).Observe(duration)
}

// RecordSgxQuoteGeneration records SGX quote generation metrics
func RecordSgxQuoteGeneration(status string, duration float64) {
	SgxQuoteGenerationTotal.WithLabelValues(status).Inc()
	SgxQuoteGenerationDuration.Observe(duration)
}

// RecordDataExtraction records data extraction metrics
func RecordDataExtraction(format, status string, duration float64) {
	DataExtractionRequestsTotal.WithLabelValues(format, status).Inc()
	DataExtractionDuration.WithLabelValues(format).Observe(duration)
}

// RecordExternalHttpRequest records external HTTP request metrics
func RecordExternalHttpRequest(target string, statusCode int, duration float64) {
	ExternalHttpRequestsTotal.WithLabelValues(target, string(rune(statusCode))).Inc()
	ExternalHttpRequestDuration.WithLabelValues(target).Observe(duration)
}

// RecordPriceFeedRequest records price feed request metrics
func RecordPriceFeedRequest(asset, status string, duration float64) {
	PriceFeedRequestsTotal.WithLabelValues(asset, status).Inc()
	PriceFeedProcessingDuration.WithLabelValues(asset).Observe(duration)
}

// RecordRandomNumberGeneration records random number generation metrics
func RecordRandomNumberGeneration(status string, duration float64) {
	RandomNumberGenerationTotal.WithLabelValues(status).Inc()
	RandomNumberGenerationDuration.Observe(duration)
}

// RecordError records error metrics
func RecordError(errorType, component string) {
	ErrorsTotal.WithLabelValues(errorType, component).Inc()
}

// RecordAttestationDataSize records attestation data size metrics
func RecordAttestationDataSize(attestationType string, sizeBytes int) {
	AttestationDataSize.WithLabelValues(attestationType).Observe(float64(sizeBytes))
}

// RecordEnclaveHealthStatus records enclave health status
func RecordEnclaveHealthStatus(status int) {
	EnclaveHealthStatus.Set(float64(status))
}

// RecordExchangeApiError records exchange API error metrics
func RecordExchangeApiError(exchange string, errorCode string) {
	ExchangeApiErrorsTotal.WithLabelValues(exchange, errorCode).Inc()
}

// RecordPriceFeedExchangeCount records price feed exchange count
func RecordPriceFeedExchangeCount(feed string, count int) {
	PriceFeedExchangeCount.WithLabelValues(feed).Set(float64(count))
}
