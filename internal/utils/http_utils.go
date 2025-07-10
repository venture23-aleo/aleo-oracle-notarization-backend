package utils

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
)

// WriteJsonSuccess writes a JSON success response with optional message and data
func WriteJsonSuccess(w http.ResponseWriter, statusCode int, data interface{}) {

	// Set the content type.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Encode the data.
	json.NewEncoder(w).Encode(data)
}

// WriteJsonError writes a JSON error response with message and error code
func WriteJsonError(w http.ResponseWriter, statusCode int, appError appErrors.AppError, requestID string) {

	// Set the content type.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Set the request ID.
	appError.RequestID = requestID

	// Encode the error response.
	json.NewEncoder(w).Encode(appError)
}

func GetRetryableHTTPClient(maxRetries int) *retryablehttp.Client {
	retryClient := retryablehttp.NewClient()
	retryClient.Logger = logger.Logger
	retryClient.RetryWaitMin = 2 * time.Second
	retryClient.RetryWaitMax = 3 * time.Second
	retryClient.RetryMax = maxRetries
	return retryClient
}

// getClientIP extracts the real client IP from various headers
func GetClientIP(r *http.Request) string {
	// Check for X-Real-IP header (set by nginx/proxy)
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		if isValidIP(realIP) {
			return realIP
		}
	}

	// Check for X-Forwarded-For header (set by load balancers)
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
		// Take the first valid IP (the original client)
		ips := strings.Split(fwd, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" && ip != "unknown" && isValidIP(ip) {
				return ip
			}
		}
	}

	// Check for X-Forwarded header
	if xForwarded := r.Header.Get("X-Forwarded"); xForwarded != "" {
		if isValidIP(xForwarded) {
			return xForwarded
		}
	}

	// Fallback to RemoteAddr (remove port if present)
	remoteAddr := r.RemoteAddr
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		remoteAddr = host
	}

	return remoteAddr
}

// isValidIP checks if a string is a valid IP address
func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}
