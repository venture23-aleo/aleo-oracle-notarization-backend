package attestation

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	aleoUtil "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/aleoutil"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
)

// This test demonstrates how to test for panics in Go using assert.Panics from testify.
// We log the start and end of each test case for clarity.
func TestPrepareOracleReport(t *testing.T) {
	testCases := []struct {
		name                 string
		quote                []byte
		expectedError        *appErrors.AppError
		expectedOracleReport []byte
		expectPanic          bool
	}{
		{
			name:          "valid quote size - 10 bytes",
			quote:         []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a},
			expectedError: nil,
			expectedOracleReport: func() []byte {
				aleoContext, _ := aleoUtil.GetAleoContext()
				oracleReport, _ := aleoContext.GetSession().FormatMessage([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a}, constants.OracleReportChunkSize)
				return oracleReport
			}(),
			expectPanic: false,
		},
		{
			name:          "valid quote size - 1000 bytes",
			quote:         []byte(strings.Repeat("0", 1000)),
			expectedError: nil,
			expectPanic:   false,
			expectedOracleReport: func() []byte {
				aleoContext, _ := aleoUtil.GetAleoContext()
				oracleReport, _ := aleoContext.GetSession().FormatMessage([]byte(strings.Repeat("0", 1000)), constants.OracleReportChunkSize)
				return oracleReport
			}(),
		},
		{
			name:          "max quote size for 10 chunks with 32 fields per each - 5120 bytes",
			quote:         []byte(strings.Repeat("0", 5120)),
			expectedError: nil,
			expectPanic:   false,
			expectedOracleReport: func() []byte {
				aleoContext, _ := aleoUtil.GetAleoContext()
				oracleReport, _ := aleoContext.GetSession().FormatMessage([]byte(strings.Repeat("0", 5120)), constants.OracleReportChunkSize)
				return oracleReport
			}(),
		},
		{
			name:          "invalid quote size - 5121 bytes exceeds max size for 10 chunks with 32 fields per each",
			quote:         []byte(strings.Repeat("0", 5121)),
			expectedError: appErrors.ErrFormattingQuote,
			expectPanic:   false,
			expectedOracleReport: func() []byte {
				aleoContext, _ := aleoUtil.GetAleoContext()
				oracleReport, _ := aleoContext.GetSession().FormatMessage([]byte(strings.Repeat("0", 5120)), constants.OracleReportChunkSize)
				return oracleReport
			}(),
		},
		{
			name:                 "invalid quote size - 1000000 bytes",
			quote:                []byte(strings.Repeat("0", 1000000)),
			expectedError:        appErrors.ErrFormattingQuote,
			expectPanic:          false,
			expectedOracleReport: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			report, err := PrepareOracleReport(tc.quote)
			if tc.expectedError != nil {
				assert.Equal(t, tc.expectedError, err, "Expected error did not match for test case: %s", tc.name)
			} else {
				assert.Nil(t, err, "Expected no error for test case: %s", tc.name)
				assert.Equal(t, tc.expectedOracleReport, report, "Expected oracle report did not match for test case: %s", tc.name)
			}
		})
	}
}

func TestPrepareOracleSignature(t *testing.T) {

	testCases := []struct {
		name              string
		oracleReport      []byte
		expectedError     *appErrors.AppError
		expectedSignature string
	}{
		{
			name:          "valid oracle report - 1 chunks and 1 field",
			oracleReport:  []byte(`{  c0: {    f0: 47390263963055590408705u128 }}`),
			expectedError: nil,
		},
		{
			name:          "valid oracle report - 1 chunk and 21 fields",
			oracleReport:  []byte(`{  c0: {    f0: 47390263963055590408705u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128 }}`),
			expectedError: nil,
		},
		{
			name:              "valid oracle report with 10 chunks and 32 fields per each",
			oracleReport:      []byte(`{  c0: {    f0: 47390263963055590408705u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c1: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c2: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c3: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c4: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c5: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c6: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c7: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c8: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c9: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  }}`),
			expectedError:     nil,
			expectedSignature: "0x0102030405060708090a",
		},
		{
			name:          "invalid oracle report",
			oracleReport:  []byte(`{  c0: {    f0: 47390263963055590408705u128`),
			expectedError: appErrors.ErrHashingReport,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := PrepareOracleSignature(testCase.oracleReport)
			assert.Equal(t, testCase.expectedError, err)
			// assert.Equal(t, testCase.expectedSignature, signature)
		})
	}
}

func TestGenerateAttestationHash(t *testing.T) {

	testCases := []struct {
		name          string
		userData      []byte
		expectedError *appErrors.AppError
		expectedHash  string
	}{
		{
			name:          "valid user data format - 1 chunk and 1 field",
			userData:      []byte(`{ c0: { f0: 0u128 }}`),
			expectedHash:  "d3bb547a08a4140c61f269e551ae2327",
			expectedError: nil,
		},
		{
			name:          "valid user data format - 1 chunk and 21 fields",
			userData:      []byte(`{ c0: { f0: 0u128, f1: 0u128, f2: 0u128, f3: 0u128, f4: 0u128, f5: 0u128, f6: 0u128, f7: 0u128, f8: 0u128, f9: 0u128, f10: 0u128, f11: 0u128, f12: 0u128, f13: 0u128, f14: 0u128, f15: 0u128, f16: 0u128, f17: 0u128, f18: 0u128, f19: 0u128, f20: 0u128, f21: 0u128 }}`),
			expectedHash:  "9a7a17da060347f2e74a4cb285c1e6c5",
			expectedError: nil,
		},
		{
			name:          "invalid user data format - testing",
			userData:      []byte(`{ c0: { f0: 0u128, f1: 0u128`),
			expectedHash:  "",
			expectedError: appErrors.ErrCreatingAttestationHash,
		},
		{
			name:          "invalid user data format - testing",
			userData:      []byte(`testing`),
			expectedHash:  "",
			expectedError: appErrors.ErrCreatingAttestationHash,
		},
		{
			name:          "invalid user data format - 100 empty bytes",
			userData:      []byte(strings.Repeat("0", 100)),
			expectedError: appErrors.ErrCreatingAttestationHash,
			expectedHash:  "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			hash, err := GenerateAttestationHash(testCase.userData)
			assert.Equal(t, testCase.expectedError, err)
			assert.Equal(t, testCase.expectedHash, hex.EncodeToString(hash))
		})
	}
}

func TestBuildCompleteOracleData(t *testing.T) {
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

				aleoContext, err := aleoUtil.GetAleoContext()
				if err != nil {
					t.Fatalf("failed to get aleo context: %v", err)
				}

				switch testCase.attestationRequest.Url {
				case constants.PriceFeedBTCURL:
					assert.Equal(t, quotePreparationData.UserDataProof[0:1], []byte{0xc})
				case constants.PriceFeedETHURL:
					assert.Equal(t, quotePreparationData.UserDataProof[0:1], []byte{0xb})
				case constants.PriceFeedAleoURL:
					assert.Equal(t, quotePreparationData.UserDataProof[0:1], []byte{0x8})
				}

				expectedUserData, _ := aleoContext.GetSession().FormatMessage(quotePreparationData.UserDataProof, 8)
				assert.Equal(t, expectedUserData, quotePreparationData.UserData)

				expectedAttestationHash, _ := aleoContext.GetSession().HashMessage(quotePreparationData.UserData)
				assert.Equal(t, expectedAttestationHash, quotePreparationData.AttestationHash)

				quote := createMockSGXQuote(quotePreparationData.AttestationHash, false)

				oracleData, err := BuildCompleteOracleData(quotePreparationData, quote)

				assert.Nil(t, err)
				assert.NotEmpty(t, oracleData)
				assert.NotEmpty(t, oracleData.EncodedPositions)
				assert.NotEmpty(t, oracleData.EncodedRequest)
				assert.NotEmpty(t, oracleData.RequestHash)
				assert.NotEmpty(t, oracleData.TimestampedRequestHash)
				assert.NotEmpty(t, oracleData.Address)
				assert.NotEmpty(t, oracleData.Report)

				expectedOracleReport, _ := aleoContext.GetSession().FormatMessage(quote, 10)
				assert.Equal(t, string(expectedOracleReport), oracleData.Report)

				expectedRequestHashString, _ := aleoContext.GetSession().HashMessageToString([]byte(oracleData.EncodedRequest))
				assert.Equal(t, expectedRequestHashString, oracleData.RequestHash)

				expectedRequestHash, _ := aleoContext.GetSession().HashMessage([]byte(oracleData.EncodedRequest))
				expectedTimestampedRequestHash, _ := PrepareOracleTimestampedRequestHash(expectedRequestHash, quotePreparationData.Timestamp)
				assert.Equal(t, expectedTimestampedRequestHash, oracleData.TimestampedRequestHash)

				oracleReportHash, _ := aleoContext.GetSession().HashMessage(expectedOracleReport)

				_, signError := aleoContext.Sign(oracleReportHash)
				assert.Nil(t, signError)
			}
		})
	}
}
