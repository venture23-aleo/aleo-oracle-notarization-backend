package services

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"

	aleo "github.com/zkportal/aleo-utils-go"
)

type SgxAleoInfo struct {
	UniqueID  string `json:"uniqueId"`  // Same as UniqueID but encoded for Aleo as 2 uint128
	SignerID  string `json:"signerId"`  // Same as SignerID but encoded for Aleo as 2 uint128
	ProductID string `json:"productId"` // Same as ProductID but encoded for Aleo as 1 uint128
}

type SgxInfo struct {
	SecurityVersion uint16        `json:"securityVersion"` // Security version of the enclave. For SGX enclaves, this is the ISVSVN value.
	Debug           bool        `json:"debug"`           // If true, the report is for a debug enclave.
	UniqueID        string      `json:"uniqueId"`        // The unique ID for the enclave. For SGX enclaves, this is the MRENCLAVE value.
	SignerID        string      `json:"signerId"`        // The signer ID for the enclave. For SGX enclaves, this is the MRSIGNER value.
	ProductID       string      `json:"productId"`       // The Product ID for the enclave. For SGX enclaves, this is the ISVPRODID value.
	Aleo            SgxAleoInfo `json:"aleo"`            // Some of the SGX report values encoded for Aleo.
	TCBStatus       uint        `json:"tcbStatus"`       // The status of the enclave's TCB level.
}

type InstanceInfo struct {
	ReportType   string `json:"reportType"`
	Info         SgxInfo   `json:"info"`
	SignerPubKey string `json:"signerPubKey"`
}

func GetSgxInfo(s aleo.Session) (SgxInfo, error){

	targetInfo, err := os.ReadFile(constants.GRAMINE_PATHS.MY_TARGET_INFO_PATH)

	if err != nil {
		log.Print("Error reading target info",err)
		return SgxInfo{},appErrors.ErrReadingTargetInfo
	}

	err = os.WriteFile(constants.GRAMINE_PATHS.TARGET_INFO_PATH, targetInfo, 0644);

	if err != nil {
		log.Print("Error writting target info",err)
		return SgxInfo{},appErrors.ErrWrittingTargetInfo
	}

	reportData := make([]byte, 64)

	err = os.WriteFile(constants.GRAMINE_PATHS.USER_REPORT_DATA_PATH,reportData,0644)

	if err != nil {
		log.Print("Error while writting report data:",err)
		return SgxInfo{},appErrors.ErrWrittingReportData
	}

	report, err := os.ReadFile(constants.GRAMINE_PATHS.REPORT_PATH)

	/*
            | Field       | Offset | Size (bytes) |
            | ----------- | ------ | ------------ |
            | FLAGS       | 48     | 8            |
            | XFRM        | 56     | 8            |
            | MRENCLAVE   | 64     | 32           |
            | MRSIGNER    | 128    | 32           |
            | ISVPRODID   | 256    | 2            |
            | ISVSVN      | 258    | 2            |
            | REPORT DATA | 320    | 64           |
    */

	debug := (report[48]&0x02) > 0
	mrEnclave := report[64:96]
	mrSigner := report[128:160]
	prodId := report[256:258]
	secVersion := report[258:260]

	// mrEnclave, err = hex.DecodeString("f471eb7d442521bced625d420d12d32e795106c18204f1fe08f91ad81c4b0f79")
	
	rawProdID := make([]byte, 16)
	copy(rawProdID, prodId)

	mrEnclaveChunk1 := new(big.Int).SetBytes(utils.ReverseBytes(mrEnclave[:len(mrEnclave)/2]))
	mrEnclaveChunk2 := new(big.Int).SetBytes(utils.ReverseBytes(mrEnclave[len(mrEnclave)/2:]))

	mrSignerChunk1 := new(big.Int).SetBytes(utils.ReverseBytes(mrSigner[:len(mrSigner)/2]))
	mrSignerChunk2 := new(big.Int).SetBytes(utils.ReverseBytes(mrSigner[len(mrSigner)/2:]))

	sgxInfo := SgxInfo{
		UniqueID: base64.StdEncoding.EncodeToString(mrEnclave),
		SignerID: base64.StdEncoding.EncodeToString(mrSigner),
		ProductID: base64.StdEncoding.EncodeToString(rawProdID),
		SecurityVersion: binary.LittleEndian.Uint16(secVersion),
		Debug: debug,
		TCBStatus: 5, // TODO: need to handle this
		Aleo: SgxAleoInfo{
			UniqueID: fmt.Sprintf("{ chunk_1: %su128, chunk_2: %su128 }",mrEnclaveChunk1, mrEnclaveChunk2),
			SignerID: fmt.Sprintf("{ chunk_1: %su128, chunk_2: %su128 }",mrSignerChunk1, mrSignerChunk2),
			ProductID: fmt.Sprintf("%du128",binary.LittleEndian.Uint16(prodId)),
		},
	}

	return sgxInfo, nil

}