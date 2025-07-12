package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/metrics"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/data_extraction"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
)

// GenerateAttestationReport handles the request to generate an attestation report.
func GenerateAttestationReport(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	status := "failed"
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.RecordAttestationRequest("attestation", status, duration)
	}()

	// Close the request body.
	defer req.Body.Close()

	// Get logger from context (request ID automatically included by middleware)
	ctx := req.Context()
	reqLogger := logger.FromContext(ctx)

	// Log the incoming request
	reqLogger.Debug("Attestation report request received", "method", req.Method, "path", req.URL.Path)

	// Validate Content-Type
	if req.Header.Get("Content-Type") != "application/json" {
		reqLogger.Error("Invalid Content-Type", "content_type", req.Header.Get("Content-Type"), "method", req.Method, "path", req.URL.Path)
		metrics.RecordError("invalid_content_type", "attestation_handler")
		utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, "")
		return
	}

	// Create the attestation request.
	var attestationRequestWithDebug attestation.AttestationRequestWithDebug

	// Decode the request body.
	if err := json.NewDecoder(req.Body).Decode(&attestationRequestWithDebug); err != nil {
		reqLogger.Error("Failed to decode request body", "error", err)
		metrics.RecordError("json_decode_failed", "attestation_handler")
		utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, "")
		return
	}

	attestationRequest := attestationRequestWithDebug.AttestationRequest

	reqLogger.Debug("Processing attestation request", "url", attestationRequest.Url, "debug", attestationRequestWithDebug.DebugRequest)

	if err := attestationRequest.Validate(configs.GetWhitelistedDomains()); err != nil {
		reqLogger.Error("Attestation request validation failed", "error", err)
		metrics.RecordError("validation_failed", "attestation_handler")
		utils.WriteJsonError(w, http.StatusBadRequest, *err, "")
		return
	}

	// Get the timestamp.
	timestamp := time.Now().Unix()
	reqLogger.Debug("Using timestamp", "timestamp", timestamp)

	// Fetch the data from the attestation request.
	reqLogger.Debug("Fetching data from target URL", "url", attestationRequest.Url)
	extractStart := time.Now()
	extractDataResult, err := data_extraction.ExtractDataFromTargetURL(ctx, attestationRequest)
	extractDuration := time.Since(extractStart).Seconds()

	// Check if the error is not nil.
	if err != nil {
		reqLogger.Error("Failed to extract data from target URL", "error", err)
		metrics.RecordError("data_extraction_failed", "attestation_handler")
		metrics.RecordDataExtraction(attestationRequest.ResponseFormat, "failed", extractDuration)
		utils.WriteJsonError(w, http.StatusInternalServerError, *err, "")
		return
	}

	if err := extractDataResult.ValidateAttestationData(attestationRequest.EncodingOptions.Value); err != nil {
		reqLogger.Error("Invalid attestation data", "error", err)
		metrics.RecordError("invalid_attestation_data", "attestation_handler")
		utils.WriteJsonError(w, http.StatusBadRequest, *err, "")
		return
	}

	reqLogger.Debug("Successfully extracted data", "status_code", extractDataResult.StatusCode, "data", extractDataResult.AttestationData)
	metrics.RecordDataExtraction(attestationRequest.ResponseFormat, "success", extractDuration)
	metrics.RecordAttestationDataSize("attestation", len(extractDataResult.AttestationData))

	attestationRequest.MaskUnacceptedHeaders()

	if attestationRequestWithDebug.DebugRequest {
		reqLogger.Debug("Returning debug response")

		// Create the attestation response.
		response := &attestation.DebugAttestationResponse{
			ReportType:           "sgx",
			AttestationRequest:   attestationRequest,
			AttestationTimestamp: timestamp,
			ResponseBody:         extractDataResult.ResponseBody,
			AttestationData:      extractDataResult.AttestationData,
			ResponseStatusCode:   extractDataResult.StatusCode,
		}

		reqLogger.Debug("Debug attestation report generated")
		utils.WriteJsonSuccess(w, http.StatusOK, response)
		return
	}

	// Prepare the oracle data before the quote.
	reqLogger.Debug("Preparing data for quote generation")
	quotePrepData, err := attestation.PrepareDataForQuoteGeneration(extractDataResult.StatusCode, extractDataResult.AttestationData, uint64(timestamp), attestationRequest)

	// Check if the error is not nil.
	if err != nil {
		reqLogger.Error("Failed to prepare data for quote generation", "error", err)
		metrics.RecordError("quote_prep_failed", "attestation_handler")
		utils.WriteJsonError(w, http.StatusBadRequest, *err, "")
		return
	}

	reqLogger.Debug("Quote preparation successful")

	// Generate the quote.
	reqLogger.Debug("Generating SGX quote")
	quoteStart := time.Now()
	quote, err := attestation.GenerateQuote(quotePrepData.AttestationHash)
	quoteDuration := time.Since(quoteStart).Seconds()

	if err != nil {
		reqLogger.Error("Failed to generate SGX quote", "error", err)
		metrics.RecordSgxQuoteGeneration("failed", quoteDuration)
		metrics.RecordError("quote_generation_failed", "attestation_handler")
		utils.WriteJsonError(w, http.StatusInternalServerError, *err, "")
		return
	}

	metrics.RecordSgxQuoteGeneration("success", quoteDuration)
	reqLogger.Debug("SGX quote generated successfully")

	// Prepare the oracle data after the quote.
	reqLogger.Debug("Building complete oracle data")
	oracleData, err := attestation.BuildCompleteOracleData(quotePrepData, quote)

	if err != nil {
		reqLogger.Error("Failed to build complete oracle data", "error", err)
		metrics.RecordError("oracle_data_build_failed", "attestation_handler")
		utils.WriteJsonError(w, http.StatusInternalServerError, *err, "")
		return
	}

	reqLogger.Debug("Oracle data built successfully")

	// Create the attestation response.
	response := &attestation.AttestationResponse{
		ReportType:           "sgx",
		AttestationRequest:   attestationRequest,
		AttestationTimestamp: timestamp,
		ResponseBody:         extractDataResult.ResponseBody,
		AttestationData:      extractDataResult.AttestationData,
		ResponseStatusCode:   extractDataResult.StatusCode,
		AttestationReport:    base64.StdEncoding.EncodeToString(quote),
		OracleData:           *oracleData,
	}

	// Log successful completion
	reqLogger.Debug("Attestation report generated successfully")

	status = "success"

	// Write the JSON success response.
	utils.WriteJsonSuccess(w, http.StatusOK, response)
}
