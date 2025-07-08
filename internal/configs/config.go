package configs

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
)

//go:embed config.json
var configFS embed.FS

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	WhitelistedDomains []string             `json:"whitelistedDomains"`
	MaxRequestSize     string               `json:"maxRequestSize"`
	DDoSProtection     DDoSProtectionConfig `json:"ddosProtection"`
	WhitelistedIPs     []string             `json:"whitelistedIPs"`
	BlockedIPs         []string             `json:"blockedIPs"`
	RateLimit          RateLimitConfig      `json:"rateLimit"`
}

// DDoSProtectionConfig holds DDoS protection settings
type DDoSProtectionConfig struct {
	BurstProtection    BurstProtectionConfig    `json:"burstProtection"`
	SuspiciousActivity SuspiciousActivityConfig `json:"suspiciousActivity"`
	IPReputation       IPReputationConfig       `json:"ipReputation"`
	CacheSettings      CacheSettingsConfig      `json:"cacheSettings"`
}

// BurstProtectionConfig holds burst protection settings
type BurstProtectionConfig struct {
	MaxBurstRequests   int `json:"maxBurstRequests"`
	BurstWindowSeconds int `json:"burstWindowSeconds"`
}

// SuspiciousActivityConfig holds suspicious activity detection settings
type SuspiciousActivityConfig struct {
	MaxRequestsPerSecond         int `json:"maxRequestsPerSecond"`
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
	Port                 int    `json:"port"`
	Host                 string `json:"host"`
	ReadTimeout          string `json:"readTimeout"`
	WriteTimeout         string `json:"writeTimeout"`
	IdleTimeout          string `json:"idleTimeout"`
	CacheCleanupInterval string `json:"cacheCleanupInterval"`
}

type RateLimitConfig struct {
	MaxRequestsPerMinute int `json:"maxRequestsPerMinute"`
	BurstSize            int `json:"burstSize"`
}

type SymbolExchanges map[string][]string

type ExchangeConfig struct {
	Name      string            `json:"name"`
	BaseURL   string            `json:"baseURL"`
	Symbols   map[string]string `json:"symbols"`
	Endpoints map[string]string `json:"endpoints"`
}

type ExchangesConfig map[string]ExchangeConfig

type PriceFeedConfig struct {
	Exchanges       ExchangesConfig `json:"exchanges"`
	SymbolExchanges SymbolExchanges `json:"symbolExchanges"`
}

// AppConfig holds application-wide configuration
type AppConfig struct {
	Port                 int             `json:"port"`
	Security             SecurityConfig  `json:"security"`
	PriceFeedConfig      PriceFeedConfig `json:"priceFeedConfig"`
	WhitelistedDomains   []string        `json:"whitelistedDomains"`
	CacheCleanupInterval string          `json:"cacheCleanupInterval"`
}

var (
	appConfigOnce sync.Once
	appConfig     AppConfig
	appConfigErr  error
)

// GetAppConfig returns application configuration from embedded JSON file
func GetAppConfig() AppConfig {
	appConfigOnce.Do(func() {
		appConfig, appConfigErr = loadAppConfigFromFS("config.json")
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
	return appConfig.WhitelistedDomains
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

	return config, nil
}

func GetExchangesConfigs() ExchangesConfig {
	appConfig := GetAppConfig()
	return appConfig.PriceFeedConfig.Exchanges
}

func GetSymbolExchanges() SymbolExchanges {
	appConfig := GetAppConfig()
	return appConfig.PriceFeedConfig.SymbolExchanges
}

// ValidateConfigs validates that all configurations loaded correctly
// Should be called during server startup to catch configuration errors early
func ValidateConfigs() error {
	var errors []string

	// Load and validate app config
	appConf, err := GetAppConfigWithError()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to load app config: %v", err))
	} else if appConfig.WhitelistedDomains == nil {
		errors = append(errors, "No whitelisted domains found in app config")
	}

	var exchangeKeys []string
	var symbolKeys []string
	exchangeConfigs := appConf.PriceFeedConfig.Exchanges
	symbolExchanges := appConf.PriceFeedConfig.SymbolExchanges

	if len(exchangeConfigs) == 0 {
		errors = append(errors, "No exchange configurations found")
	}

	if len(symbolExchanges) == 0 {
		errors = append(errors, "No symbol exchanges found")
	}

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
		exchangeKeys = append(exchangeKeys, exchangeKey)
	}

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
		symbolKeys = append(symbolKeys, symbol)
	}

	// Return combined error if any validation failed
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n%s", strings.Join(errors, "\n"))
	}

	if os.Getenv("PORT") != "" {
		port, err := strconv.Atoi(os.Getenv("PORT"))
		logger.Info("PORT", "PORT", port)
		if err != nil {
			return fmt.Errorf("failed to parse PORT environment variable: %w", err)
		}
		appConfig.Port = port
	}

	if os.Getenv("WHITELISTED_DOMAINS") != "" {
		whitelistedDomains := strings.Split(os.Getenv("WHITELISTED_DOMAINS"), ",")
		appConfig.WhitelistedDomains = whitelistedDomains
	}

	logger.Info("Configuration validation passed", "exchange_count", len(exchangeConfigs), "symbol_count", len(symbolExchanges), "exchanges", strings.Join(exchangeKeys, ","), "symbols", strings.Join(symbolKeys, ","))

	return nil
}
