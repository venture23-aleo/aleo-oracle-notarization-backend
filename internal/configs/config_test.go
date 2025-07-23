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

	coinExchanges := GetCoinExchanges()
	require.NotEmpty(t, coinExchanges, "No coin exchanges returned")

	// Log some info for debugging
	t.Logf("Loaded %d exchange configurations", len(exchangeConfigs))
	t.Logf("Loaded %d coin mappings", len(coinExchanges))

	// Test valid exchanges
	expectedExchanges := []string{"binance", "coinbase"}
	for _, expected := range expectedExchanges {
		assert.Contains(t, exchangeConfigs, expected, "Expected exchange %s not found in configs", expected)
	}

	// Test valid symbols
	expectedCoins := []string{"BTC", "ETH", "ALEO"}
	for _, expected := range expectedCoins {
		assert.Contains(t, coinExchanges, expected, "Expected coin %s not found in mappings", expected)
	}

	// Test invalid symbols
	invalidCoins := []string{"INVALID", "INVALID2"}
	for _, invalid := range invalidCoins {
		assert.NotContains(t, coinExchanges, invalid, "Invalid coin %s found in mappings", invalid)
	}

	// Test invalid exchanges
	invalidExchanges := []string{"INVALID", "INVALID2"}
	for _, invalid := range invalidExchanges {
		assert.NotContains(t, exchangeConfigs, invalid, "Invalid exchange %s found in configs", invalid)
	}
}
