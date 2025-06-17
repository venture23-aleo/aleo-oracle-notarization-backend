package api

import (
	"net/http"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/api/handlers"

	aleo "github.com/zkportal/aleo-utils-go"
)

func RegisterRoutes(mux *http.ServeMux, s aleo.Session ){
	mux.HandleFunc("/",handlers.GetHealthCheck)
	mux.HandleFunc("/notarize",handlers.GenerateAttestationReportHandler(s))
	mux.HandleFunc("/info",handlers.GetInfo(s))
	mux.HandleFunc("/whitelist",handlers.GetWhiteListedDomains)
}