package data_extraction

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	configs "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/config"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/metrics"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
)

// ExchangePrice represents a price from a single exchange
type ExchangePrice struct {
	Exchange string  `json:"exchange"` // Exchange name.
	Price    string  `json:"price"`    // Price.
	Volume   string  `json:"volume"`   // Volume.
	Token    string  `json:"token"`    // Token.
	Symbol   string  `json:"symbol"`   // Symbol.
}

// PriceFeedResult represents the result of a price feed calculation
type PriceFeedResult struct {
	Token             string          `json:"token"`             // Token.
	VolumeWeightedAvg string          `json:"volumeWeightedAvg"` // Volume-weighted average price.
	TotalVolume       string          `json:"totalVolume"`       // Total volume.
	ExchangeCount     int             `json:"exchangeCount"`     // Number of exchanges.
	Timestamp         int64           `json:"timestamp"`         // Timestamp.
	ExchangePrices    []ExchangePrice `json:"exchangePrices"`    // Exchange prices.
	Success           bool            `json:"success"`           // Success.
}

// Fetch prices from all exchanges concurrently
type fetchResult struct {
	exchange string
	price    *ExchangePrice
	err      *appErrors.AppError
}

// PriceFeedClient is the client for the price feed.
type PriceFeedClient struct {
	exchangeConfigs   configs.ExchangesConfig   // Exchange configurations.
	tokenExchanges    configs.TokenExchanges    // Token exchanges.
	tokenTradingPairs configs.TokenTradingPairs // Token trading pairs.
}

// NewPriceFeedClient creates a new PriceFeedClient with default configurations
func NewPriceFeedClient() *PriceFeedClient {
	exchangeConfigs := configs.GetExchangesConfigs()    // Get exchange configurations.
	tokenExchanges := configs.GetTokenExchanges()       // Get token exchanges.
	tokenTradingPairs := configs.GetTokenTradingPairs() // Get token trading pairs.

	return &PriceFeedClient{
		exchangeConfigs:   exchangeConfigs,
		tokenExchanges:    tokenExchanges,
		tokenTradingPairs: tokenTradingPairs,
	}
}

func GetRetryableHTTPClientForExchange(exchange string, maxRetries int) *retryablehttp.Client {
	// Create a new HTTP client with the TLS configuration
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		InsecureSkipVerify: false,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			if len(verifiedChains) == 0 {
				return fmt.Errorf("no verified chains")
			}

			// Take root of the verified chain
			rootCert := verifiedChains[0][len(verifiedChains[0])-1]

			rootCAFile := fmt.Sprintf("/rootCAs/%s.pem", exchange)

			rootCertPem, err := os.ReadFile(rootCAFile)

			if err != nil {
				return fmt.Errorf("failed to read root CA file for %s: %w", exchange, err)
			}

			block, _ := pem.Decode(rootCertPem)
			if block == nil || block.Type != "CERTIFICATE" {
				return fmt.Errorf("failed to decode PEM block for %s", exchange)
			}

    		// block.Bytes contains the DER
    		derData := block.Bytes

			if !bytes.Equal(derData, rootCert.Raw) {
				return fmt.Errorf("root CA mismatch for %s", exchange)
			}

			return nil
		},
	}

	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
	}

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = client
	retryClient.Logger = logger.Logger
	retryClient.RetryWaitMin = 2 * time.Second
	retryClient.RetryWaitMax = 3 * time.Second
	retryClient.RetryMax = maxRetries

	return retryClient
}

// FetchPriceFromExchange fetches price and volume data from a specific exchange.
//
// This function performs the following steps sequentially:
//  1. Retrieves the exchange configuration for the given exchange.
//  2. Replace the symbol in the endpoint template.
//  3. Constructs the full URL for the API request, handling cases where the BaseURL may or may not include the protocol.
//  4. Creates a retryable HTTP client for robust network requests.
//  5. Builds an HTTP GET request with the provided context.
//  6. Executes the HTTP request and handles any network errors.
//  7. Checks the HTTP response status code for success.
//  8. Reads the response body from the exchange API.
//  9. Attempts to decode the response body as a JSON object. If decoding fails and the exchange is "gate.io", attempts to decode as a JSON array and adapts the data structure accordingly.
//  10. Parses the price and volume from the decoded response using the appropriate exchange-specific parser.
//  11. Returns an ExchangePrice struct with the parsed data, or an error if any step fails.
//
// Parameters:
//   - ctx: The context for request cancellation and logging.
//   - exchange: The key identifying the exchange (e.g., "binance").
//   - token: The token (e.g., "BTC").
//   - symbol: The trading symbol (e.g., "BTCUSDT").
//
// Returns:
//   - *ExchangePrice: The parsed price and volume data from the exchange.
//   - *appErrors.AppError: An application error if any step fails, otherwise nil.
func (c *PriceFeedClient) FetchPriceFromExchange(ctx context.Context, exchange, token, symbol string) (*ExchangePrice, *appErrors.AppError) {
	reqLogger := logger.FromContext(ctx)

	// Step 1: Get exchange configuration.
	config, exists := c.exchangeConfigs[exchange]
	if !exists {
		reqLogger.Error("Exchange not configured", "exchange", exchange)
		return nil, appErrors.ErrExchangeNotConfigured
	}

	// Step 2: Replace the symbol in the endpoint template.
	endpoint := strings.Replace(config.EndpointTemplate, "{symbol}", symbol, 1)

	// Step 3: Construct the full URL. Accepting protocol scheme in the base URL for unit testing.
	var url string
	if strings.HasPrefix(config.BaseURL, "https://") || strings.HasPrefix(config.BaseURL, "http://") {
		url = fmt.Sprintf("%s%s", config.BaseURL, endpoint)
	} else {
		url = fmt.Sprintf("https://%s%s", config.BaseURL, endpoint)
	}

	// Step 4: Create retryable HTTP client.
	httpClient := GetRetryableHTTPClientForExchange(exchange, 1)

	// Step 5: Create request with context.
	req, err := retryablehttp.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		reqLogger.Error("Error creating HTTP request", "error", err, "exchange", exchange, "token", token, "symbol", symbol)
		return nil, appErrors.ErrCreatingExchangeRequest
	}

	// Step 6: Execute the HTTP request.
	resp, err := httpClient.Do(req)
	if err != nil {
		reqLogger.Error("Error fetching price from exchange", "error", err, "exchange", exchange, "token", token, "symbol", symbol)
		return nil, appErrors.ErrFetchingFromExchange
	}
	defer resp.Body.Close()

	// Step 7: Check for valid HTTP status code.
	if resp.StatusCode != http.StatusOK {
		_, err := io.Copy(io.Discard, resp.Body)
		if err != nil {
			reqLogger.Warn("Error draining response body", "error", err)
		}
		reqLogger.Error("Invalid status code", "status_code", resp.StatusCode, "exchange", exchange, "token", token, "symbol", symbol)
		return nil, appErrors.ErrExchangeInvalidStatusCode
	}

	// Step 8: Read the response body.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		reqLogger.Error("Error reading response body", "error", err, "exchange", exchange, "token", token, "symbol", symbol)
		return nil, appErrors.ErrReadingExchangeResponse
	}

	if !json.Valid(bodyBytes) {
		reqLogger.Error("Invalid JSON response", "body", string(bodyBytes))
		return nil, appErrors.ErrDecodingExchangeResponse
	}

	// Step 10: Parse price and volume from the decoded response.
	price, volume, parseErr := c.parseExchangeResponse(exchange, bodyBytes)
	if parseErr != nil {
		reqLogger.Error("Error parsing exchange response", "error", parseErr, "exchange", exchange, "token", token, "symbol", symbol)
		return nil, appErrors.ErrParsingExchangeResponse
	}

	// Step 11: Return the parsed ExchangePrice.
	return &ExchangePrice{
		Exchange: config.Name,
		Price:    price,
		Volume:   volume,
		Token:    token,
		Symbol:   symbol,
	}, nil
}

// CalculateVolumeWeightedAverage calculates the volume-weighted average price
func CalculateVolumeWeightedAverage(prices []ExchangePrice) (*big.Rat, *big.Rat, int) {
	if len(prices) == 0 {
		return nil, nil, 0
	}

	totalVolume := big.NewRat(0, 1)
	weightedSum := big.NewRat(0, 1)
	exchanges := make(map[string]bool)

	for _, p := range prices {
		if p.Price == "" || p.Volume == "" {
			continue
		}

		volumeRat, ok := new(big.Rat).SetString(p.Volume)
		if !ok || volumeRat.Sign() <= 0 {
			continue
		}

		priceRat, ok := new(big.Rat).SetString(p.Price)
		if !ok || priceRat.Sign() <= 0 {
			continue
		}

		weightedSum.Add(weightedSum, priceRat.Mul(priceRat, volumeRat))
		totalVolume.Add(totalVolume, volumeRat)
		if _, exists := exchanges[p.Exchange]; !exists {
			exchanges[p.Exchange] = true
		}
	}

	if totalVolume.Sign() <= 0 {
		return nil, nil, 0
	}

	volumeWeightedAvg := new(big.Rat).Quo(weightedSum, totalVolume)
	return volumeWeightedAvg, totalVolume, len(exchanges)
}

// GetPriceFeed fetches and calculates the volume-weighted average price for a given token
func (c *PriceFeedClient) GetPriceFeed(ctx context.Context, tokenName string, precision uint) (*PriceFeedResult, *appErrors.AppError) {
	reqLogger := logger.FromContext(ctx)

	exchanges, exists := c.tokenExchanges[strings.ToUpper(tokenName)]
	if !exists {
		reqLogger.Error("Invalid token", "token", tokenName)
		return nil, appErrors.ErrTokenNotSupported
	}

	if len(exchanges) == 0 {
		reqLogger.Error("No exchanges configured for token", "token", tokenName)
		return nil, appErrors.ErrExchangeNotConfigured
	}

	var exchangePrices []ExchangePrice

	totalTradingPairs := len(c.tokenTradingPairs[tokenName])
	if totalTradingPairs == 0 {
		reqLogger.Error("No trading pairs configured for token", "token", tokenName)
		return nil, appErrors.ErrNoTradingPairsConfigured
	}

	// Create a buffered channel to collect results from goroutines
	// Buffer size matches the number of trading pairs to prevent blocking
	results := make(chan fetchResult, totalTradingPairs)

	// Launch concurrent goroutines to fetch prices from each exchange
	// Each goroutine fetches data independently and sends results through the channel
	for _, exchange := range exchanges {
		token := strings.ToUpper(tokenName)
		// Step 1: Get exchange configuration.
		config, exists := c.exchangeConfigs[exchange]
		if !exists {
			reqLogger.Error("Exchange not configured", "exchange", exchange)
			return nil, appErrors.ErrExchangeNotConfigured
		}

		// Step 2: Get symbol list and construct endpoint for the symbol.
		symbolList, exists := config.Symbols[token]
		if !exists {
			reqLogger.Error("Token not supported by exchange", "token", token, "exchange", exchange)
			return nil, appErrors.ErrTokenNotSupported
		}

		if len(symbolList) == 0 {
			reqLogger.Error("No trading pairs configured for token", "token", token, "exchange", exchange)
			return nil, appErrors.ErrSymbolNotConfigured
		}

		for _, symbol := range symbolList {
			go func(ex string, tk string, sym string) {
				price, err := c.FetchPriceFromExchange(ctx, ex, tk, sym)
				results <- fetchResult{price: price, err: err, exchange: ex}
			}(exchange, token, symbol)
		}
	}

	// Collect results from all goroutines
	// Process results in the order they complete, not necessarily the order of exchanges
	for i := 0; i < totalTradingPairs; i++ {
		result := <-results
		if result.err != nil {
			metrics.RecordExchangeApiError(result.exchange, strconv.Itoa(int(result.err.Code)))
			// Log the error but continue processing other exchanges
			// This ensures the system is resilient to individual exchange failures
			reqLogger.Error("Failed to fetch token price", "exchange", result.exchange, "token", tokenName, "error", result.err.Details)
			continue
		}
		if result.price != nil {
			// Only add valid price data to the collection
			exchangePrices = append(exchangePrices, *result.price)
		}
	}

	reqLogger.Debug("Total trading pairs", "totalTradingPairs", totalTradingPairs)

	// Calculate volume-weighted average
	volumeWeightedAvg, totalVolume, exchangeCount := CalculateVolumeWeightedAverage(exchangePrices)

	metrics.RecordPriceFeedExchangeCount(tokenName, exchangeCount)

	// Ensure at least the minimum number of exchanges responded successfully
	if exchangeCount < configs.GetMinExchangesRequired() {
		metrics.RecordError("insufficient_exchange_data", "price_feed")
		reqLogger.Error("Insufficient exchange data", "exchangeCount", exchangeCount, "minExchangesRequired", configs.GetMinExchangesRequired())
		return nil, appErrors.ErrInsufficientExchangeData
	}

	volumeWeightAvgStr := Truncate(volumeWeightedAvg, int(precision))
	totalVolumeStr := Truncate(totalVolume, int(precision))

	return &PriceFeedResult{
		Token:             strings.ToUpper(tokenName),
		VolumeWeightedAvg: volumeWeightAvgStr,
		TotalVolume:       totalVolumeStr,
		ExchangeCount:     exchangeCount,
		Timestamp:         time.Now().Unix(),
		ExchangePrices:    exchangePrices,
		Success:           true,
	}, nil
}

// ExtractPriceFeedData handles price feed requests and always returns the volume-weighted average price (VWAP)
// This ensures consistent and reliable price data for oracle attestations
func (c *PriceFeedClient) ExtractPriceFeedData(ctx context.Context, attestationRequest attestation.AttestationRequest, token string) (ExtractDataResult, *appErrors.AppError) {

	// Start the price feed extraction
	priceFeedStart := time.Now()

	status := "failed"
	defer func() {
		priceFeedDuration := time.Since(priceFeedStart).Seconds()
		metrics.RecordPriceFeedRequest(token, status, priceFeedDuration)
	}()

	reqLogger := logger.FromContext(ctx)

	// Get the price feed data
	result, appErr := c.GetPriceFeed(ctx, token, attestationRequest.EncodingOptions.Precision)

	if appErr != nil {
		reqLogger.Error("Error getting price feed for ", "token", token, "error", appErr)
		return ExtractDataResult{}, appErr
	}

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		reqLogger.Error("Error marshalling price feed data", "error", err)
		return ExtractDataResult{}, appErrors.ErrEncodingPriceFeedData
	}

	attestationData := result.VolumeWeightedAvg

	status = "success"

	return ExtractDataResult{
		ResponseBody:    string(jsonBytes),
		AttestationData: attestationData,
		StatusCode:      http.StatusOK,
	}, nil
}
