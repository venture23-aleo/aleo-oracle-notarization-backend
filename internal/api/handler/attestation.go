package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/common"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	httpUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/httputil"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/metrics"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
	data_extraction "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/dataextraction"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/sgx"
)

func decodeOneOrMany[T any](raw []byte) ([]T, *appErrors.AppError) {
	// Array case
	decArr := json.NewDecoder(bytes.NewReader(raw))
	decArr.DisallowUnknownFields()

	var many []T
	if err := decArr.Decode(&many); err == nil {
		return many, nil
	}

	// Single case
	decObj := json.NewDecoder(bytes.NewReader(raw))
	decObj.DisallowUnknownFields()

	var one T
	if err := decObj.Decode(&one); err == nil {
		return []T{one}, nil
	}

	return nil, appErrors.ErrDecodingRequestBody
}


// GenerateAttestationReport handles the request to generate an attestation report.
// It supports both single token (object) and multiple tokens (array) in a single endpoint.
// Request body can be:
//   - Single token: { "url": "...", "requestMethod": "...", ... } or { "attestationRequest": {...}, "debugRequest": true }
//   - Multiple tokens: [{ "url": "...", ... }, { "url": "...", ... }]
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

		// panic listener
	defer func() {
		if r := recover(); r != nil {
			reqLogger.Error("Panic occurred", "error", r)
			metrics.RecordError("panic_occurred", "attestation_handler")
			httpUtil.WriteJsonError(w, http.StatusInternalServerError, appErrors.ErrInternal)
		}
	}()


	contentType := req.Header.Get("Content-Type")

	// Validate Content-Type
	if !strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		reqLogger.Error("Invalid Content-Type", "content_type", contentType)
		metrics.RecordError("invalid_content_type", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusUnsupportedMediaType, appErrors.ErrInvalidContentType)
		return
	}

	// Limit the request body size.
	req.Body = http.MaxBytesReader(w, req.Body, constants.MaxRequestBodySize)

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		reqLogger.Error("Failed to read request body", "error", err)
		metrics.RecordError("request_body_read_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, appErrors.ErrInternal)
		return
	}

	if !json.Valid(bodyBytes) {
		reqLogger.Error("Invalid JSON request body", "body", string(bodyBytes))
		metrics.RecordError("invalid_json_request_body", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrDecodingRequestBody)
		return
	}

	attestationRequests, decodeErr := decodeOneOrMany[attestation.AttestationRequestWithDebug](bodyBytes)
	if decodeErr != nil {
		reqLogger.Error("Failed to decode attestation requests", "error", decodeErr)
		metrics.RecordError("json_decode_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusBadRequest, decodeErr)
		return
	}

	requestsLength := len(attestationRequests)

	if requestsLength == 0 {
		reqLogger.Error("No attestation requests found")
		metrics.RecordError("no_attestation_requests_found", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrDecodingRequestBody)
		return
	}

	// Check if all the token URLs are the same.
	uniqueTokenURLs := make(map[string]bool)
	for _, attestationRequest := range attestationRequests {
		url := strings.TrimSpace(strings.ToLower(attestationRequest.Url))
		if _, exists := uniqueTokenURLs[url]; exists {
			reqLogger.Error("Duplicate token URL found", "url", attestationRequest.Url)
			metrics.RecordError("duplicate_token_url_found", "attestation_handler")
			httpUtil.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrDecodingRequestBody.WithDetails("Duplicate token URL found"))
			return
		}
		uniqueTokenURLs[url] = true
	}


	if requestsLength == 1 {
		status = processSingleTokenAttestation(ctx, w, attestationRequests[0], reqLogger)
	} else {
		status = processMultipleTokensAttestation(ctx, w, attestationRequests, reqLogger)
	}
}

// processSingleTokenAttestation handles a single token attestation request
func processSingleTokenAttestation(ctx context.Context, w http.ResponseWriter, attestationRequestWithDebug attestation.AttestationRequestWithDebug, reqLogger *slog.Logger) string {
	status := "failed"

	attestationRequest := attestationRequestWithDebug.AttestationRequest.Normalize()

	reqLogger.Debug("Processing single token attestation request", "url", attestationRequest.Url, "debug", attestationRequestWithDebug.DebugRequest)

	if err := attestationRequest.Validate(); err != nil {
		reqLogger.Error("Attestation request validation failed", "error", err)
		metrics.RecordError("validation_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusBadRequest, err)
		return status
	}

	// Get timestamp from roughtime server
	timestamp, err := common.GetTimestampFromRoughtime()
	if err != nil {
		reqLogger.Error("Failed to get timestamp from roughtime server", "error", err)
		metrics.RecordError("timestamp_fetch_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
		return status
	}

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
		return status
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
			ExtractedData:        extractDataResult.AttestationData,
			ResponseStatusCode:   extractDataResult.StatusCode,
		}

		reqLogger.Debug("Debug attestation report generated")
		httpUtil.WriteJsonSuccess(w, http.StatusOK, response)
		return "success"
	}

	// aleoBlockHeight, err := common.GetAleoCurrentBlockHeight()
	// if err != nil {
	// 	reqLogger.Error("Failed to get Aleo block height", "error", err)
	// 	metrics.RecordError("aleo_block_height_fetch_failed", "attestation_handler")
	// 	httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
	// 	return status
	// }

	// Prepare the oracle data before the quote.
	reqLogger.Debug("Preparing data for quote generation")
	quotePrepData, err := attestation.PrepareDataForQuoteGeneration(extractDataResult.StatusCode, extractDataResult.AttestationData, uint64(timestamp), attestationRequest)

	// Check if the error is not nil.
	if err != nil {
		reqLogger.Error("Failed to prepare data for quote generation", "error", err)
		metrics.RecordError("quote_prep_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
		return status
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
		return status
	}

	// metrics.RecordSgxQuoteGeneration("success", quoteDuration)
	reqLogger.Debug("SGX quote generated successfully")

	// Prepare the oracle data after the quote.
	reqLogger.Debug("Building complete oracle data")
	oracleData, err := attestation.BuildCompleteOracleData(quotePrepData, quote)

	if err != nil {
		reqLogger.Error("Failed to build complete oracle data", "error", err)
		metrics.RecordError("oracle_data_build_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
		return status
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
		AttestationResults:   []attestation.AttestationResultForEachToken{
			{
				AttestationData: extractDataResult.AttestationData,
				AtttestationRequest: attestationRequest,
				ResponseBody: extractDataResult.ResponseBody,
				ResponseStatusCode: extractDataResult.StatusCode,
				AttestationTimestamp: timestamp,
				RequestHash: oracleData.RequestHash,
			},
		},
	}

	// Log successful completion
	reqLogger.Debug("Attestation report generated successfully")

	// Write the JSON success response.
	httpUtil.WriteJsonSuccess(w, http.StatusOK, response)
	return "success"
}

// processMultipleTokensAttestation handles multiple tokens attestation request
func processMultipleTokensAttestation(ctx context.Context, w http.ResponseWriter, attestationRequests []attestation.AttestationRequestWithDebug, reqLogger *slog.Logger) string {
	status := "failed"

	reqLogger.Debug("Processing multiple tokens attestation", "count", len(attestationRequests))

	// Get timestamp from roughtime server
	timestamp, err := common.GetTimestampFromRoughtime()
	if err != nil {
		reqLogger.Error("Failed to get timestamp from roughtime server", "error", err)
		metrics.RecordError("timestamp_fetch_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
		return status
	}

	// aleoBlockHeight, err := common.GetAleoCurrentBlockHeight()
	// if err != nil {
	// 	reqLogger.Error("Failed to get Aleo block height", "error", err)
	// 	metrics.RecordError("aleo_block_height_fetch_failed", "attestation_handler")
	// 	httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
	// 	return status
	// }

	normalizedAttestationRequests := make([]attestation.AttestationRequest, len(attestationRequests))

	// Normalize and validate all attestation requests
	for i, attestationRequest := range attestationRequests {
		normalizedAttestationRequest := attestationRequest.AttestationRequest.Normalize()

		if err := attestationRequest.Validate(); err != nil {
			reqLogger.Error("Attestation request validation failed", "index", i, "error", err)
			metrics.RecordError("validation_failed", "attestation_handler")
			httpUtil.WriteJsonError(w, http.StatusBadRequest, err)
			return status
		}
		normalizedAttestationRequests[i] = normalizedAttestationRequest
	}

	// Process all attestation requests in parallel
	type processResult struct {
		index           int
		attestationResult attestation.AttestationResultForEachToken
		err             *appErrors.AppError
		extractDuration float64
	}

	resultChan := make(chan processResult, len(normalizedAttestationRequests))
	var wg sync.WaitGroup

	reqLogger.Debug("Starting parallel processing of attestation requests", "count", len(normalizedAttestationRequests))

	// Launch goroutines for each attestation request
	for i, normalizedAttestationRequest := range normalizedAttestationRequests {
		wg.Add(1)
		go func(idx int, req attestation.AttestationRequest) {
			defer wg.Done()

			reqLogger.Debug("Processing attestation request", "index", idx)

			// Fetch the data from the attestation request.
			extractStart := time.Now()
			extractDataResult, err := data_extraction.ExtractDataFromTargetURL(ctx, req, timestamp)
			extractDuration := time.Since(extractStart).Seconds()

			// Check if the error is not nil.
			if err != nil {
				reqLogger.Error("Failed to extract data from target URL", "index", idx, "error", err, "extractDuration", extractDuration)
				metrics.RecordError("data_extraction_failed", "attestation_handler")
				resultChan <- processResult{
					index:          idx,
					err:            err,
					extractDuration: extractDuration,
				}
				return
			}

			req.MaskUnacceptedHeaders()

			// Prepare the oracle data before the quote.
			reqLogger.Debug("Preparing data for quote generation", "index", idx)

			userDataChunk, encodedPositions, err := attestation.PrepareOracleUserDataChunk(extractDataResult.StatusCode, extractDataResult.AttestationData, uint64(timestamp), req)

			if err != nil {
				reqLogger.Error("Failed to prepare data for quote generation", "index", idx, "error", err)
				metrics.RecordError("quote_prep_failed", "attestation_handler")
				resultChan <- processResult{
					index:          idx,
					err:            err,
					extractDuration: extractDuration,
				}
				return
			}

			requestHash, err := attestation.GetRequestHashFromSingleChunk(userDataChunk, encodedPositions)
			if err != nil {
				reqLogger.Error("Failed to get request hash from single chunk", "index", idx, "error", err)
				metrics.RecordError("request_hash_get_failed", "attestation_handler")
				resultChan <- processResult{
					index:          idx,
					err:            err,
					extractDuration: extractDuration,
				}
				return
			}

			reqLogger.Debug("Successfully processed attestation request", "index", idx, "extractDuration", extractDuration)

			resultChan <- processResult{
				index:          idx,
				err:            nil,
				extractDuration: extractDuration,
				attestationResult: attestation.AttestationResultForEachToken{
					Index: idx,
					UserDataChunk: userDataChunk,
					AttestationData: extractDataResult.AttestationData,
					AtttestationRequest: req,
					ResponseBody: extractDataResult.ResponseBody,
					ResponseStatusCode: extractDataResult.StatusCode,
					AttestationTimestamp: timestamp,
					RequestHash: requestHash,
				},
				}
		}(i, normalizedAttestationRequest)
	}

	// Close the channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results and check for errors
	mergedUserDataChunks := []byte{}
	results := make([]processResult, 0, len(normalizedAttestationRequests))
	hasError := false
	var firstError *appErrors.AppError

	for result := range resultChan {
		results = append(results, result)
		if result.err != nil {
			hasError = true
			if firstError == nil {
				firstError = result.err
			}
		}
	}

	// If any request failed, return error
	if hasError {
		reqLogger.Error("One or more attestation requests failed during parallel processing", "error", firstError)
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, firstError)
		return status
	}

	// Sort results by index to maintain order and merge chunks
	attestationResults := make([]attestation.AttestationResultForEachToken, len(results))
	userDataChunksByIndex := make([][]byte, len(normalizedAttestationRequests))

	for _, result := range results {
		attestationResults[result.index] = result.attestationResult
		userDataChunksByIndex[result.index] = result.attestationResult.UserDataChunk		
	}

	// Merge all user data chunks in order
	for _, chunk := range userDataChunksByIndex {
		mergedUserDataChunks = append(mergedUserDataChunks, chunk...)
	}

	finalMergedUserDataChunks := make([]byte, constants.OracleUserDataChunkSize * constants.ChunkSizeInBytes)
	copy(finalMergedUserDataChunks, mergedUserDataChunks)

	reqLogger.Debug("All attestation requests processed successfully in parallel", "count", len(normalizedAttestationRequests))

	mergedUserData, formatErr := attestation.FormatMessage(finalMergedUserDataChunks, constants.OracleUserDataChunkSize)
	if formatErr != nil {
		reqLogger.Error("Failed to format merged user data", "error", formatErr)
		metrics.RecordError("user_data_format_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, appErrors.ErrInternal)
		return status
	}

	attestationHash, err := attestation.GenerateAttestationHash(mergedUserData)
	if err != nil {
		reqLogger.Error("Failed to generate attestation hash", "error", err)
		metrics.RecordError("attestation_hash_generation_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
		return status
	}

	// Generate the quote.
	reqLogger.Debug("Generating SGX quote")
	quoteStart := time.Now()
	quote, err := sgx.GenerateQuote(attestationHash)
	quoteDuration := time.Since(quoteStart).Seconds()

	if err != nil {
		reqLogger.Error("Failed to generate SGX quote", "error", err)
		metrics.RecordSgxQuoteGeneration("failed", quoteDuration)
		metrics.RecordError("quote_generation_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
		return status
	}

	// metrics.RecordSgxQuoteGeneration("success", quoteDuration)
	reqLogger.Debug("SGX quote generated successfully")

	oracleReport, err := attestation.PrepareOracleReport(quote)
	if err != nil {
		reqLogger.Error("Failed to prepare oracle report", "error", err)
		metrics.RecordError("oracle_report_preparation_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
		return status
	}

	// Prepare the oracle data after the quote.
	reqLogger.Debug("Building complete oracle data")
	signature, publicKey, err := attestation.PrepareOracleSignature(oracleReport)

	if err != nil {
		reqLogger.Error("Failed to build complete oracle data", "error", err)
		metrics.RecordError("oracle_data_build_failed", "attestation_handler")
		httpUtil.WriteJsonError(w, http.StatusInternalServerError, err)
		return status
	}

	reqLogger.Debug("Oracle data built successfully")

	// Create the attestation response.
	response := &attestation.AttestationResponseForMultipleTokens{
		ReportType:           "sgx",
		AttestationTimestamp: timestamp,
		AttestationReport:    base64.StdEncoding.EncodeToString(quote),
		OracleData: attestation.OracleData{
			Signature: signature,
			Report:    string(oracleReport),
			Address:   publicKey,
			UserData:  string(mergedUserData),
		},
		AttestationResults: attestationResults,
	}

	// Log successful completion
	reqLogger.Debug("Attestation report generated successfully for multiple tokens")

	// Write the JSON success response.
	httpUtil.WriteJsonSuccess(w, http.StatusOK, response)
	return "success"
}