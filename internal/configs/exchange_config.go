package configs

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed config_data/exchange_config.json config_data/symbol_exchanges.json
var exchangeConfigFS embed.FS

// ExchangeConfig defines the configuration for each exchange
type ExchangeConfig struct {
	Name      string            `json:"name"`
	BaseURL   string            `json:"baseURL"`
	Symbols   map[string]string `json:"symbols"`   // Maps our symbol to exchange symbol
	Endpoints map[string]string `json:"endpoints"` // Maps symbol to endpoint path
}

// ExchangeConfigs holds all exchange configurations
type ExchangeConfigs map[string]ExchangeConfig

// SymbolExchanges maps symbols to the exchanges that support them
type SymbolExchanges map[string][]string

// ConfigManager handles loading and managing exchange configurations
type ConfigManager struct {
	configPath string
	environment string
}

// NewConfigManager creates a new config manager
func NewConfigManager(configPath, environment string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
		environment: environment,
	}
}

// Global singleton instances with lazy initialization
var (
	exchangeConfigsOnce sync.Once
	symbolExchangesOnce sync.Once
	exchangeConfigs     ExchangeConfigs
	symbolExchanges     SymbolExchanges
	exchangeConfigsErr  error
	symbolExchangesErr  error
)

// GetExchangeConfigs returns exchange configurations from embedded JSON files
// Uses singleton pattern with lazy initialization - loads once, reuses for all requests
func GetExchangeConfigs() ExchangeConfigs {
	exchangeConfigsOnce.Do(func() {
		exchangeConfigs, exchangeConfigsErr = loadExchangeConfigsFromFS("config_data/exchange_config.json")
		if exchangeConfigsErr != nil {
			// Return empty configs if file not found or invalid
			exchangeConfigs = ExchangeConfigs{}
		}	
	})
	return exchangeConfigs
}

// GetSymbolExchanges returns symbol to exchanges mapping from embedded JSON files
// Uses singleton pattern with lazy initialization - loads once, reuses for all requests
func GetSymbolExchanges() SymbolExchanges {
	symbolExchangesOnce.Do(func() {
		symbolExchanges, symbolExchangesErr = loadSymbolExchangesFromFS("config_data/symbol_exchanges.json")
		if symbolExchangesErr != nil {
			// Return empty mappings if file not found or invalid
			symbolExchanges = SymbolExchanges{}
		}
	})
	return symbolExchanges
}

// GetExchangeConfigsWithError returns exchange configs and any loading error
func GetExchangeConfigsWithError() (ExchangeConfigs, error) {
	GetExchangeConfigs() // Ensure initialization
	return exchangeConfigs, exchangeConfigsErr
}

// GetSymbolExchangesWithError returns symbol exchanges and any loading error
func GetSymbolExchangesWithError() (SymbolExchanges, error) {
	GetSymbolExchanges() // Ensure initialization
	return symbolExchanges, symbolExchangesErr
}

func loadExchangeConfigsFromFS(path string) (ExchangeConfigs, error) {
	data, err := exchangeConfigFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded config file %s: %w", path, err)
	}
	var configs ExchangeConfigs
	if err := json.Unmarshal(data, &configs); err != nil {
		return nil, fmt.Errorf("failed to parse embedded config file %s: %w", path, err)
	}
	return configs, nil
}

func loadSymbolExchangesFromFS(path string) (SymbolExchanges, error) {
	data, err := exchangeConfigFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded symbol exchanges file %s: %w", path, err)
	}
	var symbols SymbolExchanges
	if err := json.Unmarshal(data, &symbols); err != nil {
		return nil, fmt.Errorf("failed to parse embedded symbol exchanges file %s: %w", path, err)
	}
	return symbols, nil
}