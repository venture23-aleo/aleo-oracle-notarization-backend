package http

import (
	"net"
	"net/http"
	"strings"
)

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
