package data_extraction

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	configs "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/config"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
)

func TestParseExchangeResponse(t *testing.T) {
	responseTimestamp := time.Now().UnixMilli()
	attestationTimestamp := time.Now().Unix()
	tests := []struct {
		name           string
		exchange       string
		response       []byte
		expectedPrice  float64
		expectedVolume float64
		expectedError  *appErrors.AppError
		symbol         string
		timestamp      int64
	}{
		{
			name:           "Binance valid response",
			exchange:       "binance",
			response:       []byte(fmt.Sprintf(`{"lastPrice": "1000.00", "volume": "1000.00", "symbol": "BTCUSDT", "closeTime": %d}`, responseTimestamp)),
			expectedPrice:  1000.0,
			expectedVolume: 1000.0,
			expectedError:  nil,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Bybit valid response",
			exchange:       "bybit",
			response:       []byte(fmt.Sprintf(`{"time": %d, "result":{"list": [{"lastPrice": "1000.00", "volume24h": "1000.00", "symbol": "BTCUSDT"}]}}`, responseTimestamp)),
			expectedPrice:  1000.0,
			expectedVolume: 1000.0,
			expectedError:  nil,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Gate valid response",
			exchange:       "gate",
			response:       []byte(`[{"last": "1000.00", "base_volume": "1000.00","currency_pair":"BTCUSDT"}]`),
			expectedPrice:  1000.0,
			expectedVolume: 1000.0,
			expectedError:  nil,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "MEXC valid response",
			exchange:       "mexc",
			response:       []byte(fmt.Sprintf(`{"lastPrice": "1000.00", "volume": "1000.00", "symbol": "BTCUSDT", "closeTime": %d}`, responseTimestamp)),
			expectedPrice:  1000.0,
			expectedVolume: 1000.0,
			expectedError:  nil,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "XT valid response",
			exchange:       "xt",
			response:       []byte(fmt.Sprintf(`{"result": [{"c": "1000.00", "q": "1000.00", "s": "BTCUSDT", "t": %d}]}`, responseTimestamp)),
			expectedPrice:  1000.0,
			expectedVolume: 1000.0,
			expectedError:  nil,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid exchange",
			exchange:       "invalid",
			response:       []byte(`{"error": "Invalid response"}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrExchangeNotSupported,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid binance response with invalid price",
			exchange:       "binance",
			response:       []byte(`{"lastPrice": "invalid", "volume": "1000.00"}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingPrice,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid binance response with invalid volume",
			exchange:       "binance",
			response:       []byte(`{"lastPrice": "1000.00", "volume": "invalid"}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingVolume,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid bybit response with invalid price",
			exchange:       "bybit",
			response:       []byte(`{"result":{"list": [{"lastPrice": "invalid", "volume24h": "1000.00"}]}}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingPrice,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid bybit response with invalid volume",
			exchange:       "bybit",
			response:       []byte(`{"result":{"list": [{"lastPrice": "1000.00", "volume24h": "invalid"}]}}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingVolume,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid gate response with invalid price",
			exchange:       "gate",
			response:       []byte(`[{"last": "invalid", "base_volume": "1000.00"}]`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingPrice,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid gate response with invalid volume",
			exchange:       "gate",
			response:       []byte(`[{"last": "1000.00", "base_volume": "invalid"}]`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingVolume,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid mexc response with invalid price",
			exchange:       "mexc",
			response:       []byte(`{"lastPrice": "invalid", "volume": "1000.00"}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingPrice,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid mexc response with invalid volume",
			exchange:       "mexc",
			response:       []byte(`{"lastPrice": "1000.00", "volume": "invalid"}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingVolume,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid xt response with invalid price",
			exchange:       "xt",
			response:       []byte(`{"result": [{"c": "invalid", "q": "1000.00"}]}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingPrice,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid xt response with invalid volume",
			exchange:       "xt",
			response:       []byte(`{"result": [{"c": "1000.00", "q": "invalid"}]}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingVolume,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid coinbase response with invalid price",
			exchange:       "coinbase",
			response:       []byte(`{"price": "invalid", "volume": "1000.00"}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingPrice,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid coinbase response with invalid volume",
			exchange:       "coinbase",
			response:       []byte(`{"price": "1000.00", "volume": "invalid"}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingVolume,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid crypto response with invalid price",
			exchange:       "crypto",
			response:       []byte(`{"result": {"data": [{"k": "invalid", "v": "1000.00"}]}}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingPrice,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid crypto response with invalid volume",
			exchange:       "crypto",
			response:       []byte(`{"result": {"data": [{"k": "1000.00", "v": "invalid"}]}}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrParsingVolume,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Invalid crypto response with empty result",
			exchange:       "crypto",
			response:       []byte(`{}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrMissingDataInResponse,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "Missing data in crypto response",
			exchange:       "crypto",
			response:       []byte(`{"result": {"data": []}}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrMissingDataInResponse,
			symbol:         "BTCUSDT",
			timestamp: attestationTimestamp,
		},
		{
			name:           "Missing data in bybit response",
			exchange:       "bybit",
			response:       []byte(`{"result": {"list": []}}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrMissingDataInResponse,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},

		{
			name:           "missing data in gate response",
			exchange:       "gate",
			response:       []byte(`[]`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrMissingDataInResponse,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "missing data in xt response",
			exchange:       "xt",
			response:       []byte(`{"result": []}`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrMissingDataInResponse,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "malformed json response",
			exchange:       "binance",
			response:       []byte(`test`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrDecodingExchangeResponse,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "malformed json response",
			exchange:       "bybit",
			response:       []byte(`test`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrDecodingExchangeResponse,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "malformed json response",
			exchange:       "gate",
			response:       []byte(`test`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrDecodingExchangeResponse,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "malformed json response",
			exchange:       "mexc",
			response:       []byte(`test`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrDecodingExchangeResponse,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "malformed json response",
			exchange:       "xt",
			response:       []byte(`test`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrDecodingExchangeResponse,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "malformed json response",
			exchange:       "crypto",
			response:       []byte(`test`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrDecodingExchangeResponse,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
		{
			name:           "malformed json response",
			exchange:       "coinbase",
			response:       []byte(`test`),
			expectedPrice:  0.0,
			expectedVolume: 0.0,
			expectedError:  appErrors.ErrDecodingExchangeResponse,
			symbol:         "BTCUSDT",
			timestamp:      attestationTimestamp,
		},
	}

	exchangesConfigs := configs.GetExchangesConfigs()
	tokenExchanges := configs.GetTokenExchanges()
	tokenTradingPairs := configs.GetTokenTradingPairs()

	priceFeedClient := &PriceFeedClient{
		exchangeConfigs:   exchangesConfigs,
		tokenExchanges:    tokenExchanges,
		tokenTradingPairs: tokenTradingPairs,
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.exchange, tt.name), func(t *testing.T) {
			price, volume, err := priceFeedClient.parseExchangeResponse(tt.exchange, tt.response, tt.symbol, time.Now().Unix(), "USDT")
			assert.Equal(t, tt.expectedPrice, price)
			assert.Equal(t, tt.expectedVolume, volume)
			assert.Equal(t, tt.expectedError, err)
		})
	}
}
