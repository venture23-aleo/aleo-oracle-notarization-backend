package attestation

import (

	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	aleoContext "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/aleo_context"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/logger"
)

// OracleData is the data of the oracle.
type OracleData struct {
	// Schnorr signature of a verified Attestation Report.
	Signature string `json:"signature"`

	// Aleo-encoded data that was used to create the hash included in the Attestation Report.
	//
	// See ProofPositionalInfo for an idea of what data goes into the hash.
	UserData string `json:"userData"`

	// Aleo-encoded Attestation Report.
	Report string `json:"report"`

	// Public key the signature was created against.
	Address string `json:"address"`

	// Object containing information about the positions of data included in the Attestation Report hash.
	EncodedPositions encoding.ProofPositionalInfo `json:"encodedPositions"`

	// Aleo-encoded request. Same as UserData but with zeroed Data and Timestamp fields. Can be used to validate the request in Aleo programs.
	//
	// Data and Timestamp are the only parts of UserData that can be different every time you do a notarization request.
	// By zeroing out these 2 fields, we can create a constant UserData which is going to represent a request to the attestation target.
	// When an Aleo program is going to verify that a request was done using the correct parameters, like URL, request body, request headers etc.,
	// it can take the UserData provided with the Attestation Report, replace Data and Timestamp with "0u128" and then compare the result with the constant UserData in the program.
	// If both UserDatas match, then we know that the Attestation Report was made using the correct attestation target request!
	//
	// To avoid storing the full UserData in an Aleo program, we can hash it and store only the hash in the program. See RequestHash.
	EncodedRequest string `json:"encodedRequest"`

	// Poseidon8 hash of the EncodedRequest. Can be used to verify in an Aleo program that the report was made with the correct request.
	RequestHash string `json:"requestHash"`

	// Poseidon8 hash of the RequestHash with the attestation timestamp. Can be used to verify in an Aleo program that the report was made with the correct request.
	TimestampedRequestHash string `json:"timestampedRequestHash"`

	// Object containing extra information about the attestation report.
	// If the attestation type is "nitro", it contains Aleo-encoded structs with
	// information that helps to extract user data and PCR values from the report.
	// ReportExtras *NitroReportExtras `json:"reportExtras"`
}

// PrepareOracleReport prepares the oracle report from the quote
func PrepareOracleReport(quote []byte) (oracleReport []byte, appError *appErrors.AppError) {
	aleoContext, err := aleoContext.GetAleoContext()
	if err != nil {
		logger.Error("Error getting Aleo context: ", "error", err)
		return nil, err
	}
	
	// C0 - C9 Chunks
	oracleReport, formatErr := aleoContext.GetSession().FormatMessage(quote, 10)

	if formatErr != nil {
		logger.Error("Failed to format quote: ", "error", formatErr)
		return nil, appErrors.NewAppError(appErrors.ErrFormattingQuote)
	}

	return oracleReport, nil
}

// PrepareOracleSignature prepares the oracle signature from the oracle report
func PrepareOracleSignature(oracleReport []byte) (signature string, appError *appErrors.AppError) {
	aleoContext, err := aleoContext.GetAleoContext()
	if err != nil {
		logger.Error("Error getting Aleo context: ", "error", err)
		return "", err
	}
	
	// Create the hashed message.
	hashedMessage, hashErr := aleoContext.GetSession().HashMessage(oracleReport)

	if hashErr != nil {
		logger.Error("Failed to hash report", "error", hashErr)
		return "", appErrors.NewAppError(appErrors.ErrReportHashing)
	}

	// Create the signature.
	signature, signErr := aleoContext.Sign(hashedMessage)

	// Check if the error is not nil.
	if signErr != nil {
		logger.Error("Error while generating signature: ", "error", signErr)
		return "", appErrors.NewAppError(appErrors.ErrGeneratingSignature)
	}

	return signature, nil
}

// GenerateAttestationHash generates the attestation hash from user data
func GenerateAttestationHash(userData []byte) (attestationHash []byte, err *appErrors.AppError) {
	aleoContext, err := aleoContext.GetAleoContext()
	if err != nil {
		logger.Error("Error getting Aleo context: ", "error", err)
		return nil, err
	}
	
	attestationHash, hashError := aleoContext.GetSession().HashMessage(userData)

	// Check if the error is not nil.
	if hashError != nil {
		logger.Error("Failed to hash message: ", "error", hashError)
		return nil, appErrors.NewAppError(appErrors.ErrMessageHashing)
	}

	return attestationHash, nil
}

// BuildCompleteOracleData builds the complete oracle data structure
func BuildCompleteOracleData(quotePrepData *QuotePreparationData, quote []byte) (*OracleData, *appErrors.AppError) {
	aleoContext, err := aleoContext.GetAleoContext()
	if err != nil {
		logger.Error("Error getting Aleo context: ", "error", err)
		return nil, err
	}
	
	encodedRequest, err := PrepareOracleEncodedRequest(quotePrepData.UserDataProof, quotePrepData.EncodedPositions)

	if err != nil {
		logger.Error("Failed to prepare oracle encoded request: ", "error", err)
		return nil, err
	}

	requestHash, requestHashString, err := PrepareOracleRequestHash(encodedRequest)

	if err != nil {
		logger.Error("Failed to prepare oracle request hash: ", "error", err)
		return nil, err
	}

	timestampedHash, err := PrepareOracleTimestampedRequestHash(requestHash, quotePrepData.Timestamp)

	if err != nil {
		logger.Error("Failed to prepare oracle timestamped request hash: ", "error", err)
		return nil, err
	}

	oracleReport, err := PrepareOracleReport(quote)

	if err != nil {
		logger.Error("Failed to prepare oracle report: ", "error", err)
		return nil, err
	}

	signature, err := PrepareOracleSignature(oracleReport)

	if err != nil {
		logger.Error("Failed to prepare oracle signature: ", "error", err)
		return nil, err
	}

	oracleData := &OracleData{
		UserData:               string(quotePrepData.UserData),
		EncodedPositions:       *quotePrepData.EncodedPositions,
		EncodedRequest:         string(encodedRequest),
		RequestHash:            requestHashString,
		TimestampedRequestHash: timestampedHash,
		Report:                 string(oracleReport),
		Signature:              signature,
		Address:                aleoContext.GetPublicKey(),
	}

	return oracleData, nil
} 