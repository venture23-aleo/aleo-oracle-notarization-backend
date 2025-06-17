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

func GenerateAttestationReportHandler(s aleo.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if req.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		requestId := utils.GenerateShortRequestID()
	
		timestamp := time.Now().Unix()

		defer req.Body.Close()

		var attestationRequest services.AttestationRequest

		if err := json.NewDecoder(req.Body).Decode(&attestationRequest); err != nil {
			log.Println("error reading request:", err)
			utils.WriteJsonError(w, http.StatusBadRequest, appErrors.ErrInvalidRequestData, requestId)
			return
		}

		err := attestationRequest.Validate()

		if err != nil {
			utils.WriteJsonError(w, http.StatusBadRequest, err.(appErrors.AppError), requestId)
			return
		}

		responseBody, attestationData, statusCode, err := services.FetchDataFromAttestationRequest(attestationRequest)

		if err != nil {
			utils.WriteJsonError(w, http.StatusInternalServerError, err.(appErrors.AppError),  requestId)
			return
		}

		maskedHeaders := utils.MaskUnacceptedHeaders(attestationRequest.RequestHeaders)
		attestationRequest.RequestHeaders = maskedHeaders

		oracleData, err := services.PrepareOracleDataBeforeQuote(s, statusCode, attestationData, uint64(timestamp), services.AttestationRequest(attestationRequest))

		if err != nil {
			utils.WriteJsonError(w, http.StatusBadRequest, err.(appErrors.AppError), requestId)
			return
		}

		attestationHash, err := s.HashMessage([]byte(oracleData.UserData))

		if err != nil {
			utils.WriteJsonError(w, http.StatusInternalServerError, appErrors.ErrMessageHashing,  requestId)
			return
		}

		log.Printf("Attestation hash: %v", hex.EncodeToString(attestationHash))

		quote, err := services.GenerateQuote(attestationHash)

		if err != nil {
			log.Println("error generating quote", err)
			utils.WriteJsonError(w, http.StatusInternalServerError, err.(appErrors.AppError), requestId)
			return
		}

		oracleData,err = services.PrepareOracleDataAfterQuote(s, oracleData, quote)

		if err != nil {
			log.Println("error preparing oracle data after quote", err)
			utils.WriteJsonError(w, http.StatusInternalServerError, err.(appErrors.AppError), requestId)
			return
		}

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

		utils.WriteJsonSuccess(w, http.StatusOK, response)
	}
}
