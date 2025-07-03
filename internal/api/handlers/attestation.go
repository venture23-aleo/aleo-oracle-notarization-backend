package handlers

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/data_extraction"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
)

// GenerateAttestationReport handles the request to generate an attestation report.
func GenerateAttestationReport(w http.ResponseWriter, req *http.Request) {
	// Generate a short request ID for tracing
	requestId := utils.GenerateShortRequestID()
	
	// Log the incoming request
	log.Printf("[%s] POST /attestation - Attestation report request received", requestId)

	// Validate Content-Type
	if req.Header.Get("Content-Type") != "application/json" {
		log.Printf("[%s] ERROR: Invalid Content-Type: %s for %s %s", 
			requestId, req.Header.Get("Content-Type"), req.Method, req.URL.Path)
		utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, requestId)
		return
	}

	// Close the request body.
	defer req.Body.Close()

	// Create the attestation request.
	var attestationRequestWithDebug attestation.AttestationRequestWithDebug

	// Decode the request body.
	if err := json.NewDecoder(req.Body).Decode(&attestationRequestWithDebug); err != nil {
		log.Printf("[%s] ERROR: Failed to decode request body: %v", requestId, err)
		utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, requestId)
		return
	}

	attestationRequest := attestationRequestWithDebug.AttestationRequest

	// Log request details for debugging
	log.Printf("[%s] INFO: Processing attestation request for URL: %s, Debug: %t", 
		requestId, attestationRequest.Url, attestationRequestWithDebug.DebugRequest)

	// Validate the attestation request.
	if err := attestationRequest.Validate(); err != nil {
		log.Printf("[%s] ERROR: Attestation request validation failed: %v", requestId, err)
		utils.WriteJsonError(w, http.StatusBadRequest, *err, requestId)
		return
	}

	// Get the timestamp.
	timestamp := time.Now().Unix()
	log.Printf("[%s] INFO: Using timestamp: %d", requestId, timestamp)

	// Fetch the data from the attestation request.
	log.Printf("[%s] INFO: Fetching data from target URL: %s", requestId, attestationRequest.Url)
	extractDataResult, err := data_extraction.ExtractDataFromTargetURL(attestationRequest)

	// Check if the error is not nil.
	if err != nil {
		log.Printf("[%s] ERROR: Failed to extract data from target URL: %v", requestId, err)
		utils.WriteJsonError(w, http.StatusInternalServerError, *err, requestId)
		return
	}

	log.Printf("[%s] INFO: Successfully extracted data - Status: %d, Data: %v", 
		requestId, extractDataResult.StatusCode, extractDataResult.AttestationData)

	// Mask the unaccepted headers.
	attestationRequest.MaskUnacceptedHeaders()

	if attestationRequestWithDebug.DebugRequest {
		log.Printf("[%s] INFO: Returning debug response", requestId)
		
		// Create the attestation response.
		response := &attestation.DebugAttestationResponse{
			ReportType:           "sgx",
			AttestationRequest:   attestationRequest,
			AttestationTimestamp: timestamp,
			ResponseBody:         extractDataResult.ResponseBody,
			AttestationData:      extractDataResult.AttestationData,
			ResponseStatusCode:   extractDataResult.StatusCode,
		}

		log.Printf("[%s] SUCCESS: Debug attestation report generated", requestId)
		utils.WriteJsonSuccess(w, http.StatusOK, response)
		return
	}

	// Prepare the oracle data before the quote.
	log.Printf("[%s] INFO: Preparing data for quote generation", requestId)
	quotePrepData, err := attestation.PrepareDataForQuoteGeneration(extractDataResult.StatusCode, extractDataResult.AttestationData, uint64(timestamp), attestationRequest)

	// Check if the error is not nil.
	if err != nil {
		log.Printf("[%s] ERROR: Failed to prepare data for quote generation: %v", requestId, err)
		utils.WriteJsonError(w, http.StatusBadRequest, *err, requestId)
		return
	}

	log.Printf("[%s] INFO: Quote preparation successful", requestId)

	// Generate the quote.
	log.Printf("[%s] INFO: Generating SGX quote", requestId)
	quote, err := attestation.GenerateQuote(quotePrepData.AttestationHash)
	
	if err != nil {
		log.Printf("[%s] ERROR: Failed to generate SGX quote: %v", requestId, err)
		utils.WriteJsonError(w, http.StatusInternalServerError, *err, requestId)
		return
	}

	log.Printf("[%s] INFO: SGX quote generated successfully - Length: %d bytes", requestId, len(quote))

	// Prepare the oracle data after the quote.
	log.Printf("[%s] INFO: Building complete oracle data", requestId)
	oracleData, err := attestation.BuildCompleteOracleData(quotePrepData, quote)
	
	if err != nil {
		log.Printf("[%s] ERROR: Failed to build complete oracle data: %v", requestId, err)
		utils.WriteJsonError(w, http.StatusInternalServerError, *err, requestId)
		return
	}

	log.Printf("[%s] INFO: Oracle data built successfully", requestId)

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
	log.Printf("[%s] SUCCESS: Attestation report generated successfully - Quote length: %d bytes, Oracle data ready", requestId, len(quote))

	// Write the JSON success response.
	utils.WriteJsonSuccess(w, http.StatusOK, response)
}

