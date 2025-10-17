package data_extraction

import (
	"context"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/common"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	httpUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/httputil"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/metrics"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
)

// ExtractDataResult represents the result of data extraction
type ExtractDataResult struct {
	ResponseBody    string // The response body.
	AttestationData string // The attestation data.
	StatusCode      int    // The status code.
}

// Truncate returns r truncated to `prec` decimal places as a *big.Rat.
func Truncate(r *big.Rat, prec int) string {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(prec)), nil)

	// numerator * 10^prec
	numScaled := new(big.Int).Mul(r.Num(), scale)

	// truncate by integer division
	quo := new(big.Int).Quo(numScaled, r.Denom())

	intPart := new(big.Int).Quo(quo, scale)
	fracPart := new(big.Int).Mod(quo, scale)
	fracStr := fmt.Sprintf("%0*s", prec, fracPart.String())

	return fmt.Sprintf("%s.%s", intPart.String(), fracStr)
}

func formatAttestationData(ctx context.Context, attestationData string, encodingOptionValue string, precision uint) (string, *appErrors.AppError) {
	reqLogger := logger.FromContext(ctx)

	switch encodingOptionValue {
	case constants.EncodingOptionInt:
		valueInt, ok := new(big.Int).SetString(attestationData, 10)
		if !ok {
			return "", appErrors.ErrInvalidRationalNumber
		}
		return valueInt.String(), nil
	case constants.EncodingOptionFloat:
		valueInt, ok := new(big.Rat).SetString(attestationData)
		if !ok {
			return "", appErrors.ErrInvalidRationalNumber
		}
		truncatedValue := Truncate(valueInt, int(precision))
		return truncatedValue, nil
	case constants.EncodingOptionString:
		return attestationData, nil
	default:
		reqLogger.Error("Invalid encoding option", "encodingOptionValue", encodingOptionValue)
		return "", appErrors.ErrInvalidEncodingOption
	}
}

// makeHTTPRequestToTarget creates and executes an HTTP request with common configuration
func makeHTTPRequestToTarget(ctx context.Context, attestationRequest attestation.AttestationRequest) (*http.Response, *appErrors.AppError) {
	start := time.Now()
	var statusCode int

	// Get logger from context (includes request ID)
	reqLogger := logger.FromContext(ctx)

	defer func() {
		duration := time.Since(start).Seconds()
		target, _ := common.GetHostnameFromURL(attestationRequest.Url)
		metrics.RecordExternalHttpRequest(target, statusCode, duration)
	}()

	// Create a new reader for the request body
	var bodyReader io.Reader

	// If there's a request body, create a new reader for it
	if attestationRequest.RequestBody != nil {
		bodyReader = strings.NewReader(*attestationRequest.RequestBody)
	}

	var normalizedURL, err = common.NormalizeURL(attestationRequest.Url)

	logger.Debug("Normalized URL", "url", normalizedURL)

	if err != nil {
		reqLogger.Error("Error while normalizing URL", "error", err, "url", attestationRequest.Url)
		metrics.RecordError("http_request_creation_failed", "data_extractor")
		return nil, err
	}

	// Create a new request with the context, URL, and body reader
	req, httpError := retryablehttp.NewRequestWithContext(ctx, attestationRequest.RequestMethod, normalizedURL, bodyReader)
	if httpError != nil {
		reqLogger.Error("Error while creating HTTP request", "error", httpError, "url", normalizedURL)
		metrics.RecordError("http_request_creation_failed", "data_extractor")
		return nil, appErrors.ErrInvalidHTTPRequest
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
	client := httpUtil.GetRetryableHTTPClient(3)

	// Do the request with context.
	resp, requestError := client.Do(req)
	if requestError != nil {
		reqLogger.Error("Error while fetching data", "error", requestError, "url", normalizedURL)
		metrics.RecordError("http_request_failed", "data_extractor")
		statusCode = 0 // Set to 0 to indicate an error
		return nil, appErrors.ErrFetchingData
	}

	// Set the status code.
	statusCode = resp.StatusCode

	if statusCode != http.StatusOK {
		_, err := io.Copy(io.Discard, resp.Body)
		if err != nil {
			reqLogger.Warn("Error draining response body", "error", err)
		}
		resp.Body.Close()
		reqLogger.Error("Error while fetching data", "status_code", statusCode, "url", normalizedURL)
		metrics.RecordError("http_error_response", "data_extractor")
		return nil, appErrors.ErrInvalidStatusCode.WithResponseStatusCode(statusCode)
	}

	return resp, nil
}

// // ApplyFloatPrecision applies precision formatting for float values
// func applyFloatPrecision(valueStr string, precision uint) string {
// 	if precision == 0 {
// 		return valueStr
// 	}

// 	splitted := strings.SplitN(valueStr, ".", 2) // Split at most once

// 	// If there's no decimal point, or it's already within limits.
// 	if len(splitted) == 1 || len(splitted[1]) <= int(precision) {
// 		// Safe to use as-is.
// 		return valueStr
// 	} else {
// 		// Truncate the decimal part to max precision.
// 		splitted[1] = splitted[1][:precision]
// 		return splitted[0] + "." + splitted[1]
// 	}
// }

// ExtractDataFromTargetURL fetches the data from the attestation request target URL.
// This is the main entry point that routes to specific extractors based on the request type.
func ExtractDataFromTargetURL(ctx context.Context, attestationRequest attestation.AttestationRequest, timestamp int64) (ExtractDataResult, *appErrors.AppError) {
	// Get logger from context (includes request ID)
	reqLogger := logger.FromContext(ctx)

	if common.IsPriceFeedURL(attestationRequest.Url) {
		reqLogger.Debug("Processing price feed request", "url", attestationRequest.Url)
		token := common.ExtractTokenFromPriceFeedURL(attestationRequest.Url)
		priceFeedClient := NewPriceFeedClient()
		return priceFeedClient.ExtractPriceFeedData(ctx, attestationRequest, token, timestamp)
	}

	switch attestationRequest.ResponseFormat {
	case constants.ResponseFormatHTML:
		return ExtractDataFromHTML(ctx, attestationRequest)
	case constants.ResponseFormatJSON:
		return ExtractDataFromJSON(ctx, attestationRequest)
	default:
		return ExtractDataResult{}, appErrors.ErrInvalidResponseFormat
	}
}
