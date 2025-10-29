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
			expectedEncodedRequest: "{  c0: {    f0: 65387592075003861606687669663359381809u128,    f1: 65387592075003861606687669663359381809u128,    f2: 0u128,    f3: 0u128,    f4: 65387592075003861606687669663359381809u128,    f5: 65387592075003861606687669663359381809u128,    f6: 825307441u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c1: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c2: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c3: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c4: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c5: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c6: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c7: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  }}",
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
