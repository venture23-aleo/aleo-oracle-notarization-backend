package data_extraction

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
)

// TestPriceFeedClient_BTCPrice tests the BTC price feed functionality
func TestPriceFeedClient_BTCPrice(t *testing.T) {
	// Create mock HTTP server that returns realistic exchange responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Return different responses based on the exchange endpoint
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v3/ticker/24hr"):
			// Binance response
			response := map[string]interface{}{
				"symbol":      "BTCUSDT",
				"lastPrice":   "50000.00",
				"volume":      "1000.50",
				"priceChange": "100.00",
				"bidPrice":    "49999.00",
				"askPrice":    "50001.00",
				"openPrice":   "49900.00",
				"highPrice":   "50100.00",
				"lowPrice":    "49800.00",
			}
			json.NewEncoder(w).Encode(response)

		case strings.HasPrefix(r.URL.Path, "/v5/market/tickers"):
			// Bybit response
			response := map[string]interface{}{
				"result": map[string]interface{}{
					"list": []map[string]interface{}{
						{
							"lastPrice": "50100.00",
							"volume24h": "800.25",
							"symbol":    "BTCUSDT",
							"bidPrice":  "50099.00",
							"askPrice":  "50101.00",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		case strings.HasPrefix(r.URL.Path, "/products/BTC-USD/ticker"):
			// Coinbase response
			response := map[string]interface{}{
				"price":  "50200.00",
				"volume": "1200.75",
				"bid":    "50199.00",
				"ask":    "50201.00",
				"open":   "50100.00",
				"high":   "50300.00",
				"low":    "50000.00",
			}
			json.NewEncoder(w).Encode(response)

		case strings.HasPrefix(r.URL.Path, "/v2/public/get-ticker"):
			// Crypto.com response
			response := map[string]interface{}{
				"result": map[string]interface{}{
					"data": []map[string]interface{}{
						{
							"k": "50300.00",
							"v": "900.30",
							"i": "BTC_USDT",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)

		default:
			// Return 404 for unknown endpoints
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create custom exchange configs using the mock server
	customConfigs := map[string]configs.ExchangeConfig{
		"binance": {
			Name:    "Binance",
			BaseURL: server.URL, // Remove "http://" prefix
			Symbols: map[string][]string{
				"BTC": {"BTCUSDT"},
			},
			EndpointTemplate: "/api/v3/ticker/24hr?symbol={symbol}",
		},
		"bybit": {
			Name:    "Bybit",
			BaseURL: server.URL,
			Symbols: map[string][]string{"BTC": {"BTCUSDT"}},
			EndpointTemplate: "/v5/market/tickers?category=spot&symbol={symbol}",
		},
		"coinbase": {
			Name:    "Coinbase",
			BaseURL: server.URL,
			Symbols: map[string][]string{"BTC": {"BTC-USD"}},
			EndpointTemplate: "/products/BTC-USD/ticker",
		},
		"crypto.com": {
			Name:    "Crypto.com",
			BaseURL: server.URL,
			Symbols: map[string][]string{"BTC": {"BTC_USDT"}},
			EndpointTemplate: "/v2/public/get-ticker?instrument_name={symbol}",
		},
	}

	customCoins := map[string][]string{
		"BTC": {"binance", "bybit", "coinbase", "crypto.com"},
	}

	// Create client with custom configs
	client := &PriceFeedClient{
		exchangeConfigs: customConfigs,
		coinExchanges:   customCoins,
	}

	// Test fetching individual exchange prices
	t.Run("FetchIndividualExchangePrices", func(t *testing.T) {
		// Test Binance
		price, err := client.FetchPriceFromExchange(context.Background(), "binance", "BTC", "BTCUSDT")
		assert.Nil(t, err)
		assert.NotNil(t, price)
		assert.Equal(t, "Binance", price.Exchange)
		assert.Equal(t, 50000.0, price.Price)
		assert.Equal(t, 1000.5, price.Volume)

		// Test Bybit
		price, err = client.FetchPriceFromExchange(context.Background(), "bybit", "BTC", "BTCUSDT")
		assert.Nil(t, err)
		assert.NotNil(t, price)
		assert.Equal(t, "Bybit", price.Exchange)
		assert.Equal(t, 50100.0, price.Price)
		assert.Equal(t, 800.25, price.Volume)

		// Test Coinbase
		price, err = client.FetchPriceFromExchange(context.Background(), "coinbase", "BTC", "BTC-USD")
		assert.Nil(t, err)
		assert.NotNil(t, price)
		assert.Equal(t, "Coinbase", price.Exchange)
		assert.Equal(t, 50200.0, price.Price)
		assert.Equal(t, 1200.75, price.Volume)

		// Test Crypto.com
		price, err = client.FetchPriceFromExchange(context.Background(), "crypto.com", "BTC", "BTC_USDT")
		assert.Nil(t, err)
		assert.NotNil(t, price)
		assert.Equal(t, "Crypto.com", price.Exchange)
		assert.Equal(t, 50300.0, price.Price)
		assert.Equal(t, 900.30, price.Volume)
	})

	// Test getting aggregated price feed
	t.Run("GetAggregatedPriceFeed", func(t *testing.T) {
		result, err := client.GetPriceFeed(context.Background(), "BTC")
		assert.Nil(t, err)
		assert.NotNil(t, result)

		// Check basic structure
		assert.Equal(t, "BTC", result.Symbol)
		assert.True(t, result.Success)
		assert.Equal(t, 4, result.ExchangeCount)

		// Check that we have prices from all exchanges
		assert.Equal(t, 4, len(result.ExchangePrices))

		// Verify individual exchange prices
		expectedPrices := map[string]float64{
			"Binance":    50000.0,
			"Bybit":      50100.0,
			"Coinbase":   50200.0,
			"Crypto.com": 50300.0,
		}

		for _, price := range result.ExchangePrices {
			expectedPrice, exists := expectedPrices[price.Exchange]
			assert.True(t, exists)
			assert.Equal(t, expectedPrice, price.Price)
		}

		// Check volume weighted average calculation
		// Expected calculation:
		// (50000*1000.5 + 50100*800.25 + 50200*1200.75 + 50300*900.3) / (1000.5 + 800.25 + 1200.75 + 900.3)
		// = 50150.25 (approximately)
		expectedAvg := 50151.28017837921
		actualAvg, parseErr := strconv.ParseFloat(result.VolumeWeightedAvg, 64)
		assert.Nil(t, parseErr)

		// Allow some tolerance for floating point precision
		tolerance := 0.0000001
		if abs(actualAvg-expectedAvg) > tolerance {
			t.Errorf("Expected volume weighted average around %f, got %f", expectedAvg, actualAvg)
		}

		// Check total volume
		expectedTotalVolume := 1000.5 + 800.25 + 1200.75 + 900.3
		actualTotalVolume, parseErr := strconv.ParseFloat(result.TotalVolume, 64)
		assert.Nil(t, parseErr)
		if abs(actualTotalVolume-expectedTotalVolume) > tolerance {
			t.Errorf("Expected total volume around %f, got %f", expectedTotalVolume, actualTotalVolume)
		}

		// Check timestamp is recent
		now := time.Now().Unix()
		assert.True(t, result.Timestamp >= now-60 && result.Timestamp <= now+60)
	})
}

// TestPriceFeedClient_ErrorScenarios tests error handling
func TestPriceFeedClient_ErrorScenarios(t *testing.T) {
	// Create mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return different error responses
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v3/ticker/24hr"):
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Internal server error"}`))
		case strings.HasPrefix(r.URL.Path, "/v5/market/tickers"):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Not found"}`))
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	customConfigs := map[string]configs.ExchangeConfig{
		"binance": {
			Name:    "Binance",
			BaseURL: server.URL,
			Symbols: map[string][]string{
				"BTC": {"BTCUSDT"},
			},
		},
	}

	customSymbols := map[string][]string{
		"BTC": {"binance"},
	}

	client := &PriceFeedClient{
		exchangeConfigs: customConfigs,
		coinExchanges:   customSymbols,
	}

	t.Run("TestInvalidExchange", func(t *testing.T) {
		result, err := client.FetchPriceFromExchange(context.Background(), "invalid_exchange", "BTC", "BTCUSDT")
		assert.Error(t, err)
		assert.Equal(t, appErrors.ErrExchangeNotConfigured.Code, err.Code)
		assert.Nil(t, result)
	})

	t.Run("TestInvalidSymbol", func(t *testing.T) {
		result, err := client.FetchPriceFromExchange(context.Background(), "binance", "BTC", "INVALID")
		assert.Error(t, err)
		assert.Equal(t, appErrors.ErrExchangeInvalidStatusCode.Code, err.Code)
		assert.Nil(t, result)
	})

	t.Run("TestInsufficientData", func(t *testing.T) {
		// Test with only one exchange that fails
		result, err := client.GetPriceFeed(context.Background(), "BTC")
		assert.Error(t, err)
		assert.Equal(t, appErrors.ErrInsufficientExchangeData.Code, err.Code)
		assert.Nil(t, result)
	})
}

// TestCalculateVolumeWeightedAverage tests the volume weighted average calculation
func TestCalculateVolumeWeightedAverage(t *testing.T) {
	tests := []struct {
		name           string
		prices         []ExchangePrice
		expectedAvg    float64
		expectedVolume float64
		expectedCount  int
	}{
		{
			name: "Normal case",
			prices: []ExchangePrice{
				{Exchange: "Binance", Price: 50000.0, Volume: 1000.0, Symbol: "BTC"},
				{Exchange: "Bybit", Price: 50100.0, Volume: 800.0, Symbol: "BTC"},
			},
			expectedAvg:    50044.44, // (50000*1000 + 50100*800) / (1000 + 800)
			expectedVolume: 1800.0,
			expectedCount:  2,
		},
		{
			name:           "Empty prices",
			prices:         []ExchangePrice{},
			expectedAvg:    0.0,
			expectedVolume: 0.0,
			expectedCount:  0,
		},
		{
			name: "Zero volume prices",
			prices: []ExchangePrice{
				{Exchange: "Binance", Price: 50000.0, Volume: 0.0, Symbol: "BTC"},
				{Exchange: "Bybit", Price: 50100.0, Volume: 800.0, Symbol: "BTC"},
			},
			expectedAvg:    50100.0, // Only Bybit contributes
			expectedVolume: 800.0,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avg, volume, count := CalculateVolumeWeightedAverage(tt.prices)

			tolerance := 0.01
			assert.InDelta(t, tt.expectedAvg, avg, tolerance)
			assert.InDelta(t, tt.expectedVolume, volume, tolerance)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

// Helper function to calculate absolute difference
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
