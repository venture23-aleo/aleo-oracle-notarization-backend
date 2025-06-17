package handlers

import (
	"log"
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services"

	aleo "github.com/zkportal/aleo-utils-go"
)

func GetInfo(s aleo.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		enclaveInfo, err := services.GetSgxInfo(s)

		requestId := utils.GenerateShortRequestID()

		if err != nil {
			log.Print("Error getting enclave info:", err)
			w.WriteHeader(http.StatusInternalServerError)
			utils.WriteJsonError(w, http.StatusInternalServerError, err.(appErrors.AppError), requestId)
			return
		}

		instanceInfo := services.InstanceInfo{
			ReportType:   "sgx",
			Info:         enclaveInfo,
			SignerPubKey: configs.PublicKey,
		}

		utils.WriteJsonSuccess(w, http.StatusOK, instanceInfo)
	}
}
