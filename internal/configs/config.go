package configs

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
)

//go:embed config.json
var configFS embed.FS

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
	Port               int             `json:"port"`
	MetricsPort        int             `json:"metricsPort"`
	PriceFeedConfig    PriceFeedConfig `json:"priceFeedConfig"`
	WhitelistedDomains []string        `json:"whitelistedDomains"`
	LogLevel           string          `json:"logLevel"`
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
	_, configErr := GetAppConfigWithError()
	if configErr != nil {
		errors = append(errors, fmt.Sprintf("Failed to load app config: %v", configErr))
	} else if appConfig.WhitelistedDomains == nil {
		errors = append(errors, "No whitelisted domains found in app config")
	}

	var exchangeKeys []string
	var symbolKeys []string
	exchangeConfigs := appConfig.PriceFeedConfig.Exchanges
	symbolExchanges := appConfig.PriceFeedConfig.SymbolExchanges

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

	logger.Info("Configuration validation passed", "exchange_count", len(exchangeConfigs), "symbol_count", len(symbolExchanges), "exchanges", strings.Join(exchangeKeys, ","), "symbols", strings.Join(symbolKeys, ","))

	return nil
}
