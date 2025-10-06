// Package attestation prepares oracle data, hashes, and request artifacts for SGX quoting.
package attestation
import (
	"encoding/binary"
	"fmt"

	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	aleoUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/aleoutil"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/common"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

// QuotePreparationData contains all the data needed for quote generation
type QuotePreparationData struct {
	UserDataProof    []byte                        `json:"userDataProof"`    // The user data proof.
	UserData         []byte                        `json:"userData"`         // The user data.
	EncodedPositions *encoding.ProofPositionalInfo `json:"encodedPositions"` // The encoded positions.
	AttestationHash  []byte                        `json:"attestationHash"`  // The attestation hash.
	Timestamp        uint64                        `json:"timestamp"`        // The timestamp.
}

// PrepareOracleUserData prepares the user data for oracle operations.
//
// This function orchestrates the preparation of user data required for oracle attestation and quote generation.
// The process involves:
//  1. Retrieving the Aleo context, which provides access to cryptographic session methods.
//  2. Preparing the proof data and its positional encoding using the provided status code, attestation data, timestamp, and attestation request.
//  3. Setting the token ID in the proof data based on the attestation request URL (for ALEO, BTC, or ETH price feeds).
//  4. Formatting the proof data into the required chunked message format (C0 - C7) for further processing.
//  5. Logging and returning any errors encountered during the process.
//
// Parameters:
//   - statusCode:        the HTTP status code associated with the attestation
//   - attestationData:   the attestation data as a string
//   - timestamp:         the timestamp of the attestation (as uint64)
//   - attestationRequest: the attestation request object containing URL and other metadata
//
// Returns:
//   - userDataProof:     the prepared proof data as a byte slice
//   - userData:          the formatted user data as a byte slice (C0 - C7 chunks)
//   - encodedPositions:  pointer to the positional encoding information
//   - err:               an application error if any step fails
func PrepareOracleUserData(
	statusCode int,
	attestationData string,
	timestamp uint64,
	attestationRequest AttestationRequest,
) (
	userDataProof []byte,
	userData []byte,
	encodedPositions *encoding.ProofPositionalInfo,
	err *appErrors.AppError,
) {
	// Step 1: Get the Aleo context.
	aleoContext, err := aleoUtil.GetAleoContext()
	if err != nil {
		return nil, nil, nil, err
	}

	// Step 2: Prepare the proof data.
	userDataProof, encodedPositions, err = PrepareProofData(statusCode, attestationData, int64(timestamp), attestationRequest)
	if err != nil {
		return nil, nil, nil, appErrors.ErrPreparingProofData
	}

	if common.IsPriceFeedURL(attestationRequest.Url) {
		tokenID := common.GetTokenIDFromPriceFeedURL(attestationRequest.Url)
		logger.Debug("Token ID: ", "tokenID", tokenID)
		if tokenID == 0 {
			logger.Error("Unsupported price feed URL: ", "url", attestationRequest.Url)
			return nil, nil, nil, appErrors.ErrUnsupportedPriceFeedURL
		}
		userDataProof[0] = byte(tokenID)
	}

	// Step 4: Format the proof data into C0 - C7 chunks.
	userData, formatError := aleoContext.GetSession().FormatMessage(userDataProof, constants.OracleUserDataChunkSize)

	if formatError != nil {
		logger.Error("failed to format proof data", "error", formatError)
		return nil, nil, nil, appErrors.ErrFormattingProofData
	}

	// Step 5: Return the prepared data.
	return userDataProof, userData, encodedPositions, nil
}

// PrepareOracleEncodedRequest prepares the encoded request for oracle operations.
//
// This function takes the user data proof and its associated positional encoding information,
// and produces an encoded request suitable for oracle operations. The process involves:
//  1. Retrieving the Aleo context, which provides access to cryptographic session methods.
//  2. Preparing the encoded request proof using the provided user data proof and positional info.
//  3. Formatting the encoded request proof into the required chunked message format (C0 - C7).
//  4. Logging and returning any errors encountered during the process.
//
// Parameters:
//   - userDataProof: the proof data as a byte slice
//   - encodedPositions: pointer to the positional encoding information
//
// Returns:
//   - encodedRequest: the resulting encoded request as a byte slice
//   - err: an application error if any step fails
func PrepareOracleEncodedRequest(userDataProof []byte, encodedPositions *encoding.ProofPositionalInfo) (encodedRequest []byte, err *appErrors.AppError) {
	// Step 1: Retrieve the Aleo context.
	aleoContext, err := aleoUtil.GetAleoContext()
	if err != nil {
		return nil, err
	}

	if encodedPositions == nil {
		return nil, appErrors.ErrNilEncodedPositions
	}

	// Step 2: Prepare the encoded request proof.
	encodedRequestProof, err := PrepareEncodedRequestProof(userDataProof, *encodedPositions)
	if err != nil {
		logger.Error("failed to prepare encoded request proof: ", "error", err)
		return nil, err
	}

	// Step 3: Format the encoded proof data into C0 - C7 chunks.
	encodedRequest, formatError := aleoContext.GetSession().FormatMessage(encodedRequestProof, 8)
	if formatError != nil {
		logger.Error("failed to format encoded proof data:", "error", formatError)
		return nil, appErrors.ErrFormattingEncodedProofData
	}

	// Step 4: Return the encoded request.
	return encodedRequest, nil
}

// PrepareOracleRequestHash prepares the request hash for oracle operations.
//
// This function takes an encoded request as a byte slice and generates two forms of its hash:
//  1. A byte slice hash of the encoded request.
//  2. A string representation of the hash of the encoded request.
//
// The process is as follows:
//  1. Retrieve the Aleo context, which provides access to cryptographic session methods.
//  2. Use the session's HashMessage method to hash the encoded request and obtain the byte slice hash.
//     - If hashing fails, log the error and return an application error.
//  3. Use the session's HashMessageToString method to hash the encoded request and obtain the string hash.
//     - If hashing fails, log the error and return an application error.
//  4. Return both the byte slice hash and the string hash.
//
// Parameters:
//   - encodedRequest: the encoded request as a byte slice
//
// Returns:
//   - requestHash: the resulting hash as a byte slice
//   - requestHashString: the resulting hash as a string
//   - err: an application error if any step fails
func PrepareOracleRequestHash(encodedRequest []byte) (requestHash []byte, requestHashString string, err *appErrors.AppError) {
	// Step 1: Retrieve the Aleo context.
	aleoContext, err := aleoUtil.GetAleoContext()
	if err != nil {
		return nil, "", err
	}

	// Step 2: Create the request hash - Hash the encoded request.
	requestHash, hashError := aleoContext.GetSession().HashMessage(encodedRequest)
	if hashError != nil {
		logger.Error("failed to create request hash:", "error", hashError)
		return nil, "", appErrors.ErrCreatingRequestHash
	}

	// Defensive check: ensure hash size is 16 bytes as expected by downstream consumers.
	if len(requestHash) != encoding.TARGET_ALIGNMENT {
		logger.Error("unexpected request hash length", "got", len(requestHash), "want", encoding.TARGET_ALIGNMENT)
		return nil, "", appErrors.ErrCreatingRequestHash
	}

	// Step 3: Create the request hash string - Hash the encoded request.
	requestHashString, hashError = aleoContext.GetSession().HashMessageToString(encodedRequest)
	if hashError != nil {
		logger.Error("failed to create request hash:", "error", hashError)
		return nil, "", appErrors.ErrCreatingRequestHash
	}

	// Step 4: Return both the byte slice hash and the string hash.
	return requestHash, requestHashString, nil
}

// PrepareOracleTimestampedRequestHash generates a timestamped request hash for oracle operations.
//
// This function takes a request hash (as a byte slice) and a timestamp (as a uint64), and produces a
// deterministic string hash that combines both values. The process is as follows:
//
//  1. Retrieve the Aleo context, which provides access to cryptographic session methods.
//  2. Convert the timestamp to a byte slice of length encoding.TARGET_ALIGNMENT using little-endian encoding.
//  3. Concatenate the request hash and the timestamp bytes to form the input for the timestamped hash.
//  4. Split the concatenated input into two chunks of encoding.TARGET_ALIGNMENT bytes each, reverse their byte order,
//     and convert them to big integers. These represent the request hash and the attestation timestamp, respectively.
//  5. Format these two chunks into a string using the format:
//     "{ request_hash: <chunk1>u128, attestation_timestamp: <chunk2>u128 }"
//  6. Hash this formatted string using the Aleo session's HashMessageToString method to produce the final timestamped request hash.
//  7. If any error occurs during the process, log the error and return an appropriate application error.
//
// Parameters:
//   - requestHash: the original request hash as a byte slice
//   - timestamp: the attestation timestamp as a uint64
//
// Returns:
//   - timestampedRequestHash: the resulting hash as a string
//   - err: an application error if any step fails
func PrepareOracleTimestampedRequestHash(requestHash []byte, timestamp uint64) (timestampedRequestHash string, err *appErrors.AppError) {
	// Step 1: Retrieve the Aleo context.
	aleoContext, err := aleoUtil.GetAleoContext()
	if err != nil {
		return "", err
	}

	// Defensive check: ensure requestHash has expected size.
	if len(requestHash) != encoding.TARGET_ALIGNMENT {
		logger.Error("invalid request hash length for timestamped hash", "got", len(requestHash), "want", encoding.TARGET_ALIGNMENT)
		return "", appErrors.ErrCreatingTimestampedRequestHash
	}

	// Step 2: Prepare the timestamp bytes.
	timestampBytes := make([]byte, encoding.TARGET_ALIGNMENT)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(timestamp))

	// Step 3: Append the timestamp bytes to the request hash.
	timestampedRequestHashInput := append(requestHash, timestampBytes...)

	// Step 4: Create the timestamped request hash input chunks.
	timestampedRequestHashInputChunk1, err := common.SliceToU128(timestampedRequestHashInput[:encoding.TARGET_ALIGNMENT])
	if err != nil {
		logger.Error("Failed to create timestamped request hash input chunk 1: ", "error", err)
		return "", err
	}

	timestampedRequestHashInputChunk2, err := common.SliceToU128(timestampedRequestHashInput[encoding.TARGET_ALIGNMENT : 2*encoding.TARGET_ALIGNMENT])
	if err != nil {
		logger.Error("Failed to create timestamped request hash input chunk 2: ", "error", err)
		return "", err
	}

	// Step 5: Create the timestamped request hash format message.
	timestampedRequestHashFormatMessage := fmt.Sprintf("{ request_hash: %su128, attestation_timestamp: %su128 }", timestampedRequestHashInputChunk1, timestampedRequestHashInputChunk2)

	// Step 6: Create the timestamped request hash.
	timestampedRequestHash, hashError := aleoContext.GetSession().HashMessageToString([]byte(timestampedRequestHashFormatMessage))

	// Step 7: Handle error.
	if hashError != nil {
		logger.Error("Failed to create timestamped request hash: ", "error", hashError)
		return "", appErrors.ErrCreatingTimestampedRequestHash
	}

	return timestampedRequestHash, nil
}

// PrepareDataForQuoteGeneration prepares all the data needed for quote generation.
//
// This function orchestrates the preparation of all required components for generating an SGX quote
// in the attestation process. It performs the following steps:
//
//  1. Calls PrepareOracleUserData to generate the user data proof, user data, and encoded positions
//     based on the provided status code, attestation data, timestamp, and attestation request.
//     - If an error occurs during this step, it returns the error immediately.
//  2. Calls GenerateAttestationHash to compute a hash of the user data, which serves as the attestation hash.
//     - If an error occurs during this step, it logs the error and returns it.
//  3. Constructs and returns a QuotePreparationData struct containing all the prepared data.
//
// Parameters:
//   - statusCode:        The HTTP status code to be included in the attestation.
//   - attestationData:   The attestation data as a string (e.g., random number, price, etc.).
//   - timestamp:         The attestation timestamp as a uint64.
//   - attestationRequest: The original attestation request details.
//
// Returns:
//   - *QuotePreparationData: A pointer to the struct containing all prepared data for quote generation.
//   - *appErrors.AppError:   An application error if any step fails, otherwise nil.
func PrepareDataForQuoteGeneration(
	statusCode int,
	attestationData string,
	timestamp uint64,
	attestationRequest AttestationRequest,
) (*QuotePreparationData, *appErrors.AppError) {
	// Step 1: Prepare user data proof, user data, and encoded positions.
	userDataProof, userData, encodedPositions, err := PrepareOracleUserData(statusCode, attestationData, timestamp, attestationRequest)
	if err != nil {
		return nil, err
	}

	// Step 2: Generate attestation hash from user data.
	attestationHash, err := GenerateAttestationHash(userData)
	if err != nil {
		logger.Error("Failed to generate attestation hash: ", "error", err)
		return nil, err
	}

	// Step 3: Return all prepared data in a QuotePreparationData struct.
	return &QuotePreparationData{
		UserDataProof:    userDataProof,
		UserData:         userData,
		EncodedPositions: encodedPositions,
		AttestationHash:  attestationHash,
		Timestamp:        timestamp,
	}, nil
}
