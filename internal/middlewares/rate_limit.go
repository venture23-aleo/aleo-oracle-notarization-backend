package middlewares

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
	"golang.org/x/time/rate"
)

var (
	maxRequestsPerMinute int
	cleanupInterval      time.Duration
	burstSize            int
	limiters             map[string]*rate.Limiter
	limiterLastAccess    map[string]time.Time
	lastTokenConsumption map[string]time.Time // Track when last token was consumed
	limitersMutex        sync.RWMutex
)

// Token bucket rate limiting using golang.org/x/time/rate

func InitializeRateLimit(cfg *configs.AppConfig) {
	cleanupInterval, _ = time.ParseDuration(cfg.CacheCleanupInterval)
	maxRequestsPerMinute = cfg.Security.RateLimit.MaxRequestsPerMinute
	burstSize = cfg.Security.RateLimit.BurstSize

	limiters = make(map[string]*rate.Limiter)
	limiterLastAccess = make(map[string]time.Time)
	lastTokenConsumption = make(map[string]time.Time)
	
	// Start cleanup goroutine to remove inactive limiters
	go cleanupInactiveLimiters()
}

// IPRateLimitMiddleware limits requests per IP using token bucket algorithm
func IPRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract client IP
		ip := getClientIP(r)
		
		// Check if IP is whitelisted - bypass rate limiting
		if IsWhitelistedIP(ip) {
			logger.Debug("IP ", ip, "is whitelisted, bypassing rate limit")
			w.Header().Set("X-Whitelisted", "true")
			next.ServeHTTP(w, r)
			return
		}
		
		// Get or create rate limiter for this IP
		limiter := getOrCreateLimiter(ip)
		
		// Get current token count before consuming
		currentTokens := int(limiter.Tokens())
		logger.Debug("IP ", ip, "current tokens", currentTokens)
		
		// Check if we have tokens available (without consuming)
		if currentTokens < 1 {
			logger.Debug("Rate limit exceeded", "ip", ip)
			// Rate limit exceeded - calculate when next token will be available
			// Get the last time a token was consumed and add the regeneration time
			limitersMutex.RLock()
			lastConsumption, exists := lastTokenConsumption[ip]
			limitersMutex.RUnlock()
			
			var resetTime time.Time
			if exists {
				// Calculate reset time based on when last token was consumed
				tokenRegenerationTime := time.Duration(float64(time.Minute) / float64(maxRequestsPerMinute))
				resetTime = lastConsumption.Add(tokenRegenerationTime)
			} else {
				// Fallback if no consumption record exists
				resetTime = time.Now().Add(time.Duration(float64(time.Minute) / float64(maxRequestsPerMinute)))
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(maxRequestsPerMinute))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", resetTime.Format(time.RFC3339))
			w.WriteHeader(http.StatusTooManyRequests)
			
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "rate limit exceeded",
				"message": "Too many requests from this IP. Please try again later.",
				"limit":   maxRequestsPerMinute,
				"reset":   resetTime.Format(time.RFC3339),
			})
			return
		}
		
		// Consume a token
		limiter.Allow() // This will consume 1 token since we know it's available

		// Record when this token was consumed
		limitersMutex.Lock()
		lastTokenConsumption[ip] = time.Now()
		limitersMutex.Unlock()
		
		// Calculate reset time (when the next token will be available)
		// For token bucket with burst=1, this is when the bucket will have 1 token again
		resetTime := time.Now().Add(time.Duration(float64(time.Minute) / float64(maxRequestsPerMinute)))
		
		remaining := currentTokens - 1 // Subtract the token we just consumed
		if remaining < 0 {
			remaining = 0
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(maxRequestsPerMinute))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		w.Header().Set("X-RateLimit-Reset", resetTime.Format(time.RFC3339))
		
		next.ServeHTTP(w, r)
	})
}

// getOrCreateLimiter returns a rate limiter for the given IP
func getOrCreateLimiter(ip string) *rate.Limiter {
	limitersMutex.Lock()
	defer limitersMutex.Unlock()
	
	if limiter, exists := limiters[ip]; exists {
		// Update last access time
		limiterLastAccess[ip] = time.Now()
		logger.Debug("Using existing limiter for IP ", "ip", ip)
		return limiter
	}
	
	// Create new limiter: rate per second and burst size
	// For "X requests per minute", rate = X/60 tokens per second
	ratePerSecond := float64(maxRequestsPerMinute) / 60.0
	burstSize := burstSize
	
	logger.Debug("Creating new limiter for IP ", "ip", ip, "rate", ratePerSecond, "burst", burstSize)
	
	limiter := rate.NewLimiter(rate.Limit(ratePerSecond), burstSize)
	limiters[ip] = limiter
	limiterLastAccess[ip] = time.Now()
	
	return limiter
}

// cleanupInactiveLimiters removes limiters for IPs that haven't been active
func cleanupInactiveLimiters() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		limitersMutex.Lock()
		
		logger.Debug("Cleaning up inactive limiters")
		// Remove limiters inactive for more than 5 minutes
		cutoff := time.Now().Add(-5 * time.Minute)
		for ip, lastAccess := range limiterLastAccess {
			if lastAccess.Before(cutoff) {
				delete(limiters, ip)
				delete(limiterLastAccess, ip)
				delete(lastTokenConsumption, ip)
			}
		}
		
		limitersMutex.Unlock()
	}
}

// getClientIP extracts the real client IP from various headers
func getClientIP(r *http.Request) string {
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

// GetRateLimitStats returns current cache statistics (useful for monitoring)
func GetRateLimitStats() map[string]interface{} {
	limitersMutex.RLock()
	defer limitersMutex.RUnlock()
	
	return map[string]interface{}{
		"active_limiters": len(limiters),
		"rate_limit":     maxRequestsPerMinute,
		"cleanup_interval": cleanupInterval.String(),
		"algorithm":      "token_bucket",
	}
} 