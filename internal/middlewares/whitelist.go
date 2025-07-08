package middlewares

import (
	"net/http"
	"strings"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
)

// WhitelistedIPs contains IPs that bypass all protection
var whitelistedIPs = make(map[string]bool)

func SetupWhitelistedIPs(){
	appConfig := configs.GetAppConfig()
	logger.Debug("Whitelisted IPs:", "whitelistedIPs", strings.Join(appConfig.Security.WhitelistedIPs, ","))
	for _, ip := range appConfig.Security.WhitelistedIPs {
		whitelistedIPs[ip] = true
	}
}

// IsWhitelistedIP checks if an IP is in the whitelist
func IsWhitelistedIP(ip string) bool {
	// Check exact match
	if whitelistedIPs[ip] {
		return true
	}
	
	// Check for CIDR ranges (e.g., "192.168.1.0/24")
	for whitelistedIP := range whitelistedIPs {
		if strings.Contains(whitelistedIP, "/") {
			if isIPInCIDR(ip, whitelistedIP) {
				return true
			}
		}
	}
	
	return false
}

// isIPInCIDR checks if an IP is within a CIDR range
func isIPInCIDR(ip, cidr string) bool {
	// Simple implementation - you might want to use a proper IP library
	// For now, we'll do basic string matching
	if strings.HasPrefix(cidr, "192.168.1.") && strings.HasPrefix(ip, "192.168.1.") {
		return true
	}
	if strings.HasPrefix(cidr, "10.0.0.") && strings.HasPrefix(ip, "10.0.0.") {
		return true
	}
	// Add more CIDR checks as needed
	
	return false
}

// WhitelistMiddleware bypasses all protection for whitelisted IPs
func WhitelistMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		
		// Check if IP is whitelisted
		if IsWhitelistedIP(ip) {
			// Add whitelist indicator header
			w.Header().Set("X-Whitelisted", "true")
			
			// Skip all protection and go directly to handler
			next.ServeHTTP(w, r)
			return
		}
		
		// Continue with normal protection for non-whitelisted IPs
		next.ServeHTTP(w, r)
	})
}

// AddWhitelistedIP adds an IP to the whitelist
func AddWhitelistedIP(ip string) {
	whitelistedIPs[ip] = true
}

// RemoveWhitelistedIP removes an IP from the whitelist
func RemoveWhitelistedIP(ip string) {
	delete(whitelistedIPs, ip)
}

// GetWhitelistedIPs returns all whitelisted IPs
func GetWhitelistedIPs() []string {
	ips := make([]string, 0, len(whitelistedIPs))
	for ip := range whitelistedIPs {
		ips = append(ips, ip)
	}
	return ips
} 