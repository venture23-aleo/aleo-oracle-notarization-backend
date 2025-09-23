package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	httpUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/httputil"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/metrics"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
	data_extraction "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/dataextraction"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/sgx"
)

// GenerateAttestationReport handles the request to generate an attestation report.
func GenerateAttestationReport(w http.ResponseWriter, req *http.Request) {
	start := time.Now()
	status := "failed"

	// Close the request body.
	defer req.Body.Close()

	defer func() {
		duration := time.Since(start).Seconds()
		metrics.RecordAttestationRequest("attestation", status, duration)
	}()

	// Get logger from context (request ID automatically included by middleware)
	ctx := req.Context()
	reqLogger := logger.FromContext(ctx)

	// Log the incoming request
	reqLogger.Debug("Attestation report request received", "method", req.Method, "path", req.URL.Path)

	contentType := req.Header.Get("Content-Type")

	// Validate Content-Type
	if !strings.Contains(contentType, "application/json") {
		reqLogger.Error("Invalid Content-Type", "content_type", contentType)
		metrics.RecordError("invalid_content_type", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusUnsupportedMediaType, appErrors.ErrInvalidContentType)
		return
	}

	// Limit the request body size.
	req.Body = http.MaxBytesReader(w, req.Body, constants.MaxRequestBodySize)

	// Decode the request body.
	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields() // Disallow unknown fields in the request body.

	// Create the attestation request.
	var attestationRequestWithDebug attestation.AttestationRequestWithDebug

	// Decode the request body.
	if err := decoder.Decode(&attestationRequestWithDebug); err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			reqLogger.Error("Request body too large during decode", "error", err)
			metrics.RecordError("request_body_too_large", "attestation_handler")
			httpUtil.WriteJsonError(w, http.StatusRequestEntityTooLarge, appErrors.ErrRequestBodyTooLarge)
			return
		}
		reqLogger.Error("Failed to decode request body", "error", err)
		metrics.RecordError("json_decode_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrDecodingRequestBody)
		return
	}

	attestationRequest := attestationRequestWithDebug.AttestationRequest

	reqLogger.Debug("Processing attestation request", "url", attestationRequest.Url, "debug", attestationRequestWithDebug.DebugRequest)

	if err := attestationRequest.Validate(); err != nil {
		reqLogger.Error("Attestation request validation failed", "error", err)
		metrics.RecordError("validation_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusBadRequest, err)
		return
	}

	// Get the timestamp.
	timestamp := time.Now().Unix()

	reqLogger.Debug("Fetching data from target URL", "url", attestationRequest.Url, "timestamp", timestamp)

	// Fetch the data from the attestation request.
	extractStart := time.Now()
	extractDataResult, err := data_extraction.ExtractDataFromTargetURL(ctx, attestationRequest, timestamp)
	extractDuration := time.Since(extractStart).Seconds()

	// Check if the error is not nil.
	if err != nil {
		reqLogger.Error("Failed to extract data from target URL", "error", err)
		metrics.RecordError("data_extraction_failed", "attestation_handler")
		metrics.RecordDataExtraction(attestationRequest.ResponseFormat, "failed", extractDuration)
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
		return
	}

	metrics.RecordDataExtraction(attestationRequest.ResponseFormat, "success", extractDuration)

	attestationRequest.MaskUnacceptedHeaders()

	if attestationRequestWithDebug.DebugRequest {
		reqLogger.Debug("Returning debug response")

		// Create the attestation response.
		response := &attestation.DebugAttestationResponse{
			ReportType:           constants.SGXReportType,
			AttestationRequest:   attestationRequest,
			AttestationTimestamp: timestamp,
			ResponseBody:         extractDataResult.ResponseBody,
			AttestationData:      extractDataResult.AttestationData,
			ResponseStatusCode:   extractDataResult.StatusCode,
		}

		reqLogger.Debug("Debug attestation report generated")

		status = "success"

		httpUtil.WriteJsonSuccess(w, http.StatusOK, response)
		return
	}

	// Prepare the oracle data before the quote.
	reqLogger.Debug("Preparing data for quote generation")
	quotePrepData, err := attestation.PrepareDataForQuoteGeneration(extractDataResult.StatusCode, extractDataResult.AttestationData, uint64(timestamp), attestationRequest)

	// Check if the error is not nil.
	if err != nil {
		reqLogger.Error("Failed to prepare data for quote generation", "error", err)
		metrics.RecordError("quote_prep_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
		return
	}

	reqLogger.Debug("Quote preparation successful")

	// Generate the quote.
	reqLogger.Debug("Generating SGX quote")
	quoteStart := time.Now()
	quote, err := sgx.GenerateQuote(quotePrepData.AttestationHash)
	quoteDuration := time.Since(quoteStart).Seconds()

	if err != nil {
		reqLogger.Error("Failed to generate SGX quote", "error", err)
		metrics.RecordSgxQuoteGeneration("failed", quoteDuration)
		metrics.RecordError("quote_generation_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
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
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
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
	httpUtil.WriteJsonSuccess(w, http.StatusOK, response)
}
