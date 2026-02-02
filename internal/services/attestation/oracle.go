package attestation

import (
	"crypto/rand"
	"math/big"
	"time"

	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	aleoUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/aleoutil"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
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
	// To omit this field when empty, use a pointer type so omitempty works as intended.
	EncodedPositions *encoding.ProofPositionalInfo `json:"encodedPositions,omitempty"`


	// Aleo-encoded request. Same as UserData but with zeroed Data and Timestamp fields. Can be used to validate the request in Aleo programs.
	//
	// Data and Timestamp are the only parts of UserData that can be different every time you do a notarization request.
	// By zeroing out these 2 fields, we can create a constant UserData which is going to represent a request to the attestation target.
	// When an Aleo program is going to verify that a request was done using the correct parameters, like URL, request body, request headers etc.,
	// it can take the UserData provided with the Attestation Report, replace Data and Timestamp with "0u128" and then compare the result with the constant UserData in the program.
	// If both UserDatas match, then we know that the Attestation Report was made using the correct attestation target request!
	//
	// To avoid storing the full UserData in an Aleo program, we can hash it and store only the hash in the program. See RequestHash.
	EncodedRequest string `json:"encodedRequest,omitempty"`

	// Poseidon8 hash of the EncodedRequest. Can be used to verify in an Aleo program that the report was made with the correct request.
	RequestHash string `json:"requestHash,omitempty"`

	// Poseidon8 hash of the RequestHash with the attestation timestamp. Can be used to verify in an Aleo program that the report was made with the correct request.
	TimestampedRequestHash string `json:"timestampedRequestHash,omitempty"`
}

// PrepareOracleReport prepares the oracle report from the provided quote.
//
// This function takes a byte slice representing the quote and processes it to generate
// the oracle report in a specific format. The process involves the following steps:
//
//  1. Retrieve the Aleo context required for formatting the message. If this fails, an error is logged and returned.
//  2. Use the Aleo session to format the quote into 10 chunks (C0 - C9) using the FormatMessage method.
//     If formatting fails, an error is logged and a formatting error is returned.
//  3. If successful, the formatted oracle report is returned.
//
// Parameters:
//   - quote ([]byte): The input quote to be formatted.
//
// Returns:
//   - oracleReport ([]byte): The formatted oracle report.
//   - appError (*appErrors.AppError): An application error, if any occurred during processing.
func PrepareOracleReport(quote []byte) (oracleReport []byte, appError *appErrors.AppError) {
	// Step 1: Retrieve Aleo context
	aleoContext, err := aleoUtil.GetAleoContext()
	if err != nil {
		logger.Error("Error getting Aleo context: ", "error", err)
		return nil, err
	}

	// Step 2: Format the quote into 10 chunks (C0 - C9)
	oracleReport, formatErr := aleoContext.FormatMessage(quote, constants.OracleReportChunkSize)
	if formatErr != nil {
		logger.Error("Failed to format quote: ", "error", formatErr)
		return nil, appErrors.ErrFormattingQuote
	}

	// Step 3: Return the formatted oracle report
	return oracleReport, nil
}

// PrepareOracleSignature generates a Schnorr signature for the provided oracle report.
//
// This function performs the following steps sequentially:
// 1. Retrieves the Aleo context, which provides cryptographic utilities and session management.
//   - If the context cannot be retrieved, it logs the error and returns it as an application error.
//
// 2. Hashes the oracle report using the Aleo session's HashMessage method.
//   - If hashing fails, it logs the error and returns a report hashing application error.
//
// 3. Signs the hashed message using the Aleo context's Sign method to produce a Schnorr signature.
//   - If signing fails, it logs the error and returns a signature generation application error.
//
// 4. Returns the generated signature as a string if all steps succeed.
//
// Parameters:
//   - oracleReport ([]byte): The oracle report to be signed.
//
// Returns:
//   - signature (string): The Schnorr signature of the hashed oracle report.
//   - appError (*appErrors.AppError): An application error if any step fails, otherwise nil.
func PrepareOracleSignature(oracleReport []byte) (signature string, publicKey string, appError *appErrors.AppError) {
	// Step 1: Retrieve Aleo context
	aleo, err := aleoUtil.GetAleoContext()
	if err != nil {
		logger.Error("Error getting Aleo context: ", "error", err)
		return "", "", err
	}

	// Step 2: Hash the oracle report
	hashedMessage, hashErr := aleo.HashMessage(oracleReport)
	if hashErr != nil {
		logger.Error("Failed to hash report", "error", hashErr)
		return "", "", appErrors.ErrHashingReport
	}

	// Step 3: Sign the hashed message
	signature, signErr := aleo.Sign(hashedMessage)
	if signErr != nil {
		logger.Error("Error while generating signature: ", "error", signErr)
		return "", "", appErrors.ErrGeneratingSignature
	}

	// Step 4: Introduce a small random delay to mitigate timing attacks
	
	jm, randErr := rand.Int(rand.Reader, big.NewInt(51))
	if randErr != nil {
		// rare: fallback to small deterministic jitter
		jm = big.NewInt(25)
	}

	delayMs := 50 + int(jm.Int64())
	time.Sleep(time.Duration(delayMs) * time.Millisecond)

	// Step 5: Return the signature
	return signature, aleo.GetPublicKey(), nil
}

// GenerateAttestationHash generates an attestation hash from the provided user data.
//
// This function performs the following steps sequentially:
// 1. Retrieves the Aleo context, which provides cryptographic utilities and session management.
//   - If the context cannot be retrieved, it logs the error and returns it as an application error.
//
// 2. Hashes the user data using the Aleo session's HashMessage method.
//   - If hashing fails, it logs the error and returns a message hashing application error.
//
// 3. Returns the generated attestation hash if all steps succeed.
//
// Parameters:
//   - userData ([]byte): The user data to be hashed for attestation.
//
// Returns:
//   - attestationHash ([]byte): The resulting hash of the user data.
//   - err (*appErrors.AppError): An application error if any step fails, otherwise nil.
func GenerateAttestationHash(userData []byte) (attestationHash []byte, err *appErrors.AppError) {
	// Step 1: Retrieve Aleo context
	aleoContext, err := aleoUtil.GetAleoContext()
	if err != nil {
		logger.Error("Error getting Aleo context: ", "error", err)
		return nil, err
	}

	// Step 2: Hash the user data
	attestationHash, hashError := aleoContext.HashMessage(userData)
	if hashError != nil {
		logger.Error("Failed to create attestation hash: ", "error", hashError)
		return nil, appErrors.ErrCreatingAttestationHash
	}

	// Step 3: Return the attestation hash
	return attestationHash, nil
}

// BuildCompleteOracleData builds the complete oracle data that we pass to the oracle contract.
//
// This function orchestrates the construction of the OracleData struct, which contains all the information
// required for the oracle contract to verify and process an attestation. The process involves several steps:
//
// 1. Retrieve the Aleo context, which provides cryptographic and session utilities.
// 2. Prepare the encoded request using the provided user data proof and positional information.
// 3. Generate a hash of the encoded request, and obtain its string representation.
// 4. Generate a timestamped hash by combining the request hash with the attestation timestamp.
// 5. Prepare the oracle report from the provided quote (attestation evidence).
// 6. Sign the oracle report to produce a Schnorr signature.
// 7. Assemble all the above into an OracleData struct, which is returned for use by the oracle contract.
//
// Each step is logged, and errors are handled and logged appropriately. If any step fails, the function returns
// a nil OracleData and the corresponding application error.
//
// Parameters:
//   - quotePrepData: *QuotePreparationData
//     Contains the user data, proof, encoded positions, and timestamp required for attestation.
//   - quote: []byte
//     The attestation quote (evidence) to be included in the report.
//
// Returns:
//   - *OracleData: The fully constructed oracle data ready for contract submission.
//   - *appErrors.AppError: An error object if any step fails, otherwise nil.
func BuildCompleteOracleData(quotePrepData *QuotePreparationData, quote []byte) (*OracleData, *appErrors.AppError) {

	// Step 2: Prepare the encoded request
	encodedRequest, err := PrepareOracleEncodedRequest(quotePrepData.UserDataProof, quotePrepData.EncodedPositions)
	if err != nil {
		logger.Error("Failed to prepare oracle encoded request: ", "error", err)
		return nil, err
	}

	// Step 3: Prepare the request hash
	requestHash, requestHashString, err := PrepareOracleRequestHash(encodedRequest)

	if err != nil {
		logger.Error("Failed to prepare oracle request hash: ", "error", err)
		return nil, err
	}

	// Step 4: Prepare the timestamped request hash
	timestampedHash, err := PrepareOracleTimestampedRequestHash(requestHash, quotePrepData.Timestamp)

	if err != nil {
		logger.Error("Failed to prepare oracle timestamped request hash: ", "error", err)
		return nil, err
	}

	// Step 5: Prepare the oracle report
	oracleReport, err := PrepareOracleReport(quote)

	if err != nil {
		logger.Error("Failed to prepare oracle report: ", "error", err)
		return nil, err
	}

	// Step 6: Prepare the oracle signature
	signature, publicKey, err := PrepareOracleSignature(oracleReport)

	if err != nil {
		logger.Error("Failed to prepare oracle signature: ", "error", err)
		return nil, err
	}

	// Step 7: Create the oracle data
	oracleData := &OracleData{
		UserData:               string(quotePrepData.UserData),
		EncodedPositions:       quotePrepData.EncodedPositions,
		EncodedRequest:         string(encodedRequest),
		RequestHash:            requestHashString,
		TimestampedRequestHash: timestampedHash,
		Report:                 string(oracleReport),
		Signature:              signature,
		Address:                publicKey,
	}

	return oracleData, nil
}