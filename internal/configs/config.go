package configs

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

//go:embed config_data/app_config.json
var configFS embed.FS

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	WhitelistedDomains []string `json:"whitelistedDomains"`
	MaxRequestSize     string   `json:"maxRequestSize"`
	RateLimitPerMinute int      `json:"rateLimitPerMinute"`
	DDoSProtection     DDoSProtectionConfig `json:"ddosProtection"`
	WhitelistedIPs     []string `json:"whitelistedIPs"`
	BlockedIPs         []string `json:"blockedIPs"`
}

// DDoSProtectionConfig holds DDoS protection settings
type DDoSProtectionConfig struct {
	BurstProtection     BurstProtectionConfig     `json:"burstProtection"`
	SuspiciousActivity  SuspiciousActivityConfig  `json:"suspiciousActivity"`
	IPReputation        IPReputationConfig        `json:"ipReputation"`
	CacheSettings       CacheSettingsConfig       `json:"cacheSettings"`
}

// BurstProtectionConfig holds burst protection settings
type BurstProtectionConfig struct {
	MaxBurstRequests    int `json:"maxBurstRequests"`
	BurstWindowSeconds  int `json:"burstWindowSeconds"`
}

// SuspiciousActivityConfig holds suspicious activity detection settings
type SuspiciousActivityConfig struct {
	MaxRequestsPerSecond        int `json:"maxRequestsPerSecond"`
	SuspiciousThresholdPerMinute int `json:"suspiciousThresholdPerMinute"`
}

// IPReputationConfig holds IP reputation settings
type IPReputationConfig struct {
	BlacklistDuration string `json:"blacklistDuration"`
}

// CacheSettingsConfig holds cache-related settings
type CacheSettingsConfig struct {
	DDoSCacheDuration      string `json:"ddosCacheDuration"`
	BlacklistCacheDuration string `json:"blacklistCacheDuration"`
}

type ServerConfig struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	ReadTimeout  string `json:"readTimeout"`
	WriteTimeout string `json:"writeTimeout"`
	IdleTimeout  string `json:"idleTimeout"`
	CacheCleanupInterval string `json:"cacheCleanupInterval"`
}

type RateLimitConfig struct {
	MaxRequestsPerMinute int `json:"maxRequestsPerMinute"`
	BurstSize int `json:"burstSize"`
}


// AppConfig holds application-wide configuration
type AppConfig struct {
	Security SecurityConfig `json:"security"`
	Server ServerConfig `json:"server"`
	RateLimit RateLimitConfig `json:"rateLimit"`
}

var (
	appConfigOnce sync.Once
	appConfig AppConfig
	appConfigErr error
)

// GetAppConfig returns application configuration from embedded JSON file
func GetAppConfig() AppConfig {
	appConfigOnce.Do(func() {
		appConfig, appConfigErr = loadAppConfigFromFS("config_data/app_config.json")
	})
	return appConfig
}

// GetExchangeConfigsWithError returns exchange configs and any loading error
func GetAppConfigWithError() (AppConfig, error) {
	GetAppConfig() // Ensure initialization
	return appConfig, appConfigErr
}

// GetWhitelistedDomains returns whitelisted domains from embedded config
// Note: For SGX reproducibility, this must be deterministic and not use environment variables
func GetWhitelistedDomains() []string {
	// Use only embedded config for SGX reproducibility
	appConfig := GetAppConfig()
	return appConfig.Security.WhitelistedDomains
}

func loadAppConfigFromFS(path string) (AppConfig, error) {
	data, err := configFS.ReadFile(path)
	if err != nil {
		return AppConfig{}, fmt.Errorf("failed to read embedded app config file %s: %w", path, err)
	}
	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return AppConfig{}, fmt.Errorf("failed to parse embedded app config file %s: %w", path, err)
	}
	return config,nil
}


// ValidateConfigs validates that all configurations loaded correctly
// Should be called during server startup to catch configuration errors early
func ValidateConfigs() error {
	var errors []string
	
	// Load and validate app config
	appConfig, err := GetAppConfigWithError()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to load app config: %v", err))
	} else if appConfig.Security.WhitelistedDomains == nil {
		errors = append(errors, "No whitelisted domains found in app config")
	}
	
	// Load and validate exchange configs
	exchangeConfigs, err := GetExchangeConfigsWithError()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to load exchange configs: %v", err))
	} else if len(exchangeConfigs) == 0 {
		errors = append(errors, "No exchange configurations found")
	} else {
		// Validate each exchange config
		for exchangeKey, config := range exchangeConfigs {
			if config.Name == "" {
				errors = append(errors, fmt.Sprintf("Exchange %s: missing name", exchangeKey))
			}
			if config.BaseURL == "" {
				errors = append(errors, fmt.Sprintf("Exchange %s: missing baseURL", exchangeKey))
			}
			if len(config.Symbols) == 0 {
				errors = append(errors, fmt.Sprintf("Exchange %s: no symbols configured", exchangeKey))
			}
			if len(config.Endpoints) == 0 {
				errors = append(errors, fmt.Sprintf("Exchange %s: no endpoints configured", exchangeKey))
			}
		}
	}
	
	// Load and validate symbol exchanges
	symbolExchanges, err := GetSymbolExchangesWithError()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to load symbol exchanges: %v", err))
	} else if len(symbolExchanges) == 0 {
		errors = append(errors, "No symbol exchanges mapping found")
	} else {
		// Validate symbol exchanges mapping
		for symbol, exchanges := range symbolExchanges {
			if len(exchanges) == 0 {
				errors = append(errors, fmt.Sprintf("Symbol %s: no exchanges configured", symbol))
			}
			// Check that all referenced exchanges exist in exchange configs
			for _, exchange := range exchanges {
				if _, exists := exchangeConfigs[exchange]; !exists {
					errors = append(errors, fmt.Sprintf("Symbol %s: exchange %s not found in exchange configs", symbol, exchange))
				}
			}
		}
	}
	
	// Return combined error if any validation failed
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n%s", strings.Join(errors, "\n"))
	}
	
	return nil
}