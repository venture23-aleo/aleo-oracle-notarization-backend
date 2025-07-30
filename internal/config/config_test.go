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

	tokenExchanges := GetTokenExchanges()
	require.NotEmpty(t, tokenExchanges, "No token exchanges returned")

	// Log some info for debugging
	t.Logf("Loaded %d exchange configurations", len(exchangeConfigs))
	t.Logf("Loaded %d token mappings", len(tokenExchanges))

	// Test valid exchanges
	expectedExchanges := []string{"binance", "coinbase"}
	for _, expected := range expectedExchanges {
		assert.Contains(t, exchangeConfigs, expected, "Expected exchange %s not found in configs", expected)
	}

	// Test valid symbols
	expectedTokens := []string{"BTC", "ETH", "ALEO"}
	for _, expected := range expectedTokens {
		assert.Contains(t, tokenExchanges, expected, "Expected token %s not found in mappings", expected)
	}

	// Test invalid symbols
	invalidTokens := []string{"INVALID", "INVALID2"}
	for _, invalid := range invalidTokens {
		assert.NotContains(t, tokenExchanges, invalid, "Invalid token %s found in mappings", invalid)
	}

	// Test invalid exchanges
	invalidExchanges := []string{"INVALID", "INVALID2"}
	for _, invalid := range invalidExchanges {
		assert.NotContains(t, exchangeConfigs, invalid, "Invalid exchange %s found in configs", invalid)
	}
}
