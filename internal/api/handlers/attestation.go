package handlers

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	aleo "github.com/zkportal/aleo-utils-go"
)

// GenerateAttestationReportHandler handles the request to generate an attestation report.
func GenerateAttestationReportHandler(s aleo.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		// Check if the request method is not POST.
		if req.Method != http.MethodPost {
			// Write the status code.
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Check if the content type is not JSON.
		if req.Header.Get("Content-Type") != "application/json" {
			// Write the status code.
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Generate a short request ID.
		requestId := utils.GenerateShortRequestID()

		// Close the request body.
		defer req.Body.Close()

		// Create the attestation request.
		var attestationRequest services.AttestationRequest

		// Decode the request body.
		if err := json.NewDecoder(req.Body).Decode(&attestationRequest); err != nil {
			// Log the error.
			log.Println("error reading request:", err)

			// Write the JSON error response.
			utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, requestId)
			return
		}

		// Validate the attestation request.
		err := attestationRequest.Validate()

		// Check if the error is not nil.
		if err != nil {
			// Write the JSON error response.
			utils.WriteJsonError(w, http.StatusBadRequest, err.(appErrors.AppError), requestId)
			return
		}

		// Get the timestamp.
		timestamp := time.Now().Unix()

		// Fetch the data from the attestation request.
		responseBody, attestationData, statusCode, err := services.ExtractDataFromTargetURL(attestationRequest)

		// Check if the error is not nil.
		if err != nil {
			// Write the JSON error response.
			utils.WriteJsonError(w, http.StatusInternalServerError, err.(appErrors.AppError), requestId)
			return
		}

		// Mask the unaccepted headers.
		maskedHeaders := utils.MaskUnacceptedHeaders(attestationRequest.RequestHeaders)
		attestationRequest.RequestHeaders = maskedHeaders

		// Prepare the oracle data before the quote.
		oracleData, err := services.PrepareOracleDataBeforeQuote(s, statusCode, attestationData, uint64(timestamp), services.AttestationRequest(attestationRequest))

		// Check if the error is not nil.
		if err != nil {
			// Write the JSON error response.
			utils.WriteJsonError(w, http.StatusBadRequest, err.(appErrors.AppError), requestId)
			return
		}

		// Hash the oracle data.
		attestationHash, err := s.HashMessage([]byte(oracleData.UserData))

		// Check if the error is not nil.
		if err != nil {
			// Write the JSON error response.
			utils.WriteJsonError(w, http.StatusInternalServerError, appErrors.ErrMessageHashing, requestId)
			return
		}

		// Log the attestation hash.
		log.Printf("Attestation hash: %v", hex.EncodeToString(attestationHash))

		// Generate the quote.
		quote, err := services.GenerateQuote(attestationHash)

		// Check if the error is not nil.
		if err != nil {
			// Log the error.
			log.Println("error generating quote", err)

			// Write the JSON error response.
			utils.WriteJsonError(w, http.StatusInternalServerError, err.(appErrors.AppError), requestId)
			return
		}

		// Prepare the oracle data after the quote.
		oracleData, err = services.PrepareOracleDataAfterQuote(s, oracleData, quote)

		// Check if the error is not nil.
		if err != nil {
			// Log the error.
			log.Println("error preparing oracle data after quote", err)

			// Write the JSON error response.
			utils.WriteJsonError(w, http.StatusInternalServerError, err.(appErrors.AppError), requestId)
			return
		}

		// Create the attestation response.
		response := &services.AttestationResponse{
			ReportType:           "sgx",
			AttestationRequest:   attestationRequest,
			AttestationTimestamp: timestamp,
			ResponseBody:         responseBody,
			AttestationData:      attestationData,
			ResponseStatusCode:   statusCode,
			AttestationReport:    base64.StdEncoding.EncodeToString(quote),
			OracleData:           oracleData,
		}

		// Write the JSON success response.
		utils.WriteJsonSuccess(w, http.StatusOK, response)
	}
}
