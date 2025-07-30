package attestation

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	"github.com/venture23-aleo/aleo-oracle-encoding/positionRecorder"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
)

func TestPrepareOracleUserData(t *testing.T) {
	testCases := []struct {
		name               string
		attestationRequest AttestationRequest
		statusCode         int
		attestationData    string
		timestamp          uint64
		expectedError      *appErrors.AppError
		expectedUserData   string
	}{
		{
			name: "valid attestation data - string encoding",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			statusCode:      200,
			attestationData: "attestation data",
			timestamp:       1715769600,
			expectedError:   nil,
		},
		{
			name: "invalid user data - float encoding but attestation data is not a float",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			statusCode:      200,
			attestationData: "attestation data",
			timestamp:       1715769600,
			expectedError:   appErrors.ErrPreparingProofData,
		},
		{
			name: "invalid user data - float encoding but attestation data is too long",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			statusCode:      200,
			attestationData: string(bytes.Repeat([]byte("1"), 10000)),
			timestamp:       1715769600,
			expectedError:   appErrors.ErrPreparingProofData,
		},
		{
			name: "valid price feed url - btc",
			attestationRequest: AttestationRequest{
				Url:            constants.PriceFeedBTCURL,
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "weightedAvgPrice",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			statusCode:      200,
			attestationData: "10000",
			timestamp:       1715769600,
			expectedError:   nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, _, err := PrepareOracleUserData(testCase.statusCode, testCase.attestationData, testCase.timestamp, testCase.attestationRequest)
			assert.Equal(t, testCase.expectedError, err)
		})
	}
}

func TestPrepareOracleEncodedRequest(t *testing.T) {

	testCases := []struct {
		name                   string
		userDataProof          []byte
		encodedPositions       encoding.ProofPositionalInfo
		expectedError          *appErrors.AppError
		expectedEncodedRequest string
	}{
		{
			name:          "valid user data and encoded positions",
			userDataProof: bytes.Repeat([]byte("1"), 100),
			expectedError: nil,
			encodedPositions: encoding.ProofPositionalInfo{
				Data: positionRecorder.PositionInfo{
					Pos: 2,
					Len: 1,
				},
				Timestamp: positionRecorder.PositionInfo{
					Pos: 3,
					Len: 1,
				},
				StatusCode: positionRecorder.PositionInfo{
					Pos: 4,
					Len: 1,
				},
				Url: positionRecorder.PositionInfo{
					Pos: 5,
					Len: 1,
				},
				Selector: positionRecorder.PositionInfo{
					Pos: 6,
					Len: 1,
				},
				ResponseFormat: positionRecorder.PositionInfo{
					Pos: 7,
					Len: 1,
				},
			},
		},
		{
			name:          "invalid user data and encoded positions",
			userDataProof: bytes.Repeat([]byte("1"), 10),
			encodedPositions: encoding.ProofPositionalInfo{
				Data: positionRecorder.PositionInfo{
					Pos: 2,
					Len: 1,
				},
				Timestamp: positionRecorder.PositionInfo{
					Pos: 3,
					Len: 1,
				},
				StatusCode: positionRecorder.PositionInfo{
					Pos: 4,
					Len: 1,
				},
			},
			expectedError: appErrors.ErrUserDataTooShort,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			encodedRequest, err := PrepareOracleEncodedRequest(testCase.userDataProof, &testCase.encodedPositions)
			assert.Equal(t, testCase.expectedError, err)
			if err == nil {
				assert.Equal(t, testCase.expectedEncodedRequest, string(encodedRequest))
			}
		})
	}
}

func TestPrepareDataForQuoteGeneration(t *testing.T) {
	testCases := []struct {
		name               string
		attestationRequest AttestationRequest
		expectedError      *appErrors.AppError
		statusCode         int
		attestationData    string
		timestamp          uint64
	}{
		{
			name: "invalid attestation data for float encoding",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			statusCode:      200,
			attestationData: "attestation data",
			timestamp:       1715769600,
			expectedError:   appErrors.ErrPreparingProofData,
		},
		{
			name: "large attestation data for float encoding",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			statusCode:      200,
			attestationData: "attestation data",
			timestamp:       1715769600,
			expectedError:   appErrors.ErrPreparingProofData,
		},
		{
			name: "valid attestation data with string encoding option",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value: "string",
				},
			},
			statusCode:      200,
			attestationData: "test string",
			timestamp:       1715769600,
			expectedError:   nil,
		},
		{
			name: "valid attestation data for float encoding",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			statusCode:      200,
			attestationData: "1345",
			timestamp:       1715769600,
			expectedError:   nil,
		},
		{
			name: "valid attestation data for float encoding with btc price feed url",
			attestationRequest: AttestationRequest{
				Url:            constants.PriceFeedBTCURL,
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			statusCode:      200,
			attestationData: "11345",
			timestamp:       1715769600,
			expectedError:   nil,
		},
		{
			name: "valid attestation data for float encoding with eth price feed url",
			attestationRequest: AttestationRequest{
				Url:            constants.PriceFeedETHURL,
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			statusCode:      200,
			attestationData: "4000",
			timestamp:       1715769600,
			expectedError:   nil,
		},
		{
			name: "valid attestation data for float encoding with aleo price feed url",
			attestationRequest: AttestationRequest{
				Url:            constants.PriceFeedAleoURL,
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 6,
				},
			},
			statusCode:      200,
			attestationData: "0.224",
			timestamp:       1715769600,
			expectedError:   nil,
		},
		{
			name: "invalid request",
			attestationRequest: AttestationRequest{
				Url:            "https://www.google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
			},
			statusCode:      200,
			attestationData: "attestation data",
			timestamp:       1715769600,
			expectedError:   appErrors.ErrPreparingProofData,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			quotePreparationData, err := PrepareDataForQuoteGeneration(testCase.statusCode, testCase.attestationData, testCase.timestamp, testCase.attestationRequest)
			if testCase.expectedError != nil {
				assert.Equal(t, testCase.expectedError, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, quotePreparationData)
				assert.NotEmpty(t, quotePreparationData.UserDataProof)
				assert.NotEmpty(t, quotePreparationData.UserData)
				assert.NotEmpty(t, quotePreparationData.EncodedPositions)
			}
		})
	}
}
