package attestation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	encoding "github.com/venture23-aleo/aleo-oracle-encoding"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/constants"
	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	aleoContext "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/services/aleo_context"
)

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
			expectedError:   appErrors.NewAppError(appErrors.ErrPreparingProofData),
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
			expectedError:   appErrors.NewAppError(appErrors.ErrPreparingProofData),
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
		},
		{
			name: "valid attestation data for float encoding with eth price feed url",
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
		},
		{
			name: "valid attestation data for float encoding with aleo price feed url",
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
			expectedError:   appErrors.NewAppError(appErrors.ErrPreparingProofData),
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

				aleoContext, err := aleoContext.GetAleoContext()
				if err != nil {
					t.Fatalf("failed to get aleo context: %v", err)
				}

				switch testCase.attestationRequest.Url {
				case constants.PRICE_FEED_BTC_URL:
					assert.Equal(t, quotePreparationData.UserDataProof[0:1], []byte{0xc})
				case constants.PRICE_FEED_ETH_URL:
					assert.Equal(t, quotePreparationData.UserDataProof[0:1], []byte{0xb})
				case constants.PRICE_FEED_ALEO_URL:
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

				signature, signError := aleoContext.Sign(oracleReportHash)
				assert.Nil(t, signError)
				assert.Equal(t, signature, oracleData.Signature)
			}
		})
	}
}
