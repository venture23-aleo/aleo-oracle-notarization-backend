package data_extraction

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
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

type PriceFeedClient struct {
	exchangeConfigs configs.ExchangeConfigs
	symbolExchanges configs.SymbolExchanges
}

// NewPriceFeedClient creates a new PriceFeedClient with default configurations
func NewPriceFeedClient() *PriceFeedClient {
	return &PriceFeedClient{
		exchangeConfigs: configs.GetExchangeConfigs(),
		symbolExchanges: configs.GetSymbolExchanges(),
	}
}

// FetchPriceFromExchange fetches price and volume data from a specific exchange
func (c *PriceFeedClient) FetchPriceFromExchange(exchangeKey, symbol string) (*ExchangePrice, *appErrors.AppError) {
	config, exists := c.exchangeConfigs[exchangeKey]
	if !exists {
		log.Printf("[ERROR] [FetchPriceFromExchange] Exchange not configured: %s", exchangeKey)
		return nil, appErrors.NewAppError(appErrors.ErrExchangeNotConfigured)
	}

	endpoint, exists := config.Endpoints[symbol]
	if !exists {
		log.Printf("[ERROR] [FetchPriceFromExchange] Symbol not supported by exchange: %s", symbol)
		return nil, appErrors.NewAppError(appErrors.ErrSymbolNotSupportedByExchange)
	}

	// Handle BaseURL that might already include protocol
	var url string
	if strings.HasPrefix(config.BaseURL, "http://") || strings.HasPrefix(config.BaseURL, "https://") {
		url = fmt.Sprintf("%s%s", config.BaseURL, endpoint)
	} else {
		url = fmt.Sprintf("https://%s%s", config.BaseURL, endpoint)
	}

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := httpClient.Get(url)

	if err != nil {
		log.Printf("[ERROR] [FetchPriceFromExchange] Error fetching price from exchange: %s", err.Error())
		return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeFetchFailed, err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[ERROR] [FetchPriceFromExchange] Invalid status code: %d", resp.StatusCode)
		return nil, appErrors.NewAppErrorWithResponseStatus(appErrors.ErrExchangeInvalidStatusCode, resp.StatusCode)
	}

	// Read the response body once
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] [FetchPriceFromExchange] Error reading response body: %s", err.Error())
		return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeResponseDecodeFailed, err.Error())
	}

	// Handle different response types (object vs array)
	var data map[string]interface{}
	
	// Try to decode as object first
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		// If object decoding fails and it's Gate.io, try array decoding
		if exchangeKey == "gate.io" {
			var arrayData []interface{}
			if err := json.Unmarshal(bodyBytes, &arrayData); err != nil {
				return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeResponseDecodeFailed, err.Error())
			}
			
			// Convert array to expected format for Gate.io parsing
			data = map[string]interface{}{
				"": arrayData,
			}
		} else {
			log.Printf("[ERROR] [FetchPriceFromExchange] Error decoding response body: %s", err.Error())
			return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeResponseDecodeFailed, err.Error())
		}
	}


	price, volume, parseErr := c.parseExchangeResponse(exchangeKey, data)
	if parseErr != nil {
		log.Printf("[ERROR] [FetchPriceFromExchange] Error parsing exchange response: %s", parseErr.Error())
		return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeResponseParseFailed, parseErr.Error())
	}

	return &ExchangePrice{
		Exchange: config.Name,
		Price:    price,
		Volume:   volume,
		Symbol:   symbol,
	}, nil
}

// parseExchangeResponse parses the response from different exchanges
func (c *PriceFeedClient) parseExchangeResponse(exchangeKey string, data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
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
		log.Printf("[ERROR] [parseExchangeResponse] Unsupported exchange: %s", exchangeKey)
		return 0, 0, appErrors.NewAppError(appErrors.ErrUnsupportedExchange)
	}
}

// Parse functions for each exchange
func parseBinanceResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	priceStr, ok := data["lastPrice"].(string)
	if !ok {
		log.Printf("[ERROR] [parseBinanceResponse] Invalid price format: %s", data["lastPrice"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := data["volume"].(string)
	if !ok {
		log.Printf("[ERROR] [parseBinanceResponse] Invalid volume format: %s", data["volume"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseBinanceResponse] Error parsing price: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppError(appErrors.ErrPriceParseFailed)
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseBinanceResponse] Error parsing volume: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppError(appErrors.ErrVolumeParseFailed)
	}

	return price, volume, nil
}

func parseBybitResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	result, ok := data["result"].(map[string]interface{})
	if !ok {
		log.Printf("[ERROR] [parseBybitResponse] Invalid exchange response format: %s", data["result"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidExchangeResponseFormat)
	}

	list, ok := result["list"].([]interface{})
	if !ok || len(list) == 0 {
		log.Printf("[ERROR] [parseBybitResponse] No data in response")
		return 0, 0, appErrors.NewAppError(appErrors.ErrNoDataInResponse)
	}

	item, ok := list[0].(map[string]interface{})
	if !ok {
		log.Printf("[ERROR] [parseBybitResponse] Invalid item format: %s", list[0])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidItemFormat)
	}

	priceStr, ok := item["lastPrice"].(string)
	if !ok {
		log.Printf("[ERROR] [parseBybitResponse] Invalid price format: %s", item["lastPrice"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := item["volume24h"].(string)
	if !ok {
		log.Printf("[ERROR] [parseBybitResponse] Invalid volume format: %s", item["volume24h"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseBybitResponse] Error parsing price: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppError(appErrors.ErrPriceParseFailed)
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseBybitResponse] Error parsing volume: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppError(appErrors.ErrVolumeParseFailed)
	}

	return price, volume, nil
}

func parseCoinbaseResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	priceStr, ok := data["price"].(string)
	if !ok {
		log.Printf("[ERROR] [parseCoinbaseResponse] Invalid price format: %s", data["price"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := data["volume"].(string)
	if !ok {
		log.Printf("[ERROR] [parseCoinbaseResponse] Invalid volume format: %s", data["volume"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseCoinbaseResponse] Error parsing price: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppError(appErrors.ErrPriceParseFailed)
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseCoinbaseResponse] Error parsing volume: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppError(appErrors.ErrVolumeParseFailed)
	}

	return price, volume, nil
}

func parseCryptoComResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	result, ok := data["result"].(map[string]interface{})
	if !ok {
		log.Printf("[ERROR] [parseCryptoComResponse] Invalid exchange response format: %s", data["result"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidExchangeResponseFormat)
	}

	dataArray, ok := result["data"].([]interface{})
	if !ok || len(dataArray) == 0 {
		log.Printf("[ERROR] [parseCryptoComResponse] No data in response")
		return 0, 0, appErrors.NewAppError(appErrors.ErrNoDataInResponse)
	}

	dataMap, ok := dataArray[0].(map[string]interface{})
	if !ok {
		log.Printf("[ERROR] [parseCryptoComResponse] Invalid data format: %s", dataArray[0])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidDataFormat)
	}

	// Crypto.com uses "k" for last price and "v" for volume
	priceStr, ok := dataMap["k"].(string)
	if !ok {
		log.Printf("[ERROR] [parseCryptoComResponse] Invalid price format: %s", dataMap["k"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := dataMap["v"].(string)
	if !ok {
		log.Printf("[ERROR] [parseCryptoComResponse] Invalid volume format: %s", dataMap["v"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseCryptoComResponse] Error parsing price: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppError(appErrors.ErrPriceParseFailed)
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseCryptoComResponse] Error parsing volume: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppError(appErrors.ErrVolumeParseFailed)
	}

	return price, volume, nil
}

func parseXTResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	result, ok := data["result"].([]interface{})
	if !ok || len(result) == 0 {
		log.Printf("[ERROR] [parseXTResponse] No data in response")
		return 0, 0, appErrors.NewAppError(appErrors.ErrNoDataInResponse)
	}

	item, ok := result[0].(map[string]interface{})
	if !ok {
		log.Printf("[ERROR] [parseXTResponse] Invalid item format: %s", result[0])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidItemFormat)
	}

	// XT API uses "c" for close price and "v" for volume
	priceStr, ok := item["c"].(string)
	if !ok {
		log.Printf("[ERROR] [parseXTResponse] Invalid price format: %s", item["c"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := item["v"].(string)
	if !ok {
		log.Printf("[ERROR] [parseXTResponse] Invalid volume format: %s", item["v"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseXTResponse] Error parsing price: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppError(appErrors.ErrPriceParseFailed)
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseXTResponse] Error parsing volume: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppError(appErrors.ErrVolumeParseFailed)
	}

	return price, volume, nil
}

func parseGateIOResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	list, ok := data[""].([]interface{})
	if !ok || len(list) == 0 {
		log.Printf("[ERROR] [parseGateIOResponse] No data in response")
		return 0, 0, appErrors.NewAppError(appErrors.ErrNoDataInResponse)
	}

	item, ok := list[0].(map[string]interface{})
	if !ok {
		log.Printf("[ERROR] [parseGateIOResponse] Invalid item format: %s", list[0])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidItemFormat)
	}

	priceStr, ok := item["last"].(string)
	if !ok {
		log.Printf("[ERROR] [parseGateIOResponse] Invalid price format: %s", item["last"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := item["quote_volume"].(string)
	if !ok {
		log.Printf("[ERROR] [parseGateIOResponse] Invalid volume format: %s", item["quote_volume"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseGateIOResponse] Error parsing price: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppErrorWithDetails(appErrors.ErrPriceParseFailed, parseErr.Error())
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseGateIOResponse] Error parsing volume: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppErrorWithDetails(appErrors.ErrVolumeParseFailed, parseErr.Error())
	}

	return price, volume, nil
}

func parseMEXCResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	priceStr, ok := data["lastPrice"].(string)
	if !ok {
		log.Printf("[ERROR] [parseMEXCResponse] Invalid price format: %s", data["lastPrice"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := data["volume"].(string)
	if !ok {
		log.Printf("[ERROR] [parseMEXCResponse] Invalid volume format: %s", data["volume"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseMEXCResponse] Error parsing price: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppError(appErrors.ErrPriceParseFailed)
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		log.Printf("[ERROR] [parseMEXCResponse] Error parsing volume: %s", parseErr.Error())
		return 0, 0, appErrors.NewAppError(appErrors.ErrVolumeParseFailed)
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

// Fetch prices from all exchanges concurrently
type fetchResult struct {
	exchange string
	price *ExchangePrice
	err   *appErrors.AppError
}

// GetPriceFeed fetches and calculates the volume-weighted average price for a given symbol
func (c *PriceFeedClient) GetPriceFeed(symbol string) (*PriceFeedResult, *appErrors.AppError) {
	exchanges, exists := c.symbolExchanges[strings.ToUpper(symbol)]
	if !exists {
		log.Printf("[ERROR] [GetPriceFeed] Invalid symbol: %s", symbol)
		return nil, appErrors.NewAppError(appErrors.ErrInvalidSymbol)
	}

	var exchangePrices []ExchangePrice

	// Create a buffered channel to collect results from goroutines
	// Buffer size matches the number of exchanges to prevent blocking
	results := make(chan fetchResult, len(exchanges))

	// Launch concurrent goroutines to fetch prices from each exchange
	// Each goroutine fetches data independently and sends results through the channel
	for _, exchange := range exchanges {
		go func(ex string) {
			price, err := c.FetchPriceFromExchange(ex, strings.ToUpper(symbol))
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
			log.Printf("[ERROR] [GetPriceFeed] Failed to fetch from %s: %v",result.exchange, result.err.Details)
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
		log.Printf("[ERROR] [GetPriceFeed] Insufficient exchange data: %d", exchangeCount)
		return nil, appErrors.NewAppError(appErrors.ErrInsufficientExchangeData)
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


// ExtractPriceFeedData handles price feed requests and always returns the volume-weighted average price (VWAP)
// This ensures consistent and reliable price data for oracle attestations
func ExtractPriceFeedData(attestationRequest attestation.AttestationRequest) (ExtractDataResult, *appErrors.AppError) {
	if attestationRequest.EncodingOptions.Value != "float" {
		log.Printf("[ERROR] [ExtractPriceFeedData] Invalid encoding option: %s", attestationRequest.EncodingOptions.Value)
		return ExtractDataResult{
			StatusCode: http.StatusBadRequest,
		}, appErrors.NewAppError(appErrors.ErrInvalidEncodingOption)
	}

	// Extract symbol from the price feed URL
	var symbol string
	switch attestationRequest.Url {
	case constants.PriceFeedBtcUrl:
		symbol = "BTC"
	case constants.PriceFeedEthUrl:
		symbol = "ETH"
	case constants.PriceFeedAleoUrl:
		symbol = "ALEO"
	default:
		log.Printf("[ERROR] [ExtractPriceFeedData] Unsupported price feed URL: %s", attestationRequest.Url)
		return ExtractDataResult{
			StatusCode: http.StatusBadRequest,
		}, appErrors.NewAppError(appErrors.ErrUnsupportedPriceFeedURL)
	}

	priceFeedClient := NewPriceFeedClient()

	// Get the price feed data
	result, appErr := priceFeedClient.GetPriceFeed(symbol)

	if appErr != nil {
		log.Printf("[ERROR] [ExtractPriceFeedData] Error getting price feed for %s: %v", symbol, appErr)
		return ExtractDataResult{
			StatusCode: http.StatusInternalServerError,
		}, appErr
	}

	// Marshal the response to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		log.Printf("[ERROR] [ExtractPriceFeedData] Error marshalling price feed data: %v", err)
		return ExtractDataResult{
			StatusCode: http.StatusInternalServerError,
		}, appErrors.NewAppError(appErrors.ErrJSONEncoding)
	}

	// Extract the value based on the selector
	// For price feeds, always use weightedAvgPrice (volume-weighted average price)
	var valueStr string = result.VolumeWeightedAvg

	valueStr = applyFloatPrecision(valueStr, attestationRequest.EncodingOptions.Precision)

	return ExtractDataResult{
		ResponseBody:    string(jsonBytes),
		AttestationData: valueStr,
		StatusCode:      http.StatusOK,
	}, nil
}