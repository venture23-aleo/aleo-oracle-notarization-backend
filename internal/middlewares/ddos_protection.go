package middlewares

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
)

// DDoSProtectionData tracks various metrics for DDoS detection
type DDoSProtectionData struct {
	RequestCount    int       `json:"request_count"`
	LastRequestTime time.Time `json:"last_request_time"`
	BurstCount      int       `json:"burst_count"`
	BurstStartTime  time.Time `json:"burst_start_time"`
	IsBlacklisted   bool      `json:"is_blacklisted"`
	BlacklistTime   time.Time `json:"blacklist_time"`
}

// Create caches for different protection layers
var (
	ddosCache     *cache.Cache
	blacklistCache *cache.Cache
	mu            sync.RWMutex
	config        *configs.AppConfig
)

// InitializeDDoSProtection initializes the DDoS protection with config
func InitializeDDoSProtection(cfg *configs.AppConfig) {
	config = cfg

	cacheCleanupInterval, _ := time.ParseDuration(config.Server.CacheCleanupInterval)
	
	ddosCacheDuration, _ := time.ParseDuration(config.Security.DDoSProtection.CacheSettings.DDoSCacheDuration)
	blacklistCacheDuration, _ := time.ParseDuration(config.Security.DDoSProtection.IPReputation.BlacklistDuration)
	
	// Initialize caches with config values
	ddosCache = cache.New(
		ddosCacheDuration,
		cacheCleanupInterval,
	)
	
	blacklistCache = cache.New(
		blacklistCacheDuration,
		cacheCleanupInterval,
	)
}

// DDoSProtectionMiddleware provides comprehensive DDoS protection
func DDoSProtectionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		
		// Check if IP is whitelisted - bypass all protection
		if IsWhitelistedIP(ip) {
			w.Header().Set("X-Whitelisted", "true")
			next.ServeHTTP(w, r)
			return
		}
		
		// Check if IP is blacklisted
		if isBlacklisted(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "IP temporarily blocked due to suspicious activity",
			})
			return
		}
		
		// Get or create protection data
		data := getProtectionData(ip)
		
		// Check burst protection
		if !checkBurstProtection(ip, &data) {
			blacklistIP(ip)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Burst limit exceeded",
			})
			return
		}
		
		// Check suspicious activity
		if isSuspiciousActivity(ip, &data) {
			blacklistIP(ip)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Suspicious activity detected",
			})
			return
		}
		
		// Update protection data
		updateProtectionData(ip, &data)
		
		// Add API security headers
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("X-API-Version", "1.0")
		
		next.ServeHTTP(w, r)
	})
}

func getProtectionData(ip string) DDoSProtectionData {
	mu.RLock()
	defer mu.RUnlock()
	
	if cached, found := ddosCache.Get(ip); found {
		return cached.(DDoSProtectionData)
	}
	
	return DDoSProtectionData{
		RequestCount:    0,
		LastRequestTime: time.Now(),
		BurstCount:      0,
		BurstStartTime:  time.Now(),
		IsBlacklisted:   false,
	}
}

func checkBurstProtection(ip string, data *DDoSProtectionData) bool {
	now := time.Now()
	burstWindow := time.Duration(config.Security.DDoSProtection.BurstProtection.BurstWindowSeconds) * time.Second
	
	// Reset burst counter if window has passed
	if now.Sub(data.BurstStartTime) > burstWindow {
		data.BurstCount = 0
		data.BurstStartTime = now
	}
	
	data.BurstCount++
	
	// Check if burst limit exceeded
	if data.BurstCount > config.Security.DDoSProtection.BurstProtection.MaxBurstRequests {
		return false
	}
	
	return true
}

func isSuspiciousActivity(ip string, data *DDoSProtectionData) bool {
	now := time.Now()
	
	// Check requests per second
	if now.Sub(data.LastRequestTime) < time.Second {
		data.RequestCount++
		if data.RequestCount > config.Security.DDoSProtection.SuspiciousActivity.MaxRequestsPerSecond {
			return true
		}
	} else {
		data.RequestCount = 1
	}
	
	// Check for suspicious patterns
	if data.RequestCount > config.Security.DDoSProtection.SuspiciousActivity.SuspiciousThresholdPerMinute {
		return true
	}
	
	data.LastRequestTime = now
	return false
}

func updateProtectionData(ip string, data *DDoSProtectionData) {
	mu.Lock()
	defer mu.Unlock()
	
	ddosCache.Set(ip, *data, 0)
}

func isBlacklisted(ip string) bool {
	mu.RLock()
	defer mu.RUnlock()
	
	_, found := blacklistCache.Get(ip)
	return found
}

func blacklistIP(ip string) {
	mu.Lock()
	defer mu.Unlock()
	
	blacklistCache.Set(ip, true, 0)
}

// GetDDoSStats returns current protection statistics
func GetDDoSStats() map[string]interface{} {
	mu.RLock()
	defer mu.RUnlock()
	
	ddosItems := ddosCache.Items()
	blacklistItems := blacklistCache.Items()
	
	burstWindow := time.Duration(config.Security.DDoSProtection.BurstProtection.BurstWindowSeconds) * time.Second
	
	return map[string]interface{}{
		"monitored_ips":     len(ddosItems),
		"blacklisted_ips":   len(blacklistItems),
		"burst_limit":       config.Security.DDoSProtection.BurstProtection.MaxBurstRequests,
		"burst_window":      burstWindow.String(),
		"suspicious_threshold": config.Security.DDoSProtection.SuspiciousActivity.SuspiciousThresholdPerMinute,
	}
} 