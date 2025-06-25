package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
)

// ExchangePrice represents a price from a single exchange
type ExchangePrice struct {
	Exchange string  `json:"exchange"`
	Price    float64 `json:"price"`
	Volume   float64 `json:"volume"`
	Symbol   string  `json:"symbol"`
}

// PriceFeedResult represents the result of a price feed calculation
type PriceFeedResult struct {
	Symbol            string           `json:"symbol"`
	VolumeWeightedAvg string          `json:"volumeWeightedAvg"`
	TotalVolume       string          `json:"totalVolume"`
	ExchangeCount     int              `json:"exchangeCount"`
	Timestamp         int64            `json:"timestamp"`
	ExchangePrices    []ExchangePrice  `json:"exchangePrices"`
	Success           bool             `json:"success"`
}

// ExchangeConfig defines the configuration for each exchange
type ExchangeConfig struct {
	Name     string
	BaseURL  string
	Symbols  map[string]string // Maps our symbol to exchange symbol
	Endpoints map[string]string // Maps symbol to endpoint path
}

// Binance API correct
// Bybit API correct
// Coinbase API correct
// Crypto.com API correct
// XT API correct
// Gate.io API correct
// MEXC API correct

// Exchange configurations for each supported exchange
var exchangeConfigs = map[string]ExchangeConfig{
	"binance": {
		Name: "Binance",
		BaseURL: "api.binance.com",
		Symbols: map[string]string{
			"BTC": "BTCUSDT",
			"ETH": "ETHUSDT",
		},
		Endpoints: map[string]string{
			"BTC": "/api/v3/ticker/24hr?symbol=BTCUSDT",
			"ETH": "/api/v3/ticker/24hr?symbol=ETHUSDT",
		},
	},
	"bybit": {
		Name: "Bybit",
		BaseURL: "api.bybit.com",
		Symbols: map[string]string{
			"BTC": "BTCUSDT",
			"ETH": "ETHUSDT",
		},
		Endpoints: map[string]string{
			"BTC": "/v5/market/tickers?category=spot&symbol=BTCUSDT",
			"ETH": "/v5/market/tickers?category=spot&symbol=ETHUSDT",
		},
	},
	"coinbase": {
		Name: "Coinbase",
		BaseURL: "api.exchange.coinbase.com",
		Symbols: map[string]string{
			"BTC": "BTC-USD",
			"ETH": "ETH-USD",
			"ALEO": "ALEO-USD",
		},
		Endpoints: map[string]string{
			"BTC": "/products/BTC-USD/ticker",
			"ETH": "/products/ETH-USD/ticker",
			"ALEO": "/products/ALEO-USD/ticker",
		},
	},
	"crypto.com": {
		Name: "Crypto.com",
		BaseURL: "api.crypto.com",
		Symbols: map[string]string{
			"BTC": "BTC_USDT",
			"ETH": "ETH_USDT",
		},
		Endpoints: map[string]string{
			"BTC": "/v2/public/get-ticker?instrument_name=BTC_USDT",
			"ETH": "/v2/public/get-ticker?instrument_name=ETH_USDT",
		},
	},
	"xt": {
		Name: "XT",
		BaseURL: "xt.com",
		Symbols: map[string]string{
			"ALEO": "ALEO_USDT",
		},
		Endpoints: map[string]string{
			"ALEO": "/sapi/v4/market/public/ticker/24h?symbol=ALEO_USDT",
		},
	},
	"gate.io": {
		Name: "Gate.io",
		BaseURL: "api.gateio.ws",
		Symbols: map[string]string{
			"ALEO": "ALEO_USDT",
		},
		Endpoints: map[string]string{
			"ALEO": "/api/v4/spot/tickers?currency_pair=ALEO_USDT",
		},
	},
	"mexc": {
		Name: "MEXC",
		BaseURL: "api.mexc.com",
		Symbols: map[string]string{
			"ALEO": "ALEOUSDT",
		},
		Endpoints: map[string]string{
			"ALEO": "/api/v3/ticker/24hr?symbol=ALEOUSDT",
		},
	},
}

// Symbol to exchanges mapping
var symbolExchanges = map[string][]string{
	"BTC":  {"binance", "bybit", "coinbase", "crypto.com"},
	"ETH":  {"binance", "bybit", "coinbase", "crypto.com"},
	"ALEO": {"xt", "gate.io", "coinbase", "mexc"},
}

// FetchPriceFromExchange fetches price and volume data from a specific exchange
func FetchPriceFromExchange(exchangeKey, symbol string) (*ExchangePrice, error) {
	config, exists := exchangeConfigs[exchangeKey]
	if !exists {
		return nil, appErrors.ErrExchangeNotConfigured
	}

	endpoint, exists := config.Endpoints[symbol]
	if !exists {
		return nil, appErrors.ErrSymbolNotSupportedByExchange
	}

	url := fmt.Sprintf("https://%s%s", config.BaseURL, endpoint)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", appErrors.ErrExchangeFetchFailed, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", appErrors.ErrExchangeInvalidStatusCode, resp.StatusCode)
	}

	// Read the response body once
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", appErrors.ErrExchangeResponseDecodeFailed, err.Error())
	}

	// Handle different response types (object vs array)
	var data map[string]interface{}
	
	// Try to decode as object first
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		// If object decoding fails and it's Gate.io, try array decoding
		if exchangeKey == "gate.io" {
			var arrayData []interface{}
			if err := json.Unmarshal(bodyBytes, &arrayData); err != nil {
				return nil, fmt.Errorf("%w: %s", appErrors.ErrExchangeResponseDecodeFailed, err.Error())
			}
			
			// Convert array to expected format for Gate.io parsing
			data = map[string]interface{}{
				"": arrayData,
			}
		} else {
			return nil, fmt.Errorf("%w: %s", appErrors.ErrExchangeResponseDecodeFailed, err.Error())
		}
	}

	if exchangeKey == "crypto.com" {
		log.Printf("data: %v", data)
	}

	price, volume, err := parseExchangeResponse(exchangeKey, data)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", appErrors.ErrExchangeResponseParseFailed, err.Error())
	}

	return &ExchangePrice{
		Exchange: config.Name,
		Price:    price,
		Volume:   volume,
		Symbol:   symbol,
	}, nil
}

// parseExchangeResponse parses the response from different exchanges
func parseExchangeResponse(exchangeKey string, data map[string]interface{}) (price, volume float64, err error) {
	switch exchangeKey {
	case "binance":
		return parseBinanceResponse(data)
	case "bybit":
		return parseBybitResponse(data)
	case "coinbase":
		return parseCoinbaseResponse(data)
	case "crypto.com":
		return parseCryptoComResponse(data)
	case "xt":
		return parseXTResponse(data)
	case "gate.io":
		return parseGateIOResponse(data)
	case "mexc":
		return parseMEXCResponse(data)
	default:
		return 0, 0, appErrors.ErrUnsupportedExchange
	}
}

// Parse functions for each exchange
func parseBinanceResponse(data map[string]interface{}) (price, volume float64, err error) {
	priceStr, ok := data["lastPrice"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidPriceFormat
	}

	volumeStr, ok := data["volume"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidVolumeFormat
	}

	price, err = strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrPriceParseFailed, err.Error())
	}

	volume, err = strconv.ParseFloat(volumeStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrVolumeParseFailed, err.Error())
	}

	return price, volume, nil
}

func parseBybitResponse(data map[string]interface{}) (price, volume float64, err error) {
	result, ok := data["result"].(map[string]interface{})
	if !ok {
		return 0, 0, appErrors.ErrInvalidExchangeResponseFormat
	}

	list, ok := result["list"].([]interface{})
	if !ok || len(list) == 0 {
		return 0, 0, appErrors.ErrNoDataInResponse
	}

	item, ok := list[0].(map[string]interface{})
	if !ok {
		return 0, 0, appErrors.ErrInvalidItemFormat
	}

	priceStr, ok := item["lastPrice"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidPriceFormat
	}

	volumeStr, ok := item["volume24h"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidVolumeFormat
	}

	price, err = strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrPriceParseFailed, err.Error())
	}

	volume, err = strconv.ParseFloat(volumeStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrVolumeParseFailed, err.Error())
	}

	return price, volume, nil
}

func parseCoinbaseResponse(data map[string]interface{}) (price, volume float64, err error) {
	priceStr, ok := data["price"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidPriceFormat
	}

	volumeStr, ok := data["volume"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidVolumeFormat
	}

	price, err = strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrPriceParseFailed, err.Error())
	}

	volume, err = strconv.ParseFloat(volumeStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrVolumeParseFailed, err.Error())
	}

	return price, volume, nil
}

func parseCryptoComResponse(data map[string]interface{}) (price, volume float64, err error) {
	result, ok := data["result"].(map[string]interface{})
	if !ok {
		return 0, 0, appErrors.ErrInvalidExchangeResponseFormat
	}

	dataArray, ok := result["data"].([]interface{})
	if !ok || len(dataArray) == 0 {
		return 0, 0, appErrors.ErrNoDataInResponse
	}

	dataMap, ok := dataArray[0].(map[string]interface{})
	if !ok {
		return 0, 0, appErrors.ErrInvalidDataFormat
	}

	// Crypto.com uses "k" for last price and "v" for volume
	priceStr, ok := dataMap["k"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidPriceFormat
	}

	volumeStr, ok := dataMap["v"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidVolumeFormat
	}

	price, err = strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrPriceParseFailed, err.Error())
	}

	volume, err = strconv.ParseFloat(volumeStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrVolumeParseFailed, err.Error())
	}

	return price, volume, nil
}

func parseXTResponse(data map[string]interface{}) (price, volume float64, err error) {
	result, ok := data["result"].([]interface{})
	if !ok || len(result) == 0 {
		return 0, 0, appErrors.ErrNoDataInResponse
	}

	item, ok := result[0].(map[string]interface{})
	if !ok {
		return 0, 0, appErrors.ErrInvalidItemFormat
	}

	// XT API uses "c" for close price and "v" for volume
	priceStr, ok := item["c"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidPriceFormat
	}

	volumeStr, ok := item["v"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidVolumeFormat
	}

	price, err = strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrPriceParseFailed, err.Error())
	}

	volume, err = strconv.ParseFloat(volumeStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrVolumeParseFailed, err.Error())
	}

	return price, volume, nil
}

func parseGateIOResponse(data map[string]interface{}) (price, volume float64, err error) {
	list, ok := data[""].([]interface{})
	if !ok || len(list) == 0 {
		return 0, 0, appErrors.ErrNoDataInResponse
	}

	item, ok := list[0].(map[string]interface{})
	if !ok {
		return 0, 0, appErrors.ErrInvalidItemFormat
	}

	priceStr, ok := item["last"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidPriceFormat
	}

	volumeStr, ok := item["quote_volume"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidVolumeFormat
	}

	price, err = strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrPriceParseFailed, err.Error())
	}

	volume, err = strconv.ParseFloat(volumeStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrVolumeParseFailed, err.Error())
	}

	return price, volume, nil
}

func parseMEXCResponse(data map[string]interface{}) (price, volume float64, err error) {
	priceStr, ok := data["lastPrice"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidPriceFormat
	}

	volumeStr, ok := data["volume"].(string)
	if !ok {
		return 0, 0, appErrors.ErrInvalidVolumeFormat
	}

	price, err = strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrPriceParseFailed, err.Error())
	}

	volume, err = strconv.ParseFloat(volumeStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %s", appErrors.ErrVolumeParseFailed, err.Error())
	}

	return price, volume, nil
}

// CalculateVolumeWeightedAverage calculates the volume-weighted average price
func CalculateVolumeWeightedAverage(prices []ExchangePrice) (float64, float64, int) {
	if len(prices) == 0 {
		return 0, 0, 0
	}

	var totalVolume float64
	var weightedSum float64
	validPrices := 0

	for _, price := range prices {
		if price.Price > 0 && price.Volume > 0 {
			weightedSum += price.Price * price.Volume
			totalVolume += price.Volume
			validPrices++
		}
	}

	if totalVolume == 0 {
		return 0, 0, 0
	}

	volumeWeightedAvg := weightedSum / totalVolume
	return volumeWeightedAvg, totalVolume, validPrices
}

// GetPriceFeed fetches and calculates the volume-weighted average price for a given symbol
func GetPriceFeed(symbol string) (*PriceFeedResult, error) {
	exchanges, exists := symbolExchanges[strings.ToUpper(symbol)]
	if !exists {
		return nil, appErrors.ErrInvalidSymbol
	}

	var exchangePrices []ExchangePrice

	// Fetch prices from all exchanges concurrently
	type fetchResult struct {
		exchange string
		price *ExchangePrice
		err   error
	}

	// Create a buffered channel to collect results from goroutines
	// Buffer size matches the number of exchanges to prevent blocking
	results := make(chan fetchResult, len(exchanges))

	// Launch concurrent goroutines to fetch prices from each exchange
	// Each goroutine fetches data independently and sends results through the channel
	for _, exchange := range exchanges {
		go func(ex string) {
			price, err := FetchPriceFromExchange(ex, strings.ToUpper(symbol))
			results <- fetchResult{price: price, err: err, exchange: ex}
		}(exchange)
	}

	// Collect results from all goroutines
	// Process results in the order they complete, not necessarily the order of exchanges
	for i := 0; i < len(exchanges); i++ {
		result := <-results
		if result.err != nil {
			// Log the error but continue processing other exchanges
			// This ensures the system is resilient to individual exchange failures
			log.Printf("Failed to fetch from %s: %v",result.exchange, result.err)
			continue
		}
		if result.price != nil {
			// Only add valid price data to the collection
			exchangePrices = append(exchangePrices, *result.price)
		}
	}

	// Calculate volume-weighted average
	volumeWeightedAvg, totalVolume, exchangeCount := CalculateVolumeWeightedAverage(exchangePrices)

	// Ensure at least 2 exchanges responded successfully
	if exchangeCount < 2 {
		return nil, appErrors.ErrInsufficientExchangeData
	}

	return &PriceFeedResult{
		Symbol:            strings.ToUpper(symbol),
		VolumeWeightedAvg: fmt.Sprintf("%g", volumeWeightedAvg),
		TotalVolume:       fmt.Sprintf("%g", totalVolume),
		ExchangeCount:     exchangeCount,
		Timestamp:         time.Now().Unix(),
		ExchangePrices:    exchangePrices,
		Success:           true,
	}, nil
}

// GetPriceFeedAsString returns the price feed result as a formatted string
func GetPriceFeedAsString(symbol string) (string, error) {
	result, err := GetPriceFeed(symbol)
	if err != nil {
		return "", err
	}

	// Format the price with appropriate precision
	priceStr := result.VolumeWeightedAvg
	
	return priceStr, nil
} 