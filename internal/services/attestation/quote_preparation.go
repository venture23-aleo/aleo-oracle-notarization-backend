package attestation

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/big"

	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	aleoContext "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/aleo_context"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"
)

// QuotePreparationData contains all the data needed for quote generation
type QuotePreparationData struct {
	UserDataProof    []byte                        `json:"userDataProof"`
	UserData         []byte                        `json:"userData"`
	EncodedPositions *encoding.ProofPositionalInfo `json:"encodedPositions"`
	AttestationHash  []byte                        `json:"attestationHash"`
	Timestamp        uint64                        `json:"timestamp"`
}

// PrepareOracleUserData prepares the user data for oracle operations
func PrepareOracleUserData(statusCode int, attestationData string, timestamp uint64, attestationRequest AttestationRequest) (userDataProof []byte, userData []byte, encodedPositions *encoding.ProofPositionalInfo, err *appErrors.AppError) {
	aleoContext, err := aleoContext.GetAleoContext()
	if err != nil {
		return nil, nil, nil, err
	}
	
	// Prepare the proof data.
	userDataProof, encodedPositions, err = PrepareProofData(statusCode, attestationData, int64(timestamp), attestationRequest)

	if err != nil {
		return nil, nil, nil, appErrors.NewAppError(appErrors.ErrPreparingProofData)
	}

	switch attestationRequest.Url {
	case constants.PriceFeedAleoUrl:
		userDataProof[0] = 8
	case constants.PriceFeedBtcUrl:
		userDataProof[0] = 12
	case constants.PriceFeedEthUrl:
		userDataProof[0] = 11
	}

	// C0 - C7 Chunks - Format the proof data.
	userData, formatError := aleoContext.GetSession().FormatMessage(userDataProof, 8)

	if formatError != nil {
		log.Println("failed to format proof data:", formatError)
		return nil, nil, nil, appErrors.NewAppError(appErrors.ErrFormattingProofData)
	}

	return userDataProof, userData, encodedPositions, nil
}

// PrepareOracleEncodedRequest prepares the encoded request for oracle operations
func PrepareOracleEncodedRequest(userDataProof []byte, encodedPositions *encoding.ProofPositionalInfo) (encodedRequest []byte, err *appErrors.AppError) {
	aleoContext, err := aleoContext.GetAleoContext()
	if err != nil {
		return nil, err
	}
	
	// Prepare the encoded request proof.
	encodedRequestProof, err := PrepareEncodedRequestProof(userDataProof, *encodedPositions)

	if err != nil {
		log.Println("failed to prepare encoded request proof: ", err)
		return nil, err
	}

	// C0 - C7 Chunks - Format the encoded proof data.
	encodedRequest, formatError := aleoContext.GetSession().FormatMessage(encodedRequestProof, 8)
	if formatError != nil {
		log.Println("failed to format encoded proof data:", formatError)
		return nil, appErrors.NewAppError(appErrors.ErrFormattingEncodedProofData)
	}

	return encodedRequest, nil
}

// PrepareOracleRequestHash prepares the request hash for oracle operations
func PrepareOracleRequestHash(encodedRequest []byte) (requestHash []byte, requestHashString string, err *appErrors.AppError) {
	aleoContext, err := aleoContext.GetAleoContext()
	if err != nil {
		return nil, "", err
	}
	
	// Create the request hash - Hash the encoded request.
	requestHash, hashError := aleoContext.GetSession().HashMessage(encodedRequest)

	if hashError != nil {
		log.Println("failed to create request hash:", hashError)
		return nil, "", appErrors.NewAppError(appErrors.ErrCreatingRequestHash)
	}

	// Create the request hash string - Hash the encoded request.
	requestHashString, hashError = aleoContext.GetSession().HashMessageToString(encodedRequest)

	if hashError != nil {
		log.Println("failed to create request hash:", hashError)
		return nil, "", appErrors.NewAppError(appErrors.ErrCreatingRequestHash)
	}

	return requestHash, requestHashString, nil
}

// PrepareOracleTimestampedRequestHash prepares the timestamped request hash for oracle operations
func PrepareOracleTimestampedRequestHash(requestHash []byte, timestamp uint64) (timestampedRequestHash string, err *appErrors.AppError) {
	aleoContext, err := aleoContext.GetAleoContext()
	if err != nil {
		return "", err
	}
	
	// Create the timestamped hash input - Hash the encoded request with the timestamp.
	timestampBytes := make([]byte, encoding.TARGET_ALIGNMENT)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(timestamp))

	timestampedHashInput := append(requestHash, timestampBytes...)

	timestampedHashInputChunk1 := new(big.Int).SetBytes(utils.ReverseBytes(timestampedHashInput[0:16]))
	timestampedHashInputChunk2 := new(big.Int).SetBytes(utils.ReverseBytes(timestampedHashInput[16:32]))

	// Create the timestamped format message - Format the timestamped hash input.
	timesampedFormatMessage := fmt.Sprintf("{ request_hash: %su128, attestation_timestamp: %su128 }", timestampedHashInputChunk1, timestampedHashInputChunk2)

	// Create the timestamped hash - Hash the timestamped format message.
	timestampedHash, hashError := aleoContext.GetSession().HashMessageToString([]byte(timesampedFormatMessage))

	// Check if the error is not nil.
	if hashError != nil {
		log.Println("failed to creat timestamped hash:", err)
		return "", appErrors.NewAppError(appErrors.ErrCreatingTimestampedHash)
	}

	return timestampedHash, nil
}

// PrepareDataForQuoteGeneration prepares all the data needed for quote generation
func PrepareDataForQuoteGeneration(statusCode int, attestationData string, timestamp uint64, attestationRequest AttestationRequest) (*QuotePreparationData, *appErrors.AppError) {
	userDataProof, userData, encodedPositions, err := PrepareOracleUserData(statusCode, attestationData, timestamp, attestationRequest)
	if err != nil {
		return nil, err
	}

	attestationHash, err := GenerateAttestationHash(userData)
	if err != nil {
		return nil, err
	}

	return &QuotePreparationData{
		UserDataProof:    userDataProof,
		UserData:         userData,
		EncodedPositions: encodedPositions,
		AttestationHash:  attestationHash,
		Timestamp:        timestamp,
	}, nil
} 