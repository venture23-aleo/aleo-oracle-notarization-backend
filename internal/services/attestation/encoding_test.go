package attestation

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	"github.com/venture23-aleo/aleo-oracle-encoding/positionRecorder"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
)

func TestPrepareProofData_PositionalInfo(t *testing.T) {
	testCases := []struct {
		name                   string
		attestationRequest     AttestationRequest
		expectedError          *appErrors.AppError
		statusCode             int
		attestationData        string
		timestamp              int64
		expectedPositionalInfo *encoding.ProofPositionalInfo
	}{
		{
			name: "invalid attestation data for float encoding",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 0,
				},
			},
			statusCode:             200,
			attestationData:        "attestation data",
			timestamp:              1715769600,
			expectedError:          appErrors.NewAppError(appErrors.ErrEncodingAttestationData),
			expectedPositionalInfo: nil,
		},
		{
			name: "invalid encoding option",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "strings",
					Precision: 0,
				},
			},
			statusCode:             200,
			attestationData:        "attestation data",
			timestamp:              1715769600,
			expectedError:          appErrors.NewAppError(appErrors.ErrInvalidEncodingOption),
			expectedPositionalInfo: nil,
		},
		{
			name: "large attestation data for float encoding with length >= 255",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 0,
				},
			},
			statusCode:             200,
			attestationData:        string(bytes.Repeat([]byte("1"), 255)),
			timestamp:              1715769600,
			expectedError:          appErrors.NewAppError(appErrors.ErrAttestationDataTooLarge),
			expectedPositionalInfo: nil,
		},
		{
			name: "large attestation data for float encoding with value > float64 max and length less than 255",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 0,
				},
			},
			statusCode:             200,
			attestationData:        string(bytes.Repeat([]byte("1"), 100)),
			timestamp:              1715769600,
			expectedError:          appErrors.NewAppError(appErrors.ErrEncodingAttestationData),
			expectedPositionalInfo: nil,
		},
		{
			name: "float encoding with precision > 12",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "float",
					Precision: 13,
				},
			},
			statusCode:             200,
			attestationData:        string(bytes.Repeat([]byte("1"), 10)),
			timestamp:              1715769600,
			expectedError:          appErrors.NewAppError(appErrors.ErrEncodingAttestationData),
			expectedPositionalInfo: nil,
		},
		{
			name: "large attestation data for string encoding with length >= 3072",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "string",
					Precision: 0,
				},
			},
			statusCode:             200,
			attestationData:        string(bytes.Repeat([]byte("a"), 10000)),
			timestamp:              1715769600,
			expectedError:          appErrors.NewAppError(appErrors.ErrAttestationDataTooLarge),
			expectedPositionalInfo: nil,
		},
		{
			name: "large attestation data for string encoding with length 3000",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "string",
					Precision: 0,
				},
			},
			statusCode:      200,
			attestationData: string(bytes.Repeat([]byte("a"), 3000)),
			timestamp:       1715769600,
			expectedError:   nil,
			expectedPositionalInfo: &encoding.ProofPositionalInfo{
				Data: positionRecorder.PositionInfo{
					Pos: 2,
					Len: 192,
				},
				Timestamp: positionRecorder.PositionInfo{
					Pos: 194,
					Len: 1,
				},
				StatusCode: positionRecorder.PositionInfo{
					Pos: 195,
					Len: 1,
				},
				Url: positionRecorder.PositionInfo{
					Pos: 196,
					Len: 1,
				},
				Selector: positionRecorder.PositionInfo{
					Pos: 197,
					Len: 1,
				},
				ResponseFormat: positionRecorder.PositionInfo{
					Pos: 198,
					Len: 1,
				},
				Method: positionRecorder.PositionInfo{
					Pos: 199,
					Len: 1,
				},
				EncodingOptions: positionRecorder.PositionInfo{
					Pos: 200,
					Len: 1,
				},
				RequestHeaders: positionRecorder.PositionInfo{
					Pos: 201,
					Len: 1,
				},
				OptionalFields: positionRecorder.PositionInfo{
					Pos: 202,
					Len: 4,
				},
			},
		},
		{
			name: "valid attestation data with float encoding",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
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
			expectedPositionalInfo: &encoding.ProofPositionalInfo{
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
				Method: positionRecorder.PositionInfo{
					Pos: 8,
					Len: 1,
				},
				EncodingOptions: positionRecorder.PositionInfo{
					Pos: 9,
					Len: 1,
				},
				RequestHeaders: positionRecorder.PositionInfo{
					Pos: 10,
					Len: 1,
				},
				OptionalFields: positionRecorder.PositionInfo{
					Pos: 11,
					Len: 4,
				},
			},
		},
		{
			name: "valid attestation data with float encoding and price feed url",
			attestationRequest: AttestationRequest{
				Url:            constants.PRICE_FEED_BTC_URL,
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
			expectedPositionalInfo: &encoding.ProofPositionalInfo{
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
				Method: positionRecorder.PositionInfo{
					Pos: 8,
					Len: 1,
				},
				EncodingOptions: positionRecorder.PositionInfo{
					Pos: 9,
					Len: 1,
				},
				RequestHeaders: positionRecorder.PositionInfo{
					Pos: 10,
					Len: 1,
				},
				OptionalFields: positionRecorder.PositionInfo{
					Pos: 11,
					Len: 4,
				},
			},
		},
		{
			name: "valid attestation data with string encoding",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
				EncodingOptions: encoding.EncodingOptions{
					Value:     "string",
					Precision: 0,
				},
			},
			statusCode:      200,
			attestationData: "1345",
			timestamp:       1715769600,
			expectedError:   nil,
			expectedPositionalInfo: &encoding.ProofPositionalInfo{
				Data: positionRecorder.PositionInfo{
					Pos: 2,
					Len: 192,
				},
				Timestamp: positionRecorder.PositionInfo{
					Pos: 194,
					Len: 1,
				},
				StatusCode: positionRecorder.PositionInfo{
					Pos: 195,
					Len: 1,
				},
				Url: positionRecorder.PositionInfo{
					Pos: 196,
					Len: 1,
				},
				Selector: positionRecorder.PositionInfo{
					Pos: 197,
					Len: 1,
				},
				ResponseFormat: positionRecorder.PositionInfo{
					Pos: 198,
					Len: 1,
				},
				Method: positionRecorder.PositionInfo{
					Pos: 199,
					Len: 1,
				},
				EncodingOptions: positionRecorder.PositionInfo{
					Pos: 200,
					Len: 1,
				},
				RequestHeaders: positionRecorder.PositionInfo{
					Pos: 201,
					Len: 1,
				},
				OptionalFields: positionRecorder.PositionInfo{
					Pos: 202,
					Len: 4,
				},
			},
		},
		{
			name: "invalid request with no encoding options",
			attestationRequest: AttestationRequest{
				Url:            "google.com",
				RequestMethod:  "GET",
				ResponseFormat: "json",
				Selector:       "body",
			},
			statusCode:             200,
			attestationData:        "attestation data",
			timestamp:              1715769600,
			expectedError:          appErrors.NewAppError(appErrors.ErrInvalidEncodingOption),
			expectedPositionalInfo: nil,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			t.Logf("testCase.attestationData: %v", len(testCase.attestationData))
			proofData, encodedPositions, err := PrepareProofData(testCase.statusCode, testCase.attestationData, testCase.timestamp, testCase.attestationRequest)
			if testCase.expectedError != nil {
				assert.Equal(t, testCase.expectedError, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, proofData)
				assert.NotNil(t, encodedPositions)
				assert.Equal(t, testCase.expectedPositionalInfo, encodedPositions)

				var attestationData = make([]byte, 2)
				var attestationDataLen int

				if testCase.attestationRequest.Url == constants.PRICE_FEED_BTC_URL || testCase.attestationRequest.Url == constants.PRICE_FEED_ETH_URL || testCase.attestationRequest.Url == constants.PRICE_FEED_ALEO_URL {
					attestationDataLen = len(testCase.attestationData)
				} else {
					switch testCase.attestationRequest.EncodingOptions.Value {
					case "float":
						attestationDataLen = math.MaxUint8
					case "string":
						attestationDataLen = constants.ATTESTATION_DATA_SIZE_LIMIT
					case "int":
						attestationDataLen = math.MaxUint8
					default:
						assert.Fail(t, "invalid encoding option")
					}
				}

				binary.LittleEndian.PutUint16(attestationData, uint16(attestationDataLen))
				assert.Equal(t, proofData[0:2], attestationData)

				// Ensure the attestation data length is encoded in little endian format (2 bytes)
				expectedTimestampLen := make([]byte, 2)
				binary.LittleEndian.PutUint16(expectedTimestampLen, uint16(8))
				assert.Equal(t, proofData[2:4], expectedTimestampLen)

				expectedStatusCodeLen := make([]byte, 2)
				binary.LittleEndian.PutUint16(expectedStatusCodeLen, uint16(8))
				assert.Equal(t, proofData[4:6], expectedStatusCodeLen)

				// Ensure the method length is encoded in little endian format (2 bytes)
				expectedMethodLen := make([]byte, 2)
				binary.LittleEndian.PutUint16(expectedMethodLen, uint16(len(testCase.attestationRequest.RequestMethod)))
				assert.Equal(t, proofData[6:8], expectedMethodLen)

				expectedResponseFormatLen := make([]byte, 2)
				binary.LittleEndian.PutUint16(expectedResponseFormatLen, 1)
				assert.Equal(t, proofData[8:10], expectedResponseFormatLen)

				expectedUrlLen := make([]byte, 2)
				binary.LittleEndian.PutUint16(expectedUrlLen, uint16(len(testCase.attestationRequest.Url)))
				assert.Equal(t, proofData[10:12], expectedUrlLen)

				expectedSelectorLen := make([]byte, 2)
				binary.LittleEndian.PutUint16(expectedSelectorLen, uint16(len(testCase.attestationRequest.Selector)))
				assert.Equal(t, proofData[12:14], expectedSelectorLen)

				expectedEncodingOptionsLen := make([]byte, 2)
				binary.LittleEndian.PutUint16(expectedEncodingOptionsLen, uint16(16))
				assert.Equal(t, proofData[14:16], expectedEncodingOptionsLen)

				encodedHeaders := encoding.EncodeHeaders(testCase.attestationRequest.RequestHeaders)

				expectedRequestHeadersLen := make([]byte, 2)
				binary.LittleEndian.PutUint16(expectedRequestHeadersLen, uint16(len(encodedHeaders)))
				assert.Equal(t, proofData[16:18], expectedRequestHeadersLen)

				encodedOptionalFields, _ := encoding.EncodeOptionalFields(testCase.attestationRequest.HTMLResultType, testCase.attestationRequest.RequestContentType, testCase.attestationRequest.RequestBody)

				expectedOptionalFieldsLen := make([]byte, 2)
				binary.LittleEndian.PutUint16(expectedOptionalFieldsLen, uint16(len(encodedOptionalFields)))
				assert.Equal(t, proofData[18:20], expectedOptionalFieldsLen)

			}
		})

	}
}

func TestPrepareEncodedRequestProof(t *testing.T) {
	// t.Skip()
	testCases := []struct {
		name             string
		userData         []byte
		encodedPositions encoding.ProofPositionalInfo
		expectedError    *appErrors.AppError
	}{
		{
			name:          "valid user data and encoded positions",
			userData:      bytes.Repeat([]byte("1"), 100),
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
			name:     "invalid user data and encoded positions",
			userData: bytes.Repeat([]byte("1"), 10),
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
			expectedError: appErrors.NewAppError(appErrors.ErrUserDataTooShort),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			encodedRequestProof, err := PrepareEncodedRequestProof(testCase.userData, testCase.encodedPositions)
			if testCase.expectedError != nil {
				assert.Equal(t, testCase.expectedError, err)
			} else {
				assert.Nil(t, err)
				metaHeaderLen := 2 * encoding.TARGET_ALIGNMENT
				endOffset := metaHeaderLen + (testCase.encodedPositions.Data.Len+testCase.encodedPositions.Timestamp.Len)*encoding.TARGET_ALIGNMENT
				clear(testCase.userData[metaHeaderLen:endOffset])
				assert.Equal(t, testCase.userData, encodedRequestProof)
			}
		})
	}

}
