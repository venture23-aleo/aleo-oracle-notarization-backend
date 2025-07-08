package data_extraction

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
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
	Symbol            string          `json:"symbol"`
	VolumeWeightedAvg string          `json:"volumeWeightedAvg"`
	TotalVolume       string          `json:"totalVolume"`
	ExchangeCount     int             `json:"exchangeCount"`
	Timestamp         int64           `json:"timestamp"`
	ExchangePrices    []ExchangePrice `json:"exchangePrices"`
	Success           bool            `json:"success"`
}

type PriceFeedClient struct {
	exchangeConfigs configs.ExchangesConfig
	symbolExchanges configs.SymbolExchanges
}

// NewPriceFeedClient creates a new PriceFeedClient with default configurations
func NewPriceFeedClient() *PriceFeedClient {
	exchangeConfigs := configs.GetExchangesConfigs()
	symbolExchanges := configs.GetSymbolExchanges()

	return &PriceFeedClient{
		exchangeConfigs: exchangeConfigs,
		symbolExchanges: symbolExchanges,
	}
}

// FetchPriceFromExchange fetches price and volume data from a specific exchange
func (c *PriceFeedClient) FetchPriceFromExchange(ctx context.Context, exchangeKey, symbol string) (*ExchangePrice, *appErrors.AppError) {
	reqLogger := logger.FromContext(ctx)
	config, exists := c.exchangeConfigs[exchangeKey]
	if !exists {
		reqLogger.Error("Exchange not configured", "exchange", exchangeKey)
		return nil, appErrors.NewAppError(appErrors.ErrExchangeNotConfigured)
	}

	endpoint, exists := config.Endpoints[symbol]
	if !exists {
		reqLogger.Error("Symbol not supported by exchange", "symbol", symbol, "exchange", exchangeKey)
		return nil, appErrors.NewAppError(appErrors.ErrSymbolNotSupportedByExchange)
	}

	// Handle BaseURL that might already include protocol
	var url string
	if strings.HasPrefix(config.BaseURL, "http://") || strings.HasPrefix(config.BaseURL, "https://") {
		url = fmt.Sprintf("%s%s", config.BaseURL, endpoint)
	} else {
		url = fmt.Sprintf("https://%s%s", config.BaseURL, endpoint)
	}

	httpClient := utils.GetRetryableHTTPClient(1)

	// Create request with context
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		reqLogger.Error("Error creating HTTP request", "error", err, "exchange", exchangeKey, "symbol", symbol)
		return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeFetchFailed, err.Error())
	}

	resp, err := httpClient.Do(req)

	if err != nil {
		reqLogger.Error("Error fetching price from exchange", "error", err, "exchange", exchangeKey, "symbol", symbol)
		return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeFetchFailed, err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		reqLogger.Error("Invalid status code", "status_code", resp.StatusCode, "exchange", exchangeKey, "symbol", symbol)
		return nil, appErrors.NewAppErrorWithResponseStatus(appErrors.ErrExchangeInvalidStatusCode, resp.StatusCode)
	}

	// Read the response body once
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		reqLogger.Error("Error reading response body", "error", err, "exchange", exchangeKey, "symbol", symbol)
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
				reqLogger.Error("Error decoding response body", "error", err, "exchange", exchangeKey, "symbol", symbol)
				return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeResponseDecodeFailed, err.Error())
			}

			// Convert array to expected format for Gate.io parsing
			data = map[string]interface{}{
				"": arrayData,
			}
		} else {
			reqLogger.Error("Error decoding response body", "error", err, "exchange", exchangeKey, "symbol", symbol)
			return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeResponseDecodeFailed, err.Error())
		}
	}

	price, volume, parseErr := c.parseExchangeResponse(exchangeKey, data)
	if parseErr != nil {
		reqLogger.Error("Error parsing exchange response", "error", parseErr, "exchange", exchangeKey, "symbol", symbol)
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
		logger.Error("Unsupported exchange: ", "exchangeKey", exchangeKey)
		return 0, 0, appErrors.NewAppError(appErrors.ErrUnsupportedExchange)
	}
}

// Parse functions for each exchange
func parseBinanceResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	priceStr, ok := data["lastPrice"].(string)
	if !ok {
		logger.Error("Invalid price format: ", "lastPrice", data["lastPrice"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := data["volume"].(string)
	if !ok {
		logger.Error("Invalid volume format: ", "volume", data["volume"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing price: ", "error", parseErr)
		return 0, 0, appErrors.NewAppError(appErrors.ErrPriceParseFailed)
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing volume: ", "error", parseErr)
		return 0, 0, appErrors.NewAppError(appErrors.ErrVolumeParseFailed)
	}

	return price, volume, nil
}

func parseBybitResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	result, ok := data["result"].(map[string]interface{})
	if !ok {
		logger.Error("Invalid exchange response format: ", "result", data["result"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidExchangeResponseFormat)
	}

	list, ok := result["list"].([]interface{})
	if !ok || len(list) == 0 {
		logger.Error("No data in response")
		return 0, 0, appErrors.NewAppError(appErrors.ErrNoDataInResponse)
	}

	item, ok := list[0].(map[string]interface{})
	if !ok {
		logger.Error("Invalid item format: ", "item", list[0])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidItemFormat)
	}

	priceStr, ok := item["lastPrice"].(string)
	if !ok {
		logger.Error("Invalid price format: ", "lastPrice", item["lastPrice"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := item["volume24h"].(string)
	if !ok {
		logger.Error("Invalid volume format: ", "volume24h", item["volume24h"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing price: ", "error", parseErr)
		return 0, 0, appErrors.NewAppError(appErrors.ErrPriceParseFailed)
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing volume: ", "error", parseErr)
		return 0, 0, appErrors.NewAppError(appErrors.ErrVolumeParseFailed)
	}

	return price, volume, nil
}

func parseCoinbaseResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	priceStr, ok := data["price"].(string)
	if !ok {
		logger.Error("Invalid price format: ", "price", data["price"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := data["volume"].(string)
	if !ok {
		logger.Error("Invalid volume format: ", "volume", data["volume"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing price: ", "error", parseErr)
		return 0, 0, appErrors.NewAppError(appErrors.ErrPriceParseFailed)
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing volume: ", "error", parseErr)
		return 0, 0, appErrors.NewAppError(appErrors.ErrVolumeParseFailed)
	}

	return price, volume, nil
}

func parseCryptoComResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	result, ok := data["result"].(map[string]interface{})
	if !ok {
		logger.Error("Invalid exchange response format: ", "result", data["result"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidExchangeResponseFormat)
	}

	dataArray, ok := result["data"].([]interface{})
	if !ok || len(dataArray) == 0 {
		logger.Error("No data in response")
		return 0, 0, appErrors.NewAppError(appErrors.ErrNoDataInResponse)
	}

	dataMap, ok := dataArray[0].(map[string]interface{})
	if !ok {
		logger.Error("Invalid data format: ", "dataArray", dataArray[0])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidDataFormat)
	}

	// Crypto.com uses "k" for last price and "v" for volume
	priceStr, ok := dataMap["k"].(string)
	if !ok {
		logger.Error("Invalid price format: ", "k", dataMap["k"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := dataMap["v"].(string)
	if !ok {
		logger.Error("Invalid volume format: ", "v", dataMap["v"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing price: ", "error", parseErr)
		return 0, 0, appErrors.NewAppError(appErrors.ErrPriceParseFailed)
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing volume: ", "error", parseErr)
		return 0, 0, appErrors.NewAppError(appErrors.ErrVolumeParseFailed)
	}

	return price, volume, nil
}

func parseXTResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	result, ok := data["result"].([]interface{})
	if !ok || len(result) == 0 {
		logger.Error("No data in response")
		return 0, 0, appErrors.NewAppError(appErrors.ErrNoDataInResponse)
	}

	item, ok := result[0].(map[string]interface{})
	if !ok {
		logger.Error("Invalid item format: ", "result", result[0])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidItemFormat)
	}

	// XT API uses "c" for close price and "v" for volume
	priceStr, ok := item["c"].(string)
	if !ok {
		logger.Error("Invalid price format: ", "c", item["c"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := item["v"].(string)
	if !ok {
		logger.Error("Invalid volume format: ", "v", item["v"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing price: ", "error", parseErr)
		return 0, 0, appErrors.NewAppError(appErrors.ErrPriceParseFailed)
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing volume: ", "error", parseErr)
		return 0, 0, appErrors.NewAppError(appErrors.ErrVolumeParseFailed)
	}

	return price, volume, nil
}

func parseGateIOResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	list, ok := data[""].([]interface{})
	if !ok || len(list) == 0 {
		logger.Error("No data in response")
		return 0, 0, appErrors.NewAppError(appErrors.ErrNoDataInResponse)
	}

	item, ok := list[0].(map[string]interface{})
	if !ok {
		logger.Error("Invalid item format: ", "list", list[0])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidItemFormat)
	}

	priceStr, ok := item["last"].(string)
	if !ok {
		logger.Error("Invalid price format: ", "last", item["last"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := item["base_volume"].(string)
	if !ok {
		logger.Error("Invalid volume format: ", "base_volume", item["base_volume"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing price: ", "error", parseErr)
		return 0, 0, appErrors.NewAppErrorWithDetails(appErrors.ErrPriceParseFailed, parseErr.Error())
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing volume: ", "error", parseErr)
		return 0, 0, appErrors.NewAppErrorWithDetails(appErrors.ErrVolumeParseFailed, parseErr.Error())
	}

	return price, volume, nil
}

func parseMEXCResponse(data map[string]interface{}) (price, volume float64, err *appErrors.AppError) {
	priceStr, ok := data["lastPrice"].(string)
	if !ok {
		logger.Error("Invalid price format: ", "lastPrice", data["lastPrice"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := data["volume"].(string)
	if !ok {
		logger.Error("Invalid volume format: ", "volume", data["volume"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidVolumeFormat)
	}

	price, parseErr := strconv.ParseFloat(priceStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing price: ", "error", parseErr)
		return 0, 0, appErrors.NewAppError(appErrors.ErrPriceParseFailed)
	}

	volume, parseErr = strconv.ParseFloat(volumeStr, 64)
	if parseErr != nil {
		logger.Error("Error parsing volume: ", "error", parseErr)
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
	price    *ExchangePrice
	err      *appErrors.AppError
}

// GetPriceFeed fetches and calculates the volume-weighted average price for a given symbol
func (c *PriceFeedClient) GetPriceFeed(ctx context.Context, symbol string) (*PriceFeedResult, *appErrors.AppError) {
	reqLogger := logger.FromContext(ctx)
	exchanges, exists := c.symbolExchanges[strings.ToUpper(symbol)]
	if !exists {
		reqLogger.Error("Invalid symbol", "symbol", symbol)
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
			price, err := c.FetchPriceFromExchange(ctx, ex, strings.ToUpper(symbol))
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
			reqLogger.Error("Failed to fetch token price", "exchange", result.exchange, "error", result.err.Details)
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
		reqLogger.Error("Insufficient exchange data", "exchangeCount", exchangeCount)
		return nil, appErrors.NewAppError(appErrors.ErrInsufficientExchangeData)
	}

	return &PriceFeedResult{
		Symbol:            strings.ToUpper(symbol),
		VolumeWeightedAvg: strconv.FormatFloat(volumeWeightedAvg, 'f', -1, 64),
		TotalVolume:       strconv.FormatFloat(totalVolume, 'f', -1, 64),
		ExchangeCount:     exchangeCount,
		Timestamp:         time.Now().Unix(),
		ExchangePrices:    exchangePrices,
		Success:           true,
	}, nil
}

// ExtractPriceFeedData handles price feed requests and always returns the volume-weighted average price (VWAP)
// This ensures consistent and reliable price data for oracle attestations
func ExtractPriceFeedData(ctx context.Context, attestationRequest attestation.AttestationRequest) (ExtractDataResult, *appErrors.AppError) {
	reqLogger := logger.FromContext(ctx)
	if attestationRequest.EncodingOptions.Value != "float" {
		reqLogger.Error("Invalid encoding option", "encodingOption", attestationRequest.EncodingOptions.Value)
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
		logger.Error("Unsupported price feed URL: ", "url", attestationRequest.Url)
		return ExtractDataResult{
			StatusCode: http.StatusBadRequest,
		}, appErrors.NewAppError(appErrors.ErrUnsupportedPriceFeedURL)
	}

	priceFeedClient := NewPriceFeedClient()

	if priceFeedClient == nil {
		reqLogger.Error("Failed to create price feed client")
		return ExtractDataResult{
			StatusCode: http.StatusInternalServerError,
		}, appErrors.NewAppError(appErrors.ErrExchangeFetchFailed)
	}

	// Get the price feed data
	result, appErr := priceFeedClient.GetPriceFeed(ctx, symbol)

	if appErr != nil {
		reqLogger.Error("Error getting price feed for ", "symbol", symbol, "error", appErr)
		return ExtractDataResult{
			StatusCode: http.StatusInternalServerError,
		}, appErr
	}

	// Marshal the response to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		reqLogger.Error("Error marshalling price feed data", "error", err)
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
