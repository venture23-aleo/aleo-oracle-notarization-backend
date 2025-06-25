package services

import (
	"encoding/binary"
	"fmt"
	"log"
	"math/big"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/utils"

	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/configs"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"

	encoding "github.com/zkportal/aleo-oracle-encoding"
	aleo "github.com/zkportal/aleo-utils-go"
)

// PositionInfo is the positional information of the data.
type PositionInfo struct {
	// Index of the block where the write operation started. Indexing starts from 0. Note that this number doesn't account the fact that each chunk contains 32 blocks.
	//
	// If Pos is >32, it means that there was an "overflow" to the next chunk of 32 blocks, e.g. Pos 31 means chunk 0 field 31, Pos 32 means chunk 1, field 0.
	Pos int

	// Number of blocks written in the write operation.
	Len int
}

// ProofPositionalInfo is the positional information of the proof data.
type ProofPositionalInfo struct {
	Data            PositionInfo `json:"data"`
	Timestamp       PositionInfo `json:"timestamp"`
	StatusCode      PositionInfo `json:"statusCode"`
	Method          PositionInfo `json:"method"`
	ResponseFormat  PositionInfo `json:"responseFormat"`
	Url             PositionInfo `json:"url"`
	Selector        PositionInfo `json:"selector"`
	EncodingOptions PositionInfo `json:"encodingOptions"`
	RequestHeaders  PositionInfo `json:"requestHeaders"`
	OptionalFields  PositionInfo `json:"optionalFields"` // Optional fields are HTML result type, request content type, request body. They're all encoded together.
}

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
	EncodedPositions ProofPositionalInfo `json:"encodedPositions"`

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

// PrepareOracleDataBeforeQuote prepares the oracle data before the quote.
func PrepareOracleDataBeforeQuote(s aleo.Session, statusCode int, attestationData string, timestamp uint64, attestationRequest AttestationRequest) (OracleData, error) {

	// Prepare the proof data.
	userDataProof, encodedPositions, err := PrepareProofData(statusCode, attestationData, int64(timestamp), attestationRequest)

	// log.Print("Proof data: ", hex.EncodeToString(userDataProof))

	if err != nil {
		return OracleData{}, appErrors.ErrPreparingProofData
	}

	if attestationRequest.Url == constants.PriceFeedAleoUrl {
		userDataProof[0] = 8
	} else if attestationRequest.Url == constants.PriceFeedBtcUrl {
		userDataProof[0] = 12
	} else if attestationRequest.Url == constants.PriceFeedEthUrl {
		userDataProof[0] = 11
	}

	// C0 - C7 Chunks - Format the proof data.
	userData, err := s.FormatMessage(userDataProof, 8)

	if err != nil {
		log.Println("failed to format proof data:", err)
		return OracleData{}, appErrors.ErrFormattingProofData
	}

	// attestationHash, err := s.HashMessage(userData)

	// if err != nil {
	// 	log.Printf("aleo.HashMessage(): %v\n", err)
	// 	return OracleData{}, appErrors.ErrGeneratingAttestationHash
	// }

	// log.Printf("Attestation hash: %v", hex.EncodeToString(attestationHash))

	// Prepare the encoded request proof.
	encodedProofData, err := PrepareEncodedRequestProof(userDataProof, encodedPositions)

	if err != nil {
		log.Println("failed to prepare encoded request proof: ", err)
		return OracleData{}, err.(appErrors.AppError)
	}

	// C0 - C7 Chunks - Format the encoded proof data.
	encodedRequest, err := s.FormatMessage(encodedProofData, 8)
	if err != nil {
		log.Println("failed to format encoded proof data:", err)
		return OracleData{}, appErrors.ErrFormattingEncodedProofData
	}

	// log.Print("encodedRequest", hex.EncodeToString(encodedRequest))

	// Create the request hash - Hash the encoded request.
	requestHash, err := s.HashMessage(encodedRequest)

	if err != nil {
		log.Println("failed to create request hash:", err)
		return OracleData{}, appErrors.ErrCreatingRequestHash
	}

	// Create the request hash string - Hash the encoded request.
	requestHashString, err := s.HashMessageToString(encodedRequest)

	if err != nil {
		log.Println("failed to create request hash:", err)
		return OracleData{}, appErrors.ErrCreatingRequestHash
	}

	// Create the timestamped hash input - Hash the encoded request with the timestamp.
	timestampBytes := make([]byte, encoding.TARGET_ALIGNMENT)
	binary.LittleEndian.PutUint64(timestampBytes, uint64(timestamp))

	timestampedHashInput := append(requestHash, timestampBytes...)

	timestampedHashInputChunk1 := new(big.Int).SetBytes(utils.ReverseBytes(timestampedHashInput[0:16]))
	timestampedHashInputChunk2 := new(big.Int).SetBytes(utils.ReverseBytes(timestampedHashInput[16:32]))

	// Create the timestamped format message - Format the timestamped hash input.
	timesampedFormatMessage := fmt.Sprintf("{ request_hash: %su128, attestation_timestamp: %su128 }", timestampedHashInputChunk1, timestampedHashInputChunk2)

	log.Printf("timesampedFormatMessage: %v", string(timesampedFormatMessage))

	// Create the timestamped hash - Hash the timestamped format message.
	timestampedHash, err := s.HashMessageToString([]byte(timesampedFormatMessage))

	// Check if the error is not nil.
	if err != nil {
		log.Println("failed to creat timestamped hash:", err)
		return OracleData{}, appErrors.ErrCreatingTimestampedHash
	}

	// Create the result - Set the user data, encoded positions, encoded request, request hash, and timestamped request hash.
	result := OracleData{
		UserData:               string(userData),
		EncodedPositions:       encodedPositions,
		EncodedRequest:         string(encodedRequest),
		RequestHash:            requestHashString,
		TimestampedRequestHash: timestampedHash,
	}

	return result, nil
}

// PrepareOracleDataAfterQuote prepares the oracle data after the quote.
func PrepareOracleDataAfterQuote(s aleo.Session, oracleData OracleData, quote []byte) (OracleData, error) {

	// C0 - C9 Chunks
	oracleReport, err := s.FormatMessage(quote, 10)

	if err != nil {
		log.Println("failed to format message:", err)
		return oracleData, appErrors.ErrFormattingQuote
	}

	// Create the hashed message.
	hashedMessage, err := s.HashMessage(oracleReport)

	// Check if the error is not nil.
	if err != nil {
		return oracleData, appErrors.ErrReportHashing
	}

	// Create the signature.
	signature, err := s.Sign(configs.PrivateKey, hashedMessage)

	// Check if the error is not nil.
	if err != nil {
		log.Fatalln("Sign failed:", err)
		return oracleData, appErrors.ErrGeneratingSignature
	}

	// Set the report.
	oracleData.Report = string(oracleReport)

	// Set the signature.
	oracleData.Signature = signature

	// Set the address.
	oracleData.Address = configs.PublicKey

	return oracleData, nil
}
