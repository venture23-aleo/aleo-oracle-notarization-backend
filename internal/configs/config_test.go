package configs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigs(t *testing.T) {
	// Test that our current configs are valid
	err := ValidateConfigs()
	require.Nil(t, err, "Configuration validation failed")

	// Test that we can get configs
	exchangeConfigs := GetExchangesConfigs()
	require.NotEmpty(t, exchangeConfigs, "No exchange configurations returned")

	symbolExchanges := GetSymbolExchanges()
	require.NotEmpty(t, symbolExchanges, "No symbol exchanges returned")

	// Log some info for debugging
	t.Logf("Loaded %d exchange configurations", len(exchangeConfigs))
	t.Logf("Loaded %d symbol mappings", len(symbolExchanges))

	// Test valid exchanges
	expectedExchanges := []string{"binance", "coinbase"}
	for _, expected := range expectedExchanges {
		assert.Contains(t, exchangeConfigs, expected, "Expected exchange %s not found in configs", expected)
	}

	// Test valid symbols
	expectedSymbols := []string{"BTC", "ETH", "ALEO"}
	for _, expected := range expectedSymbols {
		assert.Contains(t, symbolExchanges, expected, "Expected symbol %s not found in mappings", expected)
	}

	// Test invalid symbols
	invalidSymbols := []string{"INVALID", "INVALID2"}
	for _, invalid := range invalidSymbols {
		assert.NotContains(t, symbolExchanges, invalid, "Invalid symbol %s found in mappings", invalid)
	}

	// Test invalid exchanges
	invalidExchanges := []string{"INVALID", "INVALID2"}
	for _, invalid := range invalidExchanges {
		assert.NotContains(t, exchangeConfigs, invalid, "Invalid exchange %s found in configs", invalid)
	}
}
