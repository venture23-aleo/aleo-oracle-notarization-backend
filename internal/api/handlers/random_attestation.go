package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"time"

	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	attestation "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/attestation"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
)

// GenerateAttestedRandom handles the request to generate an attested random number
func GenerateAttestedRandom(w http.ResponseWriter, req *http.Request) {
	// Get logger from context (request ID automatically included by middleware)
	reqLogger := logger.FromContext(req.Context())

	reqLogger.Debug("Received attested random request", "query", req.URL.RawQuery)

	// Get the max parameter from query string
	maxStr := req.URL.Query().Get("max")
	if maxStr == "" {
		reqLogger.Error("Missing max parameter")
		utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, "")
		return
	}

	// Parse the max value as a big integer
	max, ok := new(big.Int).SetString(maxStr, 10)
	if !ok {
		reqLogger.Error("Invalid max parameter format", "max", maxStr)
		utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, "")
		return
	}

	// Validate the max value: must be > 1 and ≤ 2^127
	if max.Cmp(big.NewInt(1)) <= 0 {
		reqLogger.Error("Max must be greater than 1", "max", maxStr)
		utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, "")
		return
	}

	// Check if max > 2^127
	maxAllowed := new(big.Int).Lsh(big.NewInt(1), 127) // 2^127
	if max.Cmp(maxAllowed) > 0 {
		reqLogger.Error("Max must be less than or equal to 2^127", "max", maxStr)
		utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, "")
		return
	}

	reqLogger.Debug("Validated max parameter", "max", maxStr)

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
		reqLogger.Error("Error generating random number", "error", err)
		utils.WriteJsonError(w, http.StatusInternalServerError, appErrors.ErrGeneratingAttestationHash, "")
		return
	}

	reqLogger.Debug("Generated random number", "number", randomNumber.String())

	// Get the timestamp
	timestamp := time.Now().Unix()

	statusCode := 200
	attestationData := randomNumber.String()

	reqLogger.Debug("Preparing data for quote generation")
	quotePrepData, appError := attestation.PrepareDataForQuoteGeneration(statusCode, attestationData, uint64(timestamp), attestationRequest)

	// Check if the error is not nil.
	if appError != nil {
		reqLogger.Error("Error preparing data for quote generation", "error", appError)
		utils.WriteJsonError(w, http.StatusBadRequest, *appError, "")
		return
	}

	reqLogger.Debug("Generating SGX quote")

	// Generate the quote.
	quote, appError := attestation.GenerateQuote(quotePrepData.AttestationHash)
	if appError != nil {
		reqLogger.Error("Error generating quote", "error", appError)
		utils.WriteJsonError(w, http.StatusInternalServerError, *appError, "")
		return
	}

	reqLogger.Debug("Building oracle data after quote")

	// Build the complete oracle data after the quote.
	oracleData, appError := attestation.BuildCompleteOracleData(quotePrepData, quote)
	if appError != nil {
		reqLogger.Error("Error preparing oracle data after quote", "error", appError)
		utils.WriteJsonError(w, http.StatusInternalServerError, *appError, "")
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
	// Write the JSON success response
	utils.WriteJsonSuccess(w, http.StatusOK, response)
}

