package data_extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	configs "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/config"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
)

var (
	priceFeedServer                        *httptest.Server
	priceFeedServerWith404Error            *httptest.Server
	priceFeedServerWithInternalServerError *httptest.Server
)

// TestMain initializes the logger for all tests in this package
func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger.InitLogger("DEBUG")

	priceFeedServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Return different responses based on the exchange endpoint
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v3/ticker/24hr"):
			response := BinanceResponse{}
			if r.URL.Query().Get("symbol") == "BTCUSDT" {
				response.Price = "50000.00"
				response.Volume = "1000.50"
			} else if r.URL.Query().Get("symbol") == "BTCUSDC" {
				response.Price = "50100.00"
				response.Volume = "800.25"
			} else if r.URL.Query().Get("symbol") == "ETHUSDT" {
				response.Price = "3990.00"
				response.Volume = "1000.00"
			} else if r.URL.Query().Get("symbol") == "ETHUSDC" {
				response.Price = "3900.00"
				response.Volume = "1000.00"
			} else if r.URL.Query().Get("symbol") == "ALEOUSDT" {
				response := MEXCResponse{
					Price:  "0.24",
					Volume: "1000.00",
				}
				json.NewEncoder(w).Encode(response)
				return
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Symbol not found"))
				return
			}
			json.NewEncoder(w).Encode(response)

		case strings.HasPrefix(r.URL.Path, "/v5/market/tickers"):
			response := BybitResponse{}
			if r.URL.Query().Get("symbol") == "BTCUSDT" {
				response.Result.List = append(response.Result.List, BybitListItem{
					Price:  "50100.00",
					Volume: "800.25",
				})
			} else if r.URL.Query().Get("symbol") == "BTCUSDC" {
				response.Result.List = append(response.Result.List, BybitListItem{
					Price:  "50100.00",
					Volume: "800.25",
				})
			} else if r.URL.Query().Get("symbol") == "ETHUSDT" {
				response.Result.List = append(response.Result.List, BybitListItem{
					Price:  "3990.00",
					Volume: "1500.00",
				})
			} else if r.URL.Query().Get("symbol") == "ETHUSDC" {
				response.Result.List = append(response.Result.List, BybitListItem{
					Price:  "4000.00",
					Volume: "2000.00",
				})
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Symbol not found"))
				return
			}

			json.NewEncoder(w).Encode(response)

		case strings.HasPrefix(r.URL.Path, "/products/"):
			symbol := strings.Split(r.URL.Path, "/")[2]
			response := CoinbaseResponse{}
			switch symbol {
			case "BTC-USD":
				response.Price = "50200.00"
				response.Volume = "1200.75"
			case "BTC-USDT":
				response.Price = "50200.00"
				response.Volume = "1200.75"
			case "ETH-USD":
				response.Price = "4000.00"
				response.Volume = "1500.00"
			case "ETH-USDT":
				response.Price = "4000.00"
				response.Volume = "2000.00"
			case "ALEO-USD":
				response.Price = "0.24"
				response.Volume = "1000.00"
			default:
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Symbol not found"))
				return
			}
			json.NewEncoder(w).Encode(response)

		case strings.HasPrefix(r.URL.Path, "/v2/public/get-ticker"):
			response := CryptoResponse{}
			instrumentName := r.URL.Query().Get("instrument_name")
			switch instrumentName {
			case "BTC_USDT":
				response.Result.Data = append(response.Result.Data, CryptoListItem{
					Price:  "50300.00",
					Volume: "900.30",
				})
			case "BTC_USD":
				response.Result.Data = append(response.Result.Data, CryptoListItem{
					Price:  "50300.00",
					Volume: "900.30",
				})
			case "ETH_USDT":
				response.Result.Data = append(response.Result.Data, CryptoListItem{
					Price:  "4000.00",
					Volume: "2000.00",
				})
			case "ETH_USD":
				response.Result.Data = append(response.Result.Data, CryptoListItem{
					Price:  "4000.00",
					Volume: "2000.00",
				})
			default:
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Symbol not found"))
				return
			}
			json.NewEncoder(w).Encode(response)

		case strings.HasPrefix(r.URL.Path, "/api/v4/spot/tickers"):
			response := GateResponse{}
			currencyPair := r.URL.Query().Get("currency_pair")
			if currencyPair == "ALEO_USDT" {
				response = append(response, GateResponseItem{
					Price:  "0.24",
					Volume: "1000.00",
				})
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Symbol not found"))
				return
			}
			json.NewEncoder(w).Encode(response)
		case strings.HasPrefix(r.URL.Path, "/api/v3/ticker/24hr"):
			response := MEXCResponse{}
			if r.URL.Query().Get("symbol") == "ALEOUSDT" {
				response.Price = "0.24"
				response.Volume = "1000.00"
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Symbol not found"))
				return
			}
			json.NewEncoder(w).Encode(response)
		case strings.HasPrefix(r.URL.Path, "/sapi/v4/market/public/ticker/24h"):
			response := XTResponse{}
			if r.URL.Query().Get("symbol") == "ALEO_USDT" {
				response.Result = append(response.Result, XTResponseItem{
					Price:  "0.24",
					Volume: "1000.00",
				})
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Symbol not found"))
				return
			}
			json.NewEncoder(w).Encode(response)
		default:
			// Return 404 for unknown endpoints
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	defer priceFeedServer.Close()

	priceFeedServerWithInternalServerError = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal server error"}`))
	}))

	priceFeedServerWith404Error = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Not found"}`))
	}))
	defer priceFeedServerWith404Error.Close()

	defer priceFeedServerWithInternalServerError.Close()

	// Run the tests
	m.Run()
}

func TestPriceFeed_AllValidExchangeResponse(t *testing.T) {
	exchangesConfigs := configs.GetExchangesConfigs()
	tokenExchanges := configs.GetTokenExchanges()
	tokenTradingPairs := configs.GetTokenTradingPairs()

	newExchangesConfigs := make(configs.ExchangesConfig)
	for _, exchange := range exchangesConfigs {
		exchange.BaseURL = priceFeedServer.URL
		newExchangesConfigs[strings.ToLower(exchange.Name)] = exchange
	}

	priceFeedClient := &PriceFeedClient{
		exchangeConfigs:   newExchangesConfigs,
		tokenExchanges:    tokenExchanges,
		tokenTradingPairs: tokenTradingPairs,
	}

	for token, exchanges := range tokenExchanges {
		price, err := priceFeedClient.GetPriceFeed(context.Background(), token, 12)
		assert.Nil(t, err)
		assert.NotNil(t, price)
		assert.Equal(t, token, price.Token)
		assert.True(t, price.Success)
		assert.Equal(t, len(exchanges), price.ExchangeCount)
		assert.Equal(t, len(tokenTradingPairs[token]), len(price.ExchangePricesRaw))
	}
}

func TestPriceFeed_WithInternalServerError(t *testing.T) {
	exchangesConfigs := configs.GetExchangesConfigs()
	tokenExchanges := configs.GetTokenExchanges()
	tokenTradingPairs := configs.GetTokenTradingPairs()

	newExchangesConfigs := make(configs.ExchangesConfig)
	for _, exchange := range exchangesConfigs {
		exchange.BaseURL = priceFeedServerWithInternalServerError.URL
		newExchangesConfigs[strings.ToLower(exchange.Name)] = exchange
	}

	priceFeedClient := &PriceFeedClient{
		exchangeConfigs:   newExchangesConfigs,
		tokenExchanges:    tokenExchanges,
		tokenTradingPairs: tokenTradingPairs,
	}

	for token := range tokenExchanges {
		price, err := priceFeedClient.GetPriceFeed(context.Background(), token, 12)
		assert.NotNil(t, err)
		assert.Nil(t, price)
	}
}

func TestPriceFeed_With404Error(t *testing.T) {
	exchangesConfigs := configs.GetExchangesConfigs()
	tokenExchanges := configs.GetTokenExchanges()
	tokenTradingPairs := configs.GetTokenTradingPairs()

	newExchangesConfigs := make(configs.ExchangesConfig)
	for _, exchange := range exchangesConfigs {
		exchange.BaseURL = priceFeedServerWith404Error.URL
		newExchangesConfigs[strings.ToLower(exchange.Name)] = exchange
	}

	priceFeedClient := &PriceFeedClient{
		exchangeConfigs:   newExchangesConfigs,
		tokenExchanges:    tokenExchanges,
		tokenTradingPairs: tokenTradingPairs,
	}

	for token := range tokenExchanges {
		price, err := priceFeedClient.GetPriceFeed(context.Background(), token, 12)
		assert.NotNil(t, err)
		assert.Nil(t, price)
	}
}
func TestPriceFeed_PartialValidExchangeResponse(t *testing.T) {

	// Create mock HTTP server that returns realistic exchange responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Return different responses based on the exchange endpoint
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v3/ticker/24hr"):
			response := BinanceResponse{}
			if r.URL.Query().Get("symbol") == "BTCUSDT" {
				response.Price = "price"
				response.Volume = "1000.50"
			} else if r.URL.Query().Get("symbol") == "BTCUSDC" {
				response.Price = "50100.00"
				response.Volume = "volume"
			} else if r.URL.Query().Get("symbol") == "ETHUSDT" {
				response.Price = "3990.00"
				response.Volume = "1000.00"
			} else if r.URL.Query().Get("symbol") == "ETHUSDC" {
				response.Price = "test"
				response.Volume = "1000.00"
			} else if r.URL.Query().Get("symbol") == "ALEOUSDT" {
				response := MEXCResponse{
					Price:  "test",
					Volume: "1000.00",
				}
				json.NewEncoder(w).Encode(response)
				return
			}
			json.NewEncoder(w).Encode(response)

		case strings.HasPrefix(r.URL.Path, "/v5/market/tickers"):
			response := BybitResponse{}
			symbol := r.URL.Query().Get("symbol")
			switch symbol {
			case "BTCUSDT":
				response.Result.List = append(response.Result.List, BybitListItem{
					Price:  "50100.00",
					Volume: "volume",
				})
			case "BTCUSDC":
				response.Result.List = append(response.Result.List, BybitListItem{
					Price:  "price",
					Volume: "800.25",
				})
			case "ETHUSDT":
				response.Result.List = append(response.Result.List, BybitListItem{
					Price:  "test",
					Volume: "1500.00",
				})
			case "ETHUSDC":
				response.Result.List = append(response.Result.List, BybitListItem{
					Price:  "4000.00",
					Volume: "volume",
				})
			}
			json.NewEncoder(w).Encode(response)

		case strings.HasPrefix(r.URL.Path, "/products/"):
			symbol := strings.Split(r.URL.Path, "/")[2]
			response := CoinbaseResponse{}
			switch symbol {
			case "BTC-USD":
				response.Price = "price"
				response.Volume = "1200.75"
			case "BTC-USDT":
				response.Price = "50200.00"
				response.Volume = "volume"
			case "ETH-USD":
				response.Price = "test"
				response.Volume = "1500.00"
			case "ETH-USDT":
				response.Price = "4000.00"
				response.Volume = "volume"
			case "ALEO-USD":
				response.Price = "0.24"
				response.Volume = "price"
			}
			json.NewEncoder(w).Encode(response)

		case strings.HasPrefix(r.URL.Path, "/v2/public/get-ticker"):
			response := CryptoResponse{}
			instrumentName := r.URL.Query().Get("instrument_name")
			switch instrumentName {
			case "BTC_USDT":
				response.Result.Data = append(response.Result.Data, CryptoListItem{
					Price:  "50300.00",
					Volume: "900.30",
				})
			case "BTC_USD":
				response.Result.Data = append(response.Result.Data, CryptoListItem{
					Price:  "50300.00",
					Volume: "900.30",
				})
			case "ETH_USDT":
				response.Result.Data = append(response.Result.Data, CryptoListItem{
					Price:  "4000.00",
					Volume: "volume",
				})
			case "ETH_USD":
				response.Result.Data = append(response.Result.Data, CryptoListItem{
					Price:  "price",
					Volume: "2000.00",
				})
			}
			json.NewEncoder(w).Encode(response)

		case strings.HasPrefix(r.URL.Path, "/api/v4/spot/tickers"):
			response := GateResponse{}
			if r.URL.Query().Get("currency_pair") == "ALEO_USDT" {
				response = append(response, GateResponseItem{
					Price:  "price",
					Volume: "1000.00",
				})
			}
			json.NewEncoder(w).Encode(response)
		case strings.HasPrefix(r.URL.Path, "/api/v3/ticker/24hr"):
			response := MEXCResponse{}
			if r.URL.Query().Get("symbol") == "ALEOUSDT" {
				response.Price = "0.24"
				response.Volume = "volume"
			}
			json.NewEncoder(w).Encode(response)
		case strings.HasPrefix(r.URL.Path, "/sapi/v4/market/public/ticker/24h"):
			response := XTResponse{}
			if r.URL.Query().Get("symbol") == "ALEO_USDT" {
				response.Result = append(response.Result, XTResponseItem{
					Price:  "0.24",
					Volume: "volume",
				})
			}
			json.NewEncoder(w).Encode(response)
		default:
			// Return 404 for unknown endpoints
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	exchangesConfigs := configs.GetExchangesConfigs()
	tokenExchanges := configs.GetTokenExchanges()
	tokenTradingPairs := configs.GetTokenTradingPairs()

	newExchangesConfigs := make(configs.ExchangesConfig)
	for _, exchange := range exchangesConfigs {
		exchange.BaseURL = server.URL
		newExchangesConfigs[strings.ToLower(exchange.Name)] = exchange
	}

	priceFeedClient := &PriceFeedClient{
		exchangeConfigs:   newExchangesConfigs,
		tokenExchanges:    tokenExchanges,
		tokenTradingPairs: tokenTradingPairs,
	}

	for token := range tokenExchanges {
		price, err := priceFeedClient.GetPriceFeed(context.Background(), token, 12)
		assert.NotNil(t, err)
		assert.Nil(t, price)
	}
}

// TestPriceFeedClient_ErrorScenarios tests error handling
func TestGetPriceFeed_ErrorScenarios(t *testing.T) {
	// Create mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return different error responses
		switch {
		case strings.HasPrefix(r.URL.Path, "/binance"):
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "Internal server error"}`))
		case strings.HasPrefix(r.URL.Path, "/bybit"):
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Not found"}`))
		case strings.HasPrefix(r.URL.Path, "/mexc"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"price": "1000.00", "volume": "volume"}`))
		case strings.HasPrefix(r.URL.Path, "/gate"):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{"last": "1000.00", "base_volume": "volume"}]`))
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	// customConfigs := map[string]configs.ExchangeConfig{
	// 	"binance": {
	// 		Name:    "Binance",
	// 		BaseURL: server.URL,
	// 		Symbols: map[string][]string{
	// 			"BTC": {"BTCUSDT"},
	// 		},
	// 	},
	// }

	// customTokens := map[string][]string{
	// 	"BTC": {"binances"},
	// }

	// tokenTradingPairs := make(configs.TokenTradingPairs)
	// tokenTradingPairs["BTC"] = []string{"BTCUSDT"}

	testCases := []struct {
		name           string
		token          string
		testToken      string
		tokenExchanges []string
		symbols        []string
		testSymbols    []string
		exchanges      []string
		exchangeTokens []string
		expectedError  *appErrors.AppError
	}{
		{
			name:           "TestTokenNotSupported",
			token:          "BTC",
			testToken:      "INVALID",
			tokenExchanges: []string{},
			symbols:        []string{"BTCUSDT"},
			testSymbols:    []string{"BTCUSDT"},
			exchanges:      []string{"binance"},
			exchangeTokens: []string{"BTC"},
			expectedError:  appErrors.ErrTokenNotSupported,
		},
		{
			name:           "TestNoExchangesConfigured",
			token:          "BTC",
			testToken:      "BTC",
			tokenExchanges: []string{},
			symbols:        []string{"BTCUSDT"},
			testSymbols:    []string{"BTCUSDT"},
			exchanges:      []string{"binance"},
			exchangeTokens: []string{"BTC"},
			expectedError:  appErrors.ErrExchangeNotConfigured,
		},
		{
			name:           "TestNoTradingPairsConfigured",
			token:          "BTC",
			testToken:      "BTC",
			tokenExchanges: []string{"binance"},
			symbols:        []string{},
			testSymbols:    []string{},
			exchanges:      []string{"binance"},
			exchangeTokens: []string{"BTC"},
			expectedError:  appErrors.ErrNoTradingPairsConfigured,
		},
		{
			name:           "TestInvalidExchange",
			token:          "BTC",
			testToken:      "BTC",
			tokenExchanges: []string{"binance"},
			symbols:        []string{"BTCUSDT"},
			testSymbols:    []string{"BTCUSDT"},
			exchanges:      []string{"invalid_exchange"},
			exchangeTokens: []string{"BTC"},
			expectedError:  appErrors.ErrExchangeNotConfigured,
		},
		{
			name:           "TestTokenNotSupported",
			token:          "BTC",
			testToken:      "BTC",
			tokenExchanges: []string{"binance"},
			testSymbols:    []string{},
			symbols:        []string{"BTCUSDT", "BTCUSDC"},
			exchanges:      []string{"binance"},
			exchangeTokens: []string{"BTCS"},
			expectedError:  appErrors.ErrTokenNotSupported,
		},
		{
			name:           "TestSymbolNotConfigured",
			token:          "BTC",
			testToken:      "BTC",
			tokenExchanges: []string{"binance"},
			testSymbols:    []string{},
			symbols:        []string{"BTCUSDT", "BTCUSDC"},
			exchanges:      []string{"binance"},
			exchangeTokens: []string{"BTC"},
			expectedError:  appErrors.ErrSymbolNotConfigured,
		},
		{
			name:           "TestInsufficientExchangeData",
			token:          "BTC",
			testToken:      "BTC",
			tokenExchanges: []string{"binance"},
			symbols:        []string{"BTCUSDT"},
			testSymbols:    []string{"BTCUSDT"},
			exchanges:      []string{"binance"},
			exchangeTokens: []string{"BTC"},
			expectedError:  appErrors.ErrInsufficientExchangeData,
		},
		{
			name:           "TestMEXCInvalidResponse",
			token:          "ALEO",
			testToken:      "ALEO",
			tokenExchanges: []string{"mexc"},
			symbols:        []string{"ALEOUSDT"},
			testSymbols:    []string{"ALEOUSDT"},
			exchanges:      []string{"mexc"},
			exchangeTokens: []string{"ALEO"},
			expectedError:  appErrors.ErrInsufficientExchangeData,
		},
		{
			name:           "TestGateInvalidResponse",
			token:          "ALEO",
			testToken:      "ALEO",
			tokenExchanges: []string{"gate"},
			symbols:        []string{"ALEOUSDT"},
			testSymbols:    []string{"ALEOUSDT"},
			exchanges:      []string{"gate"},
			exchangeTokens: []string{"ALEO"},
			expectedError:  appErrors.ErrInsufficientExchangeData,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			exchangeConfig := make(map[string]configs.ExchangeConfig)
			symbols := make(map[string][]string)

			for _, exchangeToken := range testCase.exchangeTokens {
				symbols[exchangeToken] = testCase.testSymbols
			}

			for _, exchange := range testCase.exchanges {
				exchangeConfig[exchange] = configs.ExchangeConfig{
					Name:             exchange,
					BaseURL:          server.URL,
					EndpointTemplate: "/" + exchange,
					Symbols:          symbols,
				}
			}

			tokenExchanges := make(map[string][]string)
			tokenExchanges[testCase.token] = testCase.tokenExchanges

			tokenTradingPairs := make(configs.TokenTradingPairs)
			tokenTradingPairs[testCase.token] = testCase.symbols

			client := &PriceFeedClient{
				exchangeConfigs:   exchangeConfig,
				tokenExchanges:    tokenExchanges,
				tokenTradingPairs: tokenTradingPairs,
			}
			result, err := client.GetPriceFeed(context.Background(), testCase.testToken, 12)
			assert.Error(t, err)
			assert.Equal(t, testCase.expectedError, err)
			assert.Nil(t, result)
		})
	}
}

// TestCalculateVolumeWeightedAverage tests the volume weighted average calculation
func TestCalculateVolumeWeightedAveragePrice(t *testing.T) {
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
				{Exchange: "Binance", Price: "50000.0", Volume: "1000.0", Symbol: "BTC"},
				{Exchange: "Bybit", Price: "50100.0", Volume: "800.0", Symbol: "BTC"},
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
			name: "Partial zero volume",
			prices: []ExchangePrice{
				{Exchange: "Binance", Price: "50000.0", Volume: "0.0", Symbol: "BTC"},
				{Exchange: "Bybit", Price: "50100.0", Volume: "800", Symbol: "BTC"},
			},
			expectedAvg:    50100.0, // Only Bybit contributes
			expectedVolume: 800.0,
			expectedCount:  1,
		},
		{
			name: "All zero volume",
			prices: []ExchangePrice{
				{Exchange: "Binance", Price: "50000.0", Volume: "0.0", Symbol: "BTC"},
				{Exchange: "Bybit", Price: "50100.0", Volume: "0", Symbol: "BTC"},
			},
			expectedAvg:    0.0,
			expectedVolume: 0.0,
			expectedCount:  0,
		},
		{
			name: "All exchanges",
			prices: []ExchangePrice{
				{Exchange: "Binance", Price: "50000.0", Volume: "1000.5", Symbol: "BTC"},
				{Exchange: "Bybit", Price: "50100.0", Volume: "800.25", Symbol: "BTC"},
				{Exchange: "Coinbase", Price: "50200.0", Volume: "1200.75", Symbol: "BTC"},
				{Exchange: "Crypto.com", Price: "50300.0", Volume: "900.3", Symbol: "BTC"},
			},
			expectedAvg:    50151.2801737921,
			expectedVolume: 1000.5 + 800.25 + 1200.75 + 900.3,
			expectedCount:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avg, volume, count, _, _ := CalculateVolumeWeightedAverage(tt.prices, 3, "BTC")

			tolerance := 0.01
			assert.InDelta(t, tt.expectedAvg, avg, tolerance)
			assert.InDelta(t, tt.expectedVolume, volume, tolerance)
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

func TestNewPriceFeedClient(t *testing.T) {
	priceFeedClient := NewPriceFeedClient()
	assert.NotNil(t, priceFeedClient.exchangeConfigs)
	assert.NotNil(t, priceFeedClient.tokenExchanges)
	assert.IsType(t, configs.TokenTradingPairs{}, priceFeedClient.tokenTradingPairs)
	assert.IsType(t, configs.TokenExchanges{}, priceFeedClient.tokenExchanges)
	assert.IsType(t, configs.ExchangesConfig{}, priceFeedClient.exchangeConfigs)
}

func TestExtractPriceFeedData(t *testing.T) {

	testCases := []struct {
		name                    string
		token                   string
		baseUrl                 string
		expectedAttestationData string
		expectedError           *appErrors.AppError
	}{
		{
			name:                    "Valid price feed",
			token:                   "BTC",
			baseUrl:                 priceFeedServer.URL,
			expectedAttestationData: "50165.7894217680",
			expectedError:           nil,
		},
		{
			name:                    "Invalid price feed",
			token:                   "BTC",
			baseUrl:                 priceFeedServerWith404Error.URL,
			expectedError:           appErrors.ErrInsufficientExchangeData,
			expectedAttestationData: "",
		},
	}

	exchangesConfigs := configs.GetExchangesConfigs()
	tokenExchanges := configs.GetTokenExchanges()
	tokenTradingPairs := configs.GetTokenTradingPairs()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			newExchangesConfigs := make(configs.ExchangesConfig)
			for _, exchange := range exchangesConfigs {
				exchange.BaseURL = testCase.baseUrl
				newExchangesConfigs[strings.ToLower(exchange.Name)] = exchange
			}

			priceFeedClient := &PriceFeedClient{
				exchangeConfigs:   newExchangesConfigs,
				tokenExchanges:    tokenExchanges,
				tokenTradingPairs: tokenTradingPairs,
			}

			attestationRequest := attestation.AttestationRequest{
				Url:            fmt.Sprintf("price_feed: %s", testCase.token),
				RequestMethod:  "GET",
				ResponseFormat: "json",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 10,
				},
			}

			result, err := priceFeedClient.ExtractPriceFeedData(context.Background(), attestationRequest, "BTC")
			assert.Equal(t, testCase.expectedError, err)
			assert.Equal(t, testCase.expectedAttestationData, result.AttestationData)
		})
	}
}

func TestFetchPriceFromExchange(t *testing.T) {
	malformedJsonServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}))
	defer malformedJsonServer.Close()

	exchangesConfigs := configs.GetExchangesConfigs()
	tokenExchanges := configs.GetTokenExchanges()
	tokenTradingPairs := configs.GetTokenTradingPairs()

	testCases := []struct {
		name          string
		exchange      string
		token         string
		symbol        string
		baseUrl       string
		expectedError *appErrors.AppError
	}{
		{
			name:          "Valid exchange",
			exchange:      "binance",
			token:         "BTC",
			symbol:        "BTCUSDT",
			baseUrl:       priceFeedServer.URL,
			expectedError: nil,
		},
		{
			name:          "Invalid exchange",
			exchange:      "invalid_exchange",
			token:         "BTC",
			symbol:        "BTCUSDC",
			baseUrl:       priceFeedServer.URL,
			expectedError: appErrors.ErrExchangeNotConfigured,
		},
		{
			name:          "Invalid symbol with working price feed server",
			exchange:      "binance",
			token:         "BTC",
			symbol:        "INVALID",
			baseUrl:       priceFeedServer.URL,
			expectedError: appErrors.ErrExchangeInvalidStatusCode,
		},
		{
			name:          "Valid symbol with internal server error price feed server",
			exchange:      "binance",
			token:         "BTC",
			symbol:        "BTCUSDT",
			baseUrl:       priceFeedServerWithInternalServerError.URL,
			expectedError: appErrors.ErrFetchingFromExchange,
		},
		{
			name:          "Valid symbol with malformed json price feed server",
			exchange:      "binance",
			token:         "BTC",
			symbol:        "BTCUSDT",
			baseUrl:       malformedJsonServer.URL,
			expectedError: appErrors.ErrDecodingExchangeResponse,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			newExchangesConfigs := make(configs.ExchangesConfig)
			for _, exchange := range exchangesConfigs {
				exchange.BaseURL = testCase.baseUrl
				newExchangesConfigs[strings.ToLower(exchange.Name)] = exchange
			}

			priceFeedClient := &PriceFeedClient{
				exchangeConfigs:   newExchangesConfigs,
				tokenExchanges:    tokenExchanges,
				tokenTradingPairs: tokenTradingPairs,
			}
			_, err := priceFeedClient.FetchPriceFromExchange(context.Background(), testCase.exchange, testCase.token, testCase.symbol)
			assert.Equal(t, testCase.expectedError, err)
		})
	}

}
