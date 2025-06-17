package services

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	encoding "github.com/zkportal/aleo-oracle-encoding"
)

type AttestationRequest struct {
	Url string `json:"url"`

	RequestMethod  string  `json:"requestMethod" validate:"required"`
	Selector       string  `json:"selector,omitempty" validate:"required"`
	ResponseFormat string  `json:"responseFormat" validate:"required"`
	HTMLResultType *string `json:"htmlResultType,omitempty" validate:"required"`

	RequestBody        *string `json:"requestBody,omitempty"`
	RequestContentType *string `json:"requestContentType,omitempty" validate:"required"`

	RequestHeaders map[string]string `json:"requestHeaders,omitempty"`

	EncodingOptions encoding.EncodingOptions `json:"encodingOptions" validate:"required"`

	DebugRequest bool `json:"debugRequest,omitempty"`
}

type AttestationResponse struct {
	ReportType string `json:"reportType"`

	AttestationRequest AttestationRequest `json:"attestationRequest"`

	AttestationReport string `json:"attestationReport"`

	AttestationTimestamp int64 `json:"timestamp"`

	ResponseBody string `json:"responseBody"`

	ResponseStatusCode int `json:"responseStatusCode"`

	AttestationData string `json:"attestationData"`

	OracleData OracleData `json:"oracleData"`
}

func (ar *AttestationRequest) Validate() error {

	if ar.Url == ""{
		return appErrors.ErrMissingURL
	}

	if ar.RequestMethod == "" {
		return appErrors.ErrMissingRequestMethod
	}

	if ar.Selector == "" {
		return appErrors.ErrMissingSelector
	}

	if ar.EncodingOptions.Value == ""{
		return appErrors.ErrMissingEncodingOption
	}

	if ar.RequestMethod != http.MethodGet && ar.RequestMethod != http.MethodPost{
		return appErrors.ErrInvalidRequestMethod
	}

	if ar.RequestMethod == http.MethodPost && ar.RequestBody == nil{
		return appErrors.ErrMissingRequestBody
	}

	if ar.ResponseFormat != "html" && ar.ResponseFormat != "json" {
		return appErrors.ErrInvalidResponseFormat
	}

	if ar.ResponseFormat  == "html" && ar.HTMLResultType == nil {
		return appErrors.ErrMissingHTMLResultType
	}

	if ar.ResponseFormat  == "html" && *ar.HTMLResultType != "value" && *ar.HTMLResultType != "element" {
		return appErrors.ErrInvalidHTMLResultType
	}

	if ar.EncodingOptions.Value != "string" && ar.EncodingOptions.Value != "float" && ar.EncodingOptions.Value != "integer"{
		return appErrors.ErrInvalidEncodingOption
	}

	if !utils.IsAcceptedDomain(ar.Url) {
			return appErrors.ErrUnacceptedDomain
	}

	return nil
}

func wrapRawQuoteAsOpenEnclaveEvidence(rawQuoteBuffer []byte) ([]byte, error) {

	oeVersion := make([]byte, 4)
	binary.LittleEndian.PutUint32(oeVersion, 1)

	oeType := make([]byte, 4)
	binary.LittleEndian.PutUint32(oeType, 2)

	quoteLength := make([]byte, 8)
	binary.LittleEndian.PutUint32(quoteLength, uint32(len(rawQuoteBuffer)))

	var buf bytes.Buffer

	buf.Write(oeVersion)
	buf.Write(oeType)
	buf.Write(quoteLength)
	buf.Write(rawQuoteBuffer)

	return buf.Bytes(), nil
}

// dev/attestation/user_report_data -write
// dev/attestation/quote - read

var quoteLock sync.Mutex

func GenerateQuote(inputData []byte) ([]byte, error){

	quoteLock.Lock()
	defer quoteLock.Unlock()

	reportData := make([]byte, 64)

	copy(reportData,inputData)

	fmt.Printf("64-byte report data: %x\n", reportData)

	err := os.WriteFile(constants.GRAMINE_PATHS.USER_REPORT_DATA_PATH,reportData,0644)

	if err != nil {
		log.Print("Error while writting report data:",err)
		return nil,appErrors.ErrWrittingReportData
	}

	quote, err := os.ReadFile(constants.GRAMINE_PATHS.QUOTE_PATH)

	if err != nil {
		log.Print("Generate Quote err: ",err)
		return nil,appErrors.ErrReadingQuote
	}

	finalQuote, err := wrapRawQuoteAsOpenEnclaveEvidence(quote)

	if err != nil {
		return nil,appErrors.ErrWrappingQuote
	}

	return finalQuote, nil
}