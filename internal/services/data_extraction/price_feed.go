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
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/metrics"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
)

// ExchangePrice represents a price from a single exchange
type ExchangePrice struct {
	Exchange string  `json:"exchange"` // Exchange name.
	Price    float64 `json:"price"`    // Price.
	Volume   float64 `json:"volume"`   // Volume.
	Symbol   string  `json:"symbol"`   // Symbol.
}

// PriceFeedResult represents the result of a price feed calculation
type PriceFeedResult struct {
	Symbol            string          `json:"symbol"` // Symbol.
	VolumeWeightedAvg string          `json:"volumeWeightedAvg"` // Volume-weighted average price.
	TotalVolume       string          `json:"totalVolume"` // Total volume.
	ExchangeCount     int             `json:"exchangeCount"` // Number of exchanges.
	Timestamp         int64           `json:"timestamp"` // Timestamp.
	ExchangePrices    []ExchangePrice `json:"exchangePrices"` // Exchange prices.
	Success           bool            `json:"success"` // Success.
}

// PriceFeedClient is the client for the price feed.
type PriceFeedClient struct {
	exchangeConfigs configs.ExchangesConfig // Exchange configurations.
	symbolExchanges configs.SymbolExchanges // Symbol exchanges.
}

// NewPriceFeedClient creates a new PriceFeedClient with default configurations
func NewPriceFeedClient() *PriceFeedClient {
	exchangeConfigs := configs.GetExchangesConfigs() // Get exchange configurations.
	symbolExchanges := configs.GetSymbolExchanges() // Get symbol exchanges.

	return &PriceFeedClient{
		exchangeConfigs: exchangeConfigs,
		symbolExchanges: symbolExchanges,
	}
}

// FetchPriceFromExchange fetches price and volume data from a specific exchange.
//
// This function performs the following steps sequentially:
// 	1. Retrieves the exchange configuration for the given exchangeKey.
// 	2. Retrieves the endpoint for the given symbol from the exchange configuration.
// 	3. Constructs the full URL for the API request, handling cases where the BaseURL may or may not include the protocol.
// 	4. Creates a retryable HTTP client for robust network requests.
// 	5. Builds an HTTP GET request with the provided context.
// 	6. Executes the HTTP request and handles any network errors.
// 	7. Checks the HTTP response status code for success.
// 	8. Reads the response body from the exchange API.
// 	9. Attempts to decode the response body as a JSON object. If decoding fails and the exchange is "gate.io", attempts to decode as a JSON array and adapts the data structure accordingly.
// 	10. Parses the price and volume from the decoded response using the appropriate exchange-specific parser.
// 	11. Returns an ExchangePrice struct with the parsed data, or an error if any step fails.
//	
// Parameters:
//   - ctx: The context for request cancellation and logging.
//   - exchangeKey: The key identifying the exchange (e.g., "binance").
//   - symbol: The trading symbol (e.g., "BTC").
//
// Returns:
//   - *ExchangePrice: The parsed price and volume data from the exchange.
//   - *appErrors.AppError: An application error if any step fails, otherwise nil.
func (c *PriceFeedClient) FetchPriceFromExchange(ctx context.Context, exchangeKey, symbol string) (*ExchangePrice, *appErrors.AppError) {
	reqLogger := logger.FromContext(ctx)

	// Step 1: Get exchange configuration.
	config, exists := c.exchangeConfigs[exchangeKey]
	if !exists {
		reqLogger.Error("Exchange not configured", "exchange", exchangeKey)
		return nil, appErrors.NewAppError(appErrors.ErrExchangeNotConfigured)
	}

	// Step 2: Get endpoint for the symbol.
	endpoint, exists := config.Endpoints[symbol]
	if !exists {
		reqLogger.Error("Symbol not supported by exchange", "symbol", symbol, "exchange", exchangeKey)
		return nil, appErrors.NewAppError(appErrors.ErrSymbolNotSupportedByExchange)
	}

	// Step 3: Construct the full URL, handling protocol presence.
	var url string
	if strings.HasPrefix(config.BaseURL, "http://") || strings.HasPrefix(config.BaseURL, "https://") {
		url = fmt.Sprintf("%s%s", config.BaseURL, endpoint)
	} else {
		url = fmt.Sprintf("https://%s%s", config.BaseURL, endpoint)
	}

	// Step 4: Create retryable HTTP client.
	httpClient := utils.GetRetryableHTTPClient(1)

	// Step 5: Create request with context.
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		reqLogger.Error("Error creating HTTP request", "error", err, "exchange", exchangeKey, "symbol", symbol)
		return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeFetchFailed, err.Error())
	}

	// Step 6: Execute the HTTP request.
	resp, err := httpClient.Do(req)
	if err != nil {
		reqLogger.Error("Error fetching price from exchange", "error", err, "exchange", exchangeKey, "symbol", symbol)
		return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeFetchFailed, err.Error())
	}
	defer resp.Body.Close()

	// Step 7: Check for valid HTTP status code.
	if resp.StatusCode != http.StatusOK {
		reqLogger.Error("Invalid status code", "status_code", resp.StatusCode, "exchange", exchangeKey, "symbol", symbol)
		return nil, appErrors.NewAppErrorWithResponseStatus(appErrors.ErrExchangeInvalidStatusCode, resp.StatusCode)
	}

	// Step 8: Read the response body.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		reqLogger.Error("Error reading response body", "error", err, "exchange", exchangeKey, "symbol", symbol)
		return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeResponseDecodeFailed, err.Error())
	}

	// Step 9: Attempt to decode the response as a JSON object.
	var data map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		// If object decoding fails and it's Gate.io, try array decoding.
		if exchangeKey == "gate.io" {
			var arrayData []interface{}
			if err := json.Unmarshal(bodyBytes, &arrayData); err != nil {
				reqLogger.Error("Error decoding response body", "error", err, "exchange", exchangeKey, "symbol", symbol)
				return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeResponseDecodeFailed, err.Error())
			}
			// Convert array to expected format for Gate.io parsing.
			data = map[string]interface{}{
				"": arrayData,
			}
		} else {
			reqLogger.Error("Error decoding response body", "error", err, "exchange", exchangeKey, "symbol", symbol)
			return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeResponseDecodeFailed, err.Error())
		}
	}

	// Step 10: Parse price and volume from the decoded response.
	price, volume, parseErr := c.parseExchangeResponse(exchangeKey, data)
	if parseErr != nil {
		reqLogger.Error("Error parsing exchange response", "error", parseErr, "exchange", exchangeKey, "symbol", symbol)
		return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrExchangeResponseParseFailed, parseErr.Error())
	}

	// Step 11: Return the parsed ExchangePrice.
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

	// XT API uses "c" for close price and "q" for base volume
	priceStr, ok := item["c"].(string)
	if !ok {
		logger.Error("Invalid price format: ", "c", item["c"])
		return 0, 0, appErrors.NewAppError(appErrors.ErrInvalidPriceFormat)
	}

	volumeStr, ok := item["q"].(string)
	if !ok {
		logger.Error("Invalid volume format: ", "q", item["q"])
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
			metrics.RecordExchangeApiError(result.exchange, strconv.Itoa(int(result.err.Code)))
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

	metrics.RecordPriceFeedExchangeCount(symbol, exchangeCount)

	// Ensure at least 2 exchanges responded successfully
	if exchangeCount < 2 {
		metrics.RecordError("insufficient_exchange_data", "price_feed")
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

	// Check if the encoding option is valid.
	if attestationRequest.EncodingOptions.Value != "float" {
		reqLogger.Error("Invalid encoding option", "encodingOption", attestationRequest.EncodingOptions.Value)
		return ExtractDataResult{
			StatusCode: http.StatusBadRequest,
		}, appErrors.NewAppError(appErrors.ErrInvalidEncodingOption)
	}

	// Extract symbol from the price feed URL
	var symbol string
	switch attestationRequest.Url {
	case constants.PRICE_FEED_BTC_URL:
		symbol = "BTC"
	case constants.PRICE_FEED_ETH_URL:
		symbol = "ETH"
	case constants.PRICE_FEED_ALEO_URL:
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
