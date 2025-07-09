package data_extraction

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/metrics"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
)

// ExtractDataResult represents the result of data extraction
type ExtractDataResult struct {
	ResponseBody    string
	AttestationData string
	StatusCode      int
}

// makeHTTPRequest creates and executes an HTTP request with common configuration
func makeHTTPRequest(ctx context.Context, attestationRequest attestation.AttestationRequest) (*http.Response, *appErrors.AppError) {
	start := time.Now()
	var statusCode int
	defer func() {
		duration := time.Since(start).Seconds()
		target := extractTargetFromURL(attestationRequest.Url)
		metrics.RecordExternalHttpRequest(target, statusCode, duration)
	}()

	// Create a new reader for the request body
	var bodyReader io.Reader

	// Get logger from context (includes request ID)
	reqLogger := logger.FromContext(ctx)

	// If there's a request body, create a new reader for it
	if attestationRequest.RequestBody != nil {
		bodyReader = strings.NewReader(*attestationRequest.RequestBody)
	}

	var url string
	if strings.HasPrefix(attestationRequest.Url, "http") || strings.HasPrefix(attestationRequest.Url, "https") {
		url = attestationRequest.Url
	} else {
		url = "https://" + attestationRequest.Url
	}

	// Create a new request with the context, URL, and body reader
	req, err := retryablehttp.NewRequestWithContext(ctx, attestationRequest.RequestMethod, url, bodyReader)
	if err != nil {
		reqLogger.Error("Error while creating HTTP request", "error", err, "url", url)
		metrics.RecordError("http_request_creation_failed", "data_extractor")
		return nil, appErrors.NewAppError(appErrors.ErrInvalidHTTPRequest)
	}

	// Set the request headers.
	for key, value := range attestationRequest.RequestHeaders {
		req.Header.Set(key, value)
	}

	// Set the request content type.
	if attestationRequest.RequestContentType != nil {
		req.Header.Set("Content-Type", *attestationRequest.RequestContentType)
	}

	// Create the client.
	client := utils.GetRetryableHTTPClient(3)

	// Do the request with context.
	resp, err := client.Do(req)
	if err != nil {
		reqLogger.Error("Error while fetching data", "error", err, "url", url)
		metrics.RecordError("http_request_failed", "data_extractor")
		statusCode = 0 // Set to 0 to indicate an error
		return nil, appErrors.NewAppError(appErrors.ErrFetchingData)
	}

	// Set the status code.
	statusCode = resp.StatusCode

	// Check if the status code is an error code.
	if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		reqLogger.Error("Error while fetching data", "status_code", resp.StatusCode, "url", url)
		metrics.RecordError("http_error_response", "data_extractor")
		return resp, appErrors.NewAppErrorWithResponseStatus(appErrors.ErrFetchingData, resp.StatusCode)
	}

	return resp, nil
}

// extractTargetFromURL extracts a simplified target name from URL for metrics
func extractTargetFromURL(url string) string {
	if strings.HasPrefix(url, "http://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "https://") {
		url = url[8:]
	}
	if idx := strings.Index(url, "/"); idx != -1 {
		url = url[:idx]
	}
	if idx := strings.Index(url, ":"); idx != -1 {
		url = url[:idx]
	}
	return url
}

// ApplyFloatPrecision applies precision formatting for float values
func applyFloatPrecision(valueStr string, precision uint) string {
	if precision == 0 {
		return valueStr
	}

	splitted := strings.SplitN(valueStr, ".", 2) // Split at most once

	// If there's no decimal point, or it's already within limits.
	if len(splitted) == 1 || len(splitted[1]) <= int(precision) {
		// Safe to use as-is.
		return valueStr
	} else {
		// Truncate the decimal part to max precision.
		splitted[1] = splitted[1][:precision]
		return splitted[0] + "." + splitted[1]
	}
}

// ExtractDataFromTargetURL fetches the data from the attestation request target URL.
// This is the main entry point that routes to specific extractors based on the request type.
func ExtractDataFromTargetURL(ctx context.Context, attestationRequest attestation.AttestationRequest) (ExtractDataResult, *appErrors.AppError) {
	// Get logger from context (includes request ID)
	reqLogger := logger.FromContext(ctx)

	// Check if the URL is a price feed request
	if attestationRequest.Url == constants.PriceFeedBtcUrl ||
		attestationRequest.Url == constants.PriceFeedEthUrl ||
		attestationRequest.Url == constants.PriceFeedAleoUrl {
		reqLogger.Debug("Processing price feed request", "url", attestationRequest.Url)
		asset := extractAssetFromPriceFeedURL(attestationRequest.Url)

		// Start the price feed extraction
		priceFeedStart := time.Now()
		result, err := ExtractPriceFeedData(ctx, attestationRequest)

		// Record the price feed extraction duration
		priceFeedDuration := time.Since(priceFeedStart).Seconds()

		// Record the price feed request status
		if err != nil {
			metrics.RecordPriceFeedRequest(asset, "failed", priceFeedDuration)
		} else {
			metrics.RecordPriceFeedRequest(asset, "success", priceFeedDuration)
		}

		// Return the result
		return result, err
	} else if attestationRequest.ResponseFormat == "html" {
		// Process HTML request
		reqLogger.Info("Processing HTML request", "url", attestationRequest.Url)
		return ExtractDataFromHTML(ctx, attestationRequest)
	} else if attestationRequest.ResponseFormat == "json" {
		// Process JSON request
		reqLogger.Debug("Processing JSON request", "url", attestationRequest.Url)
		return ExtractDataFromJSON(ctx, attestationRequest)
	} else {
		// Return an error for invalid response format
		reqLogger.Error("Invalid response format", "format", attestationRequest.ResponseFormat)
		metrics.RecordError("invalid_response_format", "data_extractor")
		return ExtractDataResult{
			StatusCode: http.StatusNotFound,
		}, appErrors.NewAppError(appErrors.ErrInvalidResponseFormat)
	}
}

// extractAssetFromPriceFeedURL extracts the asset name from price feed URL
func extractAssetFromPriceFeedURL(url string) string {
	switch url {
	case constants.PriceFeedBtcUrl:
		return "btc"
	case constants.PriceFeedEthUrl:
		return "eth"
	case constants.PriceFeedAleoUrl:
		return "aleo"
	default:
		return "unknown"
	}
}
