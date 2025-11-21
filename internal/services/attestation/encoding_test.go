package attestation

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	"github.com/venture23-aleo/aleo-oracle-encoding/positionRecorder"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/common"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
)

func TestPrepareProofData_WithPositionalInfo(t *testing.T) {
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
			expectedError:          appErrors.ErrEncodingAttestationData,
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
			expectedError:          appErrors.ErrInvalidEncodingOption,
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
			expectedError:          appErrors.ErrAttestationDataTooLarge,
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
			expectedError:          appErrors.ErrEncodingAttestationData,
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
			expectedError:          appErrors.ErrEncodingAttestationData,
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
			expectedError:          appErrors.ErrAttestationDataTooLarge,
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
				EncodingOptions: encoding.EncodingOptions{
					Value:     "",
					Precision: 0,
				},
			},
			statusCode:             200,
			attestationData:        string(bytes.Repeat([]byte("1"), 16)),
			timestamp:              1715769600,
			expectedError:          appErrors.ErrInvalidEncodingOption,
			expectedPositionalInfo: nil,
		},
	}

	for _, testCase := range testCases {

		t.Run(testCase.name, func(t *testing.T) {
			t.Logf("testCase.attestationData: %v", len(testCase.attestationData))
			aleoBlockHeight, blockHeightError := common.GetAleoCurrentBlockHeight()
			if blockHeightError != nil {
				t.Fatalf("failed to get aleo block height: %v", blockHeightError)
			}
			proofData, encodedPositions, err := PrepareProofData(testCase.statusCode, testCase.attestationData, testCase.timestamp, int64(aleoBlockHeight), testCase.attestationRequest)
			if testCase.expectedError != nil {
				assert.Equal(t, testCase.expectedError, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, proofData)
				assert.NotNil(t, encodedPositions)
				assert.Equal(t, testCase.expectedPositionalInfo, encodedPositions)

				var attestationData = make([]byte, 2)
				var attestationDataLen int

				if common.IsPriceFeedURL(testCase.attestationRequest.Url) {
					attestationDataLen = len(testCase.attestationData)
				} else {
					switch testCase.attestationRequest.EncodingOptions.Value {
					case "float":
						attestationDataLen = math.MaxUint8
					case "string":
						attestationDataLen = constants.AttestationDataSizeLimit
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
			expectedError: appErrors.ErrUserDataTooShort,
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

func TestPrepareAttestationData(t *testing.T) {
	testCases := []struct {
		name            string
		attestationData string
		encodingOptions encoding.EncodingOptions
		expectedResult  string
		expectedError   *appErrors.AppError
		description     string
		statusCode      int
		timestamp       int64
	}{
		// String encoding tests
		{
			name:            "string encoding - short data",
			attestationData: "hello world",
			encodingOptions: encoding.EncodingOptions{
				Value: "string",
			},
			expectedResult: func() string {
				// Pad with null bytes to AttestationDataSizeLimit
				padded := "hello world"
				for len(padded) < constants.AttestationDataSizeLimit {
					padded += "\x00"
				}
				return padded
			}(),
			expectedError: nil,
			description:   "String encoding should pad with null bytes to AttestationDataSizeLimit",
			statusCode:    200,
			timestamp:     1715769600,
		},
		{
			name:            "string encoding - empty data",
			attestationData: "",
			encodingOptions: encoding.EncodingOptions{
				Value: "string",
			},
			expectedResult: func() string {
				// Pad with null bytes to AttestationDataSizeLimit
				padded := ""
				for len(padded) < constants.AttestationDataSizeLimit {
					padded += "\x00"
				}
				return padded
			}(),
			expectedError: nil,
			description:   "Empty string should be padded with null bytes to AttestationDataSizeLimit",
			statusCode:    200,
			timestamp:     1715769600,
		},
		{
			name:            "string encoding - exact size data",
			attestationData: string(bytes.Repeat([]byte("a"), constants.AttestationDataSizeLimit)),
			encodingOptions: encoding.EncodingOptions{
				Value: "string",
			},
			expectedResult: string(bytes.Repeat([]byte("a"), constants.AttestationDataSizeLimit)),
			expectedError:  nil,
			description:    "String of exact size should remain unchanged",
			statusCode:     200,
			timestamp:      1715769600,
		},

		// Float encoding tests
		{
			name:            "float encoding - with decimal point",
			attestationData: "123.45",
			encodingOptions: encoding.EncodingOptions{
				Value:     "float",
				Precision: 2,
			},
			expectedResult: func() string {
				// Pad with '0' to MaxUint8 length
				padded := "123.45"
				for len(padded) < math.MaxUint8 {
					padded += "0"
				}
				return padded
			}(),
			expectedError: nil,
			description:   "Float with decimal point should be padded with '0' to MaxUint8 length",
			statusCode:    200,
			timestamp:     1715769600,
		},
		{
			name:            "float encoding - without decimal point",
			attestationData: "123",
			encodingOptions: encoding.EncodingOptions{
				Value:     "float",
				Precision: 0,
			},
			expectedResult: func() string {
				// Add decimal point and pad with '0' to MaxUint8 length
				padded := "123."
				for len(padded) < math.MaxUint8 {
					padded += "0"
				}
				return padded
			}(),
			expectedError: nil,
			description:   "Float without decimal point should add '.' and pad with '0' to MaxUint8 length",
			statusCode:    200,
			timestamp:     1715769600,
		},
		{
			name:            "float encoding - zero value",
			attestationData: "0",
			encodingOptions: encoding.EncodingOptions{
				Value:     "float",
				Precision: 0,
			},
			expectedResult: func() string {
				// Add decimal point and pad with '0' to MaxUint8 length
				padded := "0."
				for len(padded) < math.MaxUint8 {
					padded += "0"
				}
				return padded
			}(),
			expectedError: nil,
			description:   "Zero value should add '.' and pad with '0' to MaxUint8 length",
			statusCode:    200,
			timestamp:     1715769600,
		},
		{
			name:            "float encoding - negative value",
			attestationData: "-123.45",
			encodingOptions: encoding.EncodingOptions{
				Value: "float",
			},
			expectedResult: "",
			expectedError:  appErrors.ErrEncodingAttestationData,
			description:    "Negative float should cause encoding error",
			statusCode:     200,
			timestamp:      1715769600,
		},

		// Integer encoding tests
		{
			name:            "int encoding - positive number",
			attestationData: "123",
			encodingOptions: encoding.EncodingOptions{
				Value: "int",
			},
			expectedResult: func() string {
				// Prepend '0' characters to reach MaxUint8 length
				padLength := math.MaxUint8 - len("123")
				padString := ""
				for i := 0; i < padLength; i++ {
					padString += "0"
				}
				return padString + "123"
			}(),
			expectedError: nil,
			description:   "Integer should be prepended with '0' to reach MaxUint8 length",
			statusCode:    200,
			timestamp:     1715769600,
		},
		{
			name:            "int encoding - zero value",
			attestationData: "0",
			encodingOptions: encoding.EncodingOptions{
				Value: "int",
			},
			expectedResult: func() string {
				// Prepend '0' characters to reach MaxUint8 length
				padLength := math.MaxUint8 - len("0")
				padString := ""
				for i := 0; i < padLength; i++ {
					padString += "0"
				}
				return padString + "0"
			}(),
			expectedError: nil,
			description:   "Zero integer should be prepended with '0' to reach MaxUint8 length",
			statusCode:    200,
			timestamp:     1715769600,
		},
		{
			name:            "int encoding - negative number",
			attestationData: "-123",
			encodingOptions: encoding.EncodingOptions{
				Value: "int",
			},
			expectedResult: "",
			expectedError:  appErrors.ErrEncodingAttestationData,
			description:    "Negative integer should cause encoding error",
			statusCode:     200,
			timestamp:      1715769600,
		},
		{
			name:            "int encoding - large number",
			attestationData: "999999999",
			encodingOptions: encoding.EncodingOptions{
				Value: "int",
			},
			expectedResult: func() string {
				// Prepend '0' characters to reach MaxUint8 length
				padLength := math.MaxUint8 - len("999999999")
				padString := ""
				for i := 0; i < padLength; i++ {
					padString += "0"
				}
				return padString + "999999999"
			}(),
			expectedError: nil,
			description:   "Large integer should be prepended with '0' to reach MaxUint8 length",
			statusCode:    200,
			timestamp:     1715769600,
		},

		// Edge cases
		{
			name:            "int encoding - maximum length number",
			attestationData: string(bytes.Repeat([]byte("9"), math.MaxUint8)),
			encodingOptions: encoding.EncodingOptions{
				Value: "int",
			},
			expectedResult: string(bytes.Repeat([]byte("9"), math.MaxUint8)),
			expectedError:  appErrors.ErrEncodingAttestationData,
			description:    "Integer of maximum length should cause encoding error",
			statusCode:     200,
			timestamp:      1715769600,
		},
		{
			name:            "float encoding - maximum length with decimal",
			attestationData: string(bytes.Repeat([]byte("9"), math.MaxUint8-1)) + ".",
			encodingOptions: encoding.EncodingOptions{
				Value: "float",
			},
			expectedResult: string(bytes.Repeat([]byte("9"), math.MaxUint8-1)) + ".",
			expectedError:  appErrors.ErrEncodingAttestationData,
			description:    "Float of maximum length with decimal point should cause encoding error",
			statusCode:     200,
			timestamp:      1715769600,
		},

		// Invalid encoding option
		{
			name:            "invalid encoding option",
			attestationData: "test data",
			encodingOptions: encoding.EncodingOptions{
				Value: "invalid",
			},
			expectedResult: "",
			expectedError:  appErrors.ErrInvalidEncodingOption,
			description:    "Invalid encoding option should return error",
			statusCode:     200,
			timestamp:      1715769600,
		},
		{
			name:            "empty encoding option",
			attestationData: "test data",
			encodingOptions: encoding.EncodingOptions{
				Value: "",
			},
			expectedResult: "",
			expectedError:  appErrors.ErrInvalidEncodingOption,
			description:    "Empty encoding option should return error",
			statusCode:     200,
			timestamp:      1715769600,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Test the function indirectly through PrepareProofData
			// We'll use a non-price-feed URL to ensure prepareAttestationData is called
			req := AttestationRequest{
				Url:             "google.com", // Non-price-feed URL
				RequestMethod:   "GET",
				ResponseFormat:  "json",
				Selector:        "body",
				EncodingOptions: testCase.encodingOptions,
			}

			aleoBlockHeight, blockHeightError := common.GetAleoCurrentBlockHeight()
			if blockHeightError != nil {
				t.Fatalf("failed to get aleo block height: %v", blockHeightError)
			}
			// Call PrepareProofData which internally calls prepareAttestationData
			result, _, err := PrepareProofData(testCase.statusCode, testCase.attestationData, testCase.timestamp, int64(aleoBlockHeight), req)

			if testCase.expectedError != nil {
				assert.Equal(t, testCase.expectedError, err, testCase.description)
				assert.Nil(t, result, "Result should be nil when error is expected")
			} else {
				assert.Nil(t, err, testCase.description)
				assert.NotNil(t, result, "Result should not be nil when no error is expected")

				// For valid cases, we can't easily extract the prepared attestation data
				// from the result buffer, so we just verify that no error occurred
				// and the result is not nil
			}
		})
	}
}

func TestPrepareAttestationData_EdgeCases(t *testing.T) {
	testCases := []struct {
		name            string
		attestationData string
		encodingOptions encoding.EncodingOptions
		expectedError   *appErrors.AppError
		description     string
	}{
		{
			name:            "string encoding - very long data",
			attestationData: string(bytes.Repeat([]byte("a"), constants.AttestationDataSizeLimit+100)),
			encodingOptions: encoding.EncodingOptions{
				Value: "string",
			},
			expectedError: nil, // Should be truncated/padded to exact size
			description:   "Very long string should be handled properly",
		},
		{
			name:            "float encoding - scientific notation",
			attestationData: "1.23e-4",
			encodingOptions: encoding.EncodingOptions{
				Value: "float",
			},
			expectedError: nil,
			description:   "Scientific notation should be handled as float",
		},
		{
			name:            "float encoding - multiple decimal points",
			attestationData: "123.45.67",
			encodingOptions: encoding.EncodingOptions{
				Value: "float",
			},
			expectedError: nil,
			description:   "Multiple decimal points should be handled",
		},
		{
			name:            "int encoding - hex string",
			attestationData: "0xFF",
			encodingOptions: encoding.EncodingOptions{
				Value: "int",
			},
			expectedError: appErrors.ErrEncodingAttestationData,
			description:   "Hex string should cause encoding error",
		},
		{
			name:            "int encoding - binary string",
			attestationData: "0b101010",
			encodingOptions: encoding.EncodingOptions{
				Value: "int",
			},
			expectedError: appErrors.ErrEncodingAttestationData,
			description:   "Binary string should cause encoding error",
		},
		{
			name:            "float encoding - negative number edge case",
			attestationData: "-0.001",
			encodingOptions: encoding.EncodingOptions{
				Value:     "float",
				Precision: 3,
			},
			expectedError: appErrors.ErrEncodingAttestationData,
			description:   "Negative float should cause encoding error",
		},
		{
			name:            "int encoding - negative number edge case",
			attestationData: "-999999",
			encodingOptions: encoding.EncodingOptions{
				Value: "int",
			},
			expectedError: appErrors.ErrEncodingAttestationData,
			description:   "Negative integer should cause encoding error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := AttestationRequest{
				Url:             "google.com",
				RequestMethod:   "GET",
				ResponseFormat:  "json",
				Selector:        "body",
				EncodingOptions: testCase.encodingOptions,
			}

			aleoBlockHeight := 224254
			_, _, err := PrepareProofData(200, testCase.attestationData, 1715769600, int64(aleoBlockHeight), req)

			if testCase.expectedError != nil {
				assert.Equal(t, testCase.expectedError, err, testCase.description)
			} else {
				// For edge cases, we mainly want to ensure no panic occurs
				// The actual behavior might vary depending on the encoding library
				assert.NotPanics(t, func() {
				aleoBlockHeight, blockHeightError := common.GetAleoCurrentBlockHeight()
				if blockHeightError != nil {
					t.Fatalf("failed to get aleo block height: %v", blockHeightError)
				}
					_, _, _ = PrepareProofData(200, testCase.attestationData, 1715769600, int64(aleoBlockHeight), req)
				}, testCase.description)
			}
		})
	}
}

func TestPrepareAttestationData_Performance(t *testing.T) {
	// Test performance with large data
	largeData := string(bytes.Repeat([]byte("a"), 1000))

	req := AttestationRequest{
		Url:            "google.com",
		RequestMethod:  "GET",
		ResponseFormat: "json",
		Selector:       "body",
		EncodingOptions: encoding.EncodingOptions{
			Value: "string",
		},
	}

	// Benchmark the function
	start := time.Now()
	for i := 0; i < 100; i++ {
		aleoBlockHeight, blockHeightError := common.GetAleoCurrentBlockHeight()
		if blockHeightError != nil {
			t.Fatalf("failed to get aleo block height: %v", blockHeightError)
		}
		_, _, err := PrepareProofData(200, largeData, 1715769600, int64(aleoBlockHeight), req)
		assert.Nil(t, err)
	}
	duration := time.Since(start)

	// Ensure it completes within reasonable time (1 second for 100 iterations)
	assert.Less(t, duration, time.Second, "Performance test should complete within 1 second for 100 iterations")
}

func TestPrepareAttestationData_PriceFeedExclusion(t *testing.T) {
	// Test that price feed URLs bypass prepareAttestationData
	testCases := []struct {
		name            string
		url             string
		attestationData string
		encodingOptions encoding.EncodingOptions
		description     string
	}{
		{
			name:            "price feed btc - should bypass prepareAttestationData",
			url:             "price_feed: btc",
			attestationData: "123.45",
			encodingOptions: encoding.EncodingOptions{
				Value:     "float",
				Precision: 2,
			},
			description: "Price feed URL should not call prepareAttestationData",
		},
		{
			name:            "price feed eth - should bypass prepareAttestationData",
			url:             "price_feed: eth",
			attestationData: "456.78",
			encodingOptions: encoding.EncodingOptions{
				Value:     "float",
				Precision: 2,
			},
			description: "Price feed URL should not call prepareAttestationData",
		},
		{
			name:            "price feed aleo - should bypass prepareAttestationData",
			url:             "price_feed: aleo",
			attestationData: "789.01",
			encodingOptions: encoding.EncodingOptions{
				Value:     "float",
				Precision: 2,
			},
			description: "Price feed URL should not call prepareAttestationData",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req := AttestationRequest{
				Url:             testCase.url,
				RequestMethod:   "GET",
				ResponseFormat:  "json",
				Selector:        "weightedAvgPrice",
				EncodingOptions: testCase.encodingOptions,
			}

			// For price feed URLs, the attestation data should be used as-is
			// without calling prepareAttestationData
			aleoBlockHeight, blockHeightError := common.GetAleoCurrentBlockHeight()
			if blockHeightError != nil {
				t.Fatalf("failed to get aleo block height: %v", blockHeightError)
			}
			result, _, err := PrepareProofData(200, testCase.attestationData, 1715769600, int64(aleoBlockHeight), req)

			assert.Nil(t, err, testCase.description)
			assert.NotNil(t, result, "Result should not be nil for price feed URLs")
		})
	}
}
