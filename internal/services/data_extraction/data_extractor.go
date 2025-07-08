package data_extraction

import (
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
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
	// Create the body reader.
	var bodyReader io.Reader

	reqLogger := logger.FromContext(ctx)

	// Check if the request body is not nil.
	if attestationRequest.RequestBody != nil {
		bodyReader = strings.NewReader(*attestationRequest.RequestBody)
	}

	// Create the URL.
	var url string
	if strings.HasPrefix(attestationRequest.Url, "http") || strings.HasPrefix(attestationRequest.Url, "https") {
		url = attestationRequest.Url
	} else {
		url = "https://" + attestationRequest.Url
	}

	// Create the request with context.
	req, err := retryablehttp.NewRequestWithContext(ctx, attestationRequest.RequestMethod, url, bodyReader)
	if err != nil {
		reqLogger.Error("Error while creating HTTP request", "error", err, "url", url)
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
		return nil, appErrors.NewAppError(appErrors.ErrFetchingData)
	}

	// Check if the status code is greater than or equal to 400 and less than 600.
	if resp.StatusCode >= 400 && resp.StatusCode < 600 {
		reqLogger.Error("Error while fetching data", "status_code", resp.StatusCode, "url", url)
		return resp, appErrors.NewAppErrorWithResponseStatus(appErrors.ErrFetchingData, resp.StatusCode)
	}

	return resp, nil
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
		return ExtractPriceFeedData(ctx, attestationRequest)
	} else if attestationRequest.ResponseFormat == "html" {
		reqLogger.Info("Processing HTML request", "url", attestationRequest.Url)
		return ExtractDataFromHTML(ctx, attestationRequest)
	} else if attestationRequest.ResponseFormat == "json" {
		reqLogger.Debug("Processing JSON request", "url", attestationRequest.Url)
		return ExtractDataFromJSON(ctx, attestationRequest)
	} else {
		reqLogger.Error("Invalid response format", "format", attestationRequest.ResponseFormat)
		return ExtractDataResult{
			StatusCode: http.StatusNotFound,
		}, appErrors.NewAppError(appErrors.ErrInvalidResponseFormat)
	}
}
