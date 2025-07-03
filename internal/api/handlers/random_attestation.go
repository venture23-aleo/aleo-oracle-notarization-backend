package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"time"

	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	attestation "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
)

// GenerateAttestedRandom handles the request to generate an attested random number
func GenerateAttestedRandom(w http.ResponseWriter, req *http.Request) {
	// Generate a short request ID
	requestId := utils.GenerateShortRequestID()

	log.Printf("[INFO] [%s] Received attested random request: %s", requestId, req.URL.RawQuery)

	// Get the max parameter from query string
	maxStr := req.URL.Query().Get("max")
	if maxStr == "" {
		log.Printf("[ERROR] [%s] Missing max parameter", requestId)
		utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, requestId)
		return
	}

	// Parse the max value as a big integer
	max, ok := new(big.Int).SetString(maxStr, 10)
	if !ok {
		log.Printf("[ERROR] [%s] Invalid max parameter format: %s", requestId, maxStr)
		utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, requestId)
		return
	}

	// Validate the max value: must be > 1 and ≤ 2^127
	if max.Cmp(big.NewInt(1)) <= 0 {
		log.Printf("[ERROR] [%s] max must be greater than 1: %s", requestId, maxStr)
		utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, requestId)
		return
	}

	// Check if max > 2^127
	maxAllowed := new(big.Int).Lsh(big.NewInt(1), 127) // 2^127
	if max.Cmp(maxAllowed) > 0 {
		log.Printf("[ERROR] [%s] max must be less than or equal to 2^127: %s", requestId, maxStr)
		utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, requestId)
		return
	}

	log.Printf("[INFO] [%s] Validated max parameter: %s", requestId, maxStr)

	attestationRequest := attestation.AttestationRequest{
		Url: fmt.Sprintf("crypto/rand:%s", max.String()),
		RequestMethod: "GET",
		ResponseFormat: "json",
		EncodingOptions: encoding.EncodingOptions{
			Value: "int",
			Precision: 0,
		},
	}

	// Generate a cryptographically secure random number in range [0, max)
	randomNumber, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Printf("[ERROR] [%s] Error generating random number: %v", requestId, err)
		utils.WriteJsonError(w, http.StatusInternalServerError, appErrors.ErrGeneratingAttestationHash, requestId)
		return
	}

	log.Printf("[INFO] [%s] Generated random number: %s", requestId, randomNumber.String())

	// Get the timestamp
	timestamp := time.Now().Unix()

	statusCode := 200
	attestationData := randomNumber.String()

	log.Printf("[INFO] [%s] Preparing data for quote generation", requestId)
	quotePrepData, appError := attestation.PrepareDataForQuoteGeneration(statusCode, attestationData, uint64(timestamp), attestationRequest)

	// Check if the error is not nil.
	if appError != nil {
		log.Printf("[ERROR] [%s] Error preparing data for quote generation: %v", requestId, appError)
		utils.WriteJsonError(w, http.StatusBadRequest, *appError , requestId)
		return
	}

	log.Printf("[INFO] [%s] Generating SGX quote", requestId)

	// Generate the quote.
	quote, appError := attestation.GenerateQuote(quotePrepData.AttestationHash)
	if appError != nil {
		log.Printf("[ERROR] [%s] Error generating quote: %v", requestId, appError)
		utils.WriteJsonError(w, http.StatusInternalServerError, *appError, requestId)
		return
	}

	log.Printf("[INFO] [%s] Building oracle data after quote", requestId)

	// Build the complete oracle data after the quote.
	oracleData, appError := attestation.BuildCompleteOracleData(quotePrepData, quote)
	if appError != nil {
		log.Printf("[ERROR] [%s] Error preparing oracle data after quote: %v", requestId, appError)
		utils.WriteJsonError(w, http.StatusInternalServerError, *appError, requestId)
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

	log.Printf("[INFO] [%s] Successfully generated attested random response", requestId)
	// Write the JSON success response
	utils.WriteJsonSuccess(w, http.StatusOK, response)
}

