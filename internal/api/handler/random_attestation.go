package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"time"

	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/common"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	httpUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/httputil"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/metrics"
	attestation "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/sgx"
)

var (
	minMaxValue = big.NewInt(2)                        // >1
	maxMaxValue = new(big.Int).Lsh(big.NewInt(1), 127) // 2^127
)

// GenerateAttestedRandom handles the request to generate an attested random number
func GenerateAttestedRandom(w http.ResponseWriter, req *http.Request) {
	start := time.Now()

	// Set the status to failed by default
	status := "failed"
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.RecordRandomNumberGeneration(status, duration)
	}()

	// Get logger from context (request ID automatically included by middleware)
	reqLogger := logger.FromContext(req.Context())

	reqLogger.Debug("Received attested random request", "query", req.URL.RawQuery)

	// Get the max parameter from query string
	maxStr := req.URL.Query().Get("max")
	if maxStr == "" {
		reqLogger.Error("Missing max parameter")
		metrics.RecordError("missing_max_parameter", "random_handler")
		httpUtil.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrMissingMaxParameter)
		return
	}

	// Parse the max value as a big integer
	max, ok := new(big.Int).SetString(maxStr, 10)
	if !ok {
		reqLogger.Error("Invalid max parameter format", "max", maxStr)
		metrics.RecordError("invalid_max_format", "random_handler")
		httpUtil.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidMaxValueFormat)
		return
	}

	// Validate the max value: must be > 1 and ≤ 2^127
	if max.Cmp(minMaxValue) < 0 {
		reqLogger.Error("Max must be greater than 1", "max", maxStr)
		metrics.RecordError("max_too_small", "random_handler")
		httpUtil.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidMaxValue)
		return
	}

	// Check if max > 2^127
	if max.Cmp(maxMaxValue) > 0 {
		reqLogger.Error("Max must be less than or equal to 2^127", "max", maxStr)
		metrics.RecordError("max_too_large", "random_handler")
		httpUtil.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidMaxValue)
		return
	}

	reqLogger.Debug("Validated max parameter", "max", maxStr)

	attestationRequest := attestation.AttestationRequest{
		Url:            fmt.Sprintf("crypto/rand:%s", max.String()),
		RequestMethod:  "GET",
		ResponseFormat: "json",
		EncodingOptions: encoding.EncodingOptions{
			Value:     "int",
			Precision: 0,
		},
	}

	// Generate a cryptographically secure random number in range [0, max)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		reqLogger.Error("Error generating random number", "error", err)
		metrics.RecordError("random_generation_failed", "random_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, appErrors.ErrGeneratingRandomNumber)
		return
	}

	reqLogger.Debug("Generated random number", "number", randomNumber.String())

	// Get the timestamp
	timestamp, appError := common.GetTimestampFromRoughtime()
	if appError != nil {
		reqLogger.Error("Failed to get timestamp from roughtime server", "error", appError)
		metrics.RecordError("timestamp_fetch_failed", "random_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, appError)
		return
	}

	statusCode := 200
	attestationData := randomNumber.String()

	aleoBlockHeight, blockHeightError := common.GetAleoCurrentBlockHeight()
	if blockHeightError != nil {
		reqLogger.Error("Failed to get Aleo block height", "error", err)
		metrics.RecordError("aleo_block_height_fetch_failed", "random_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, blockHeightError)
		return
	}

	reqLogger.Debug("Preparing data for quote generation")
	quotePrepData, appError := attestation.PrepareDataForQuoteGeneration(statusCode, attestationData, uint64(timestamp), int64(aleoBlockHeight), attestationRequest)

	// Check if the error is not nil.
	if appError != nil {
		reqLogger.Error("Error preparing data for quote generation", "error", appError)
		metrics.RecordError("quote_prep_failed", "random_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, appError)
		return
	}

	reqLogger.Debug("Generating SGX quote")

	// Generate the quote.
	quoteStart := time.Now()
	quote, appError := sgx.GenerateQuote(quotePrepData.AttestationHash)
	quoteDuration := time.Since(quoteStart).Seconds()

	if appError != nil {
		reqLogger.Error("Error generating quote", "error", appError)
		metrics.RecordSgxQuoteGeneration("failed", quoteDuration)
		metrics.RecordError("quote_generation_failed", "random_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, appError)
		return
	}

	metrics.RecordSgxQuoteGeneration("success", quoteDuration)
	reqLogger.Debug("Building oracle data after quote")

	// Build the complete oracle data after the quote.
	oracleData, appError := attestation.BuildCompleteOracleData(quotePrepData, quote)
	if appError != nil {
		reqLogger.Error("Error preparing oracle data after quote", "error", appError)
		metrics.RecordError("oracle_data_build_failed", "random_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, appError)
		return
	}

	// Create the random response
	response := attestation.AttestationResponse{
		ReportType:           "sgx",
		AttestationRequest:   attestationRequest,
		AttestationTimestamp: timestamp,
		AttestationReport:    base64.StdEncoding.EncodeToString(quote),
		OracleData:           *oracleData,
		ResponseBody:         randomNumber.String(),
		AttestationData:      randomNumber.String(),
		ResponseStatusCode:   statusCode,
	}

	reqLogger.Debug("Successfully generated attested random response")

	// Set the status to success
	status = "success"

	// Write the JSON success response
	httpUtil.WriteJsonSuccess(w, http.StatusOK, response)
}
