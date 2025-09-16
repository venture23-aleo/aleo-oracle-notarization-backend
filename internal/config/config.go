package configs

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/sgx"
)

//go:embed config.json
var configFS embed.FS

type TokenExchanges map[string][]string

type ExchangeConfig struct {
	Name             string              `json:"name"`
	BaseURL          string              `json:"baseURL"`
	Symbols          map[string][]string `json:"symbols"`
	EndpointTemplate string              `json:"endpointTemplate"`
}

type ExchangesConfig map[string]ExchangeConfig

type PriceFeedConfig struct {
	ExchangesConfig      ExchangesConfig `json:"exchangesConfig"`
	TokenExchanges       TokenExchanges  `json:"tokenExchanges"`
	MinExchangesRequired int             `json:"minExchangesRequired"`
}

// AppConfig holds application-wide configuration
type AppConfig struct {
	Port               int             `json:"port"`
	MetricsPort        int             `json:"metricsPort"`
	PriceFeedConfig    PriceFeedConfig `json:"priceFeedConfig"`
	WhitelistedDomains []string        `json:"whitelistedDomains"`
	LogLevel           string          `json:"logLevel"`
}

type TokenTradingPairs map[string][]string

var (
	appConfigOnce sync.Once
	appConfig     AppConfig
	appConfigErr  error
)

var (
	tokenTradingPairsOnce sync.Once
	tokenTradingPairs     = make(TokenTradingPairs)
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

// GetWhitelistedDomains returns whitelisted domains from the app config
func GetWhitelistedDomains() []string {
	appConfig := GetAppConfig()
	return appConfig.WhitelistedDomains
}

// GetExchangesConfigs returns the exchanges configs from the app config
func GetExchangesConfigs() ExchangesConfig {
	appConfig := GetAppConfig()
	return appConfig.PriceFeedConfig.ExchangesConfig
}

// GetTokenExchanges returns the token exchanges from the app config
func GetTokenExchanges() TokenExchanges {
	appConfig := GetAppConfig()
	return appConfig.PriceFeedConfig.TokenExchanges
}

// GetMinExchangesRequired returns the minimum number of exchanges required from the app config
func GetMinExchangesRequired() int {
	appConfig := GetAppConfig()
	return appConfig.PriceFeedConfig.MinExchangesRequired
}

func loadTokenTradingPairs() {
	exchangesConfigs := GetExchangesConfigs()
	tokenExchanges := GetTokenExchanges()
	for token, exchanges := range tokenExchanges {
		for _, exchange := range exchanges {
			exchangeInfo, exists := exchangesConfigs[exchange]
			if !exists {
				logger.Error("Exchange not found", "exchange", exchange)
				continue
			}
			symbols, exists := exchangeInfo.Symbols[token]
			if !exists {
				logger.Error("No symbols found for token", "token", token)
				continue
			}
			for _, symbol := range symbols {
				if _, exists := tokenTradingPairs[token]; !exists {
					tokenTradingPairs[token] = []string{symbol}
				} else {
					tokenTradingPairs[token] = append(tokenTradingPairs[token], symbol)
				}
			}
		}
	}
}

func GetTokenTradingPairs() TokenTradingPairs {
	tokenTradingPairsOnce.Do(func() {
		loadTokenTradingPairs()
	})
	return tokenTradingPairs
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
	var tokenKeys []string

	exchangesConfigs := appConfig.PriceFeedConfig.ExchangesConfig
	tokenExchanges := appConfig.PriceFeedConfig.TokenExchanges
	minExchangesRequired := appConfig.PriceFeedConfig.MinExchangesRequired

	if minExchangesRequired < 1 {
		errors = append(errors, "Min exchanges required must be at least 1")
	}

	if len(exchangesConfigs) == 0 {
		errors = append(errors, "No exchange configurations found")
	}

	if len(tokenExchanges) == 0 {
		errors = append(errors, "No token exchanges found")
	}

	for exchangeKey, config := range exchangesConfigs {
		if config.Name == "" {
			errors = append(errors, fmt.Sprintf("Exchange %s: missing name", exchangeKey))
		}
		if config.BaseURL == "" {
			errors = append(errors, fmt.Sprintf("Exchange %s: missing baseURL", exchangeKey))
		}
		if len(config.Symbols) == 0 {
			errors = append(errors, fmt.Sprintf("Exchange %s: no symbols configured", exchangeKey))
		}
		if config.EndpointTemplate == "" {
			errors = append(errors, fmt.Sprintf("Exchange %s: no endpointTemplate configured", exchangeKey))
		}
		exchangeKeys = append(exchangeKeys, exchangeKey)
	}

	// Validate symbol exchanges mapping
	for token, exchanges := range tokenExchanges {
		if len(exchanges) == 0 {
			errors = append(errors, fmt.Sprintf("Token %s: no exchanges configured", token))
		}
		// Check that all referenced exchanges exist in exchange configs
		for _, exchange := range exchanges {
			if _, exists := exchangesConfigs[exchange]; !exists {
				errors = append(errors, fmt.Sprintf("Token %s: exchange %s not found in exchange configs", token, exchange))
			}
		}
		tokenKeys = append(tokenKeys, token)
	}

	// Enforce SGX startup check
	err := sgx.EnforceSGXStartup()
	if err != nil {
		errors = append(errors, fmt.Sprintf("SGX startup check failed: %v", err))
	}

	// Return combined error if any validation failed
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed:\n%s", strings.Join(errors, "\n"))
	}

	logger.Info("Configuration validation passed", "exchange_count", len(exchangesConfigs), "token_count", len(tokenExchanges), "exchanges", strings.Join(exchangeKeys, ","), "tokens", strings.Join(tokenKeys, ","))

	return nil
}
