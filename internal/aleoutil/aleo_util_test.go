package aleo

import (
	"bytes"
	"io"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

// TestMain initializes the logger for all tests in this package
func TestMain(m *testing.M) {
	// Initialize logger for tests
	logger.InitLogger("DEBUG")

	// Run the tests
	// os.Exit(m.Run())
	m.Run()
}

func TestInitAleoContext(t *testing.T) {
	assert.False(t, IsAleoContextInitialized())
	InitAleoContext()
	t.Logf("Aleo context initialized: %v", aleoManager.initialized)
	assert.True(t, IsAleoContextInitialized())
}

func TestCheckAleoContextProperties(t *testing.T) {
	assert.True(t, IsAleoContextInitialized())

	aleoCtx, err := GetAleoContext()
	assert.Nil(t, err)
	assert.NotNil(t, aleoCtx)
	assert.NotEmpty(t, aleoCtx.GetPublicKey())
	assert.True(t, aleoManager.initialized)
}

// Message length below 16 bytes should fail and message length with 16 bytes and above should succeed
func TestGenerateSignature(t *testing.T) {
	assert.True(t, IsAleoContextInitialized())

	aleoCtx, err := GetAleoContext()
	assert.Nil(t, err)
	assert.NotNil(t, aleoCtx)

	testCases := []struct {
		message       []byte
		expectedError bool
	}{
		{bytes.Repeat([]byte{1}, 1), true},
		{bytes.Repeat([]byte{1}, 100), false},
		{bytes.Repeat([]byte{1}, 15), true},
		{bytes.Repeat([]byte{1}, 16), false},
		{bytes.Repeat([]byte{1}, 17), false},
		{bytes.Repeat([]byte{1}, 8), true},
	}

	for _, testCase := range testCases {
		signature, signErr := aleoCtx.Sign(testCase.message)
		if testCase.expectedError {
			assert.NotNil(t, signErr)
			assert.Empty(t, signature)
		} else {
			assert.Nil(t, signErr)
			assert.NotEmpty(t, signature)
		}
	}

}

// TestAleoContext_FormatMessage tests the FormatMessage function. Target chunks is the number of chunks to split the message into. It should be with 1 and 32
func TestFormatMessage(t *testing.T) {
	assert.True(t, IsAleoContextInitialized())

	aleoCtx, err := GetAleoContext()
	assert.Nil(t, err)
	assert.NotNil(t, aleoCtx)

	testCases := []struct {
		message        []byte
		expectedString string
		name           string
		expectedError  bool
		targetChunks   int
	}{
		{
			message:        bytes.Repeat([]byte{1}, 8),
			expectedString: "",
			name:           "0 bytes",
			expectedError:  true,
			targetChunks:   0,
		},
		{
			message:        bytes.Repeat([]byte{1}, 8),
			expectedString: "{  c0: {    f0: 72340172838076673u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  }}",
			name:           "8 bytes",
			expectedError:  false,
			targetChunks:   1,
		},
		{
			message:        bytes.Repeat([]byte{1}, 16),
			expectedString: "{  c0: {    f0: 1334440654591915542993625911497130241u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  },  c1: {    f0: 0u128,    f1: 0u128,    f2: 0u128,    f3: 0u128,    f4: 0u128,    f5: 0u128,    f6: 0u128,    f7: 0u128,    f8: 0u128,    f9: 0u128,    f10: 0u128,    f11: 0u128,    f12: 0u128,    f13: 0u128,    f14: 0u128,    f15: 0u128,    f16: 0u128,    f17: 0u128,    f18: 0u128,    f19: 0u128,    f20: 0u128,    f21: 0u128,    f22: 0u128,    f23: 0u128,    f24: 0u128,    f25: 0u128,    f26: 0u128,    f27: 0u128,    f28: 0u128,    f29: 0u128,    f30: 0u128,    f31: 0u128  }}",
			name:           "16 bytes",
			expectedError:  false,
			targetChunks:   2,
		},
		{
			message:        bytes.Repeat([]byte{1}, 100),
			expectedString: "",
			name:           "100 bytes",
			expectedError:  true,
			targetChunks:   34,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			formattedMessage, formatErr := aleoCtx.FormatMessage(testCase.message, testCase.targetChunks)
			if testCase.expectedError {
				assert.NotNil(t, formatErr)
				assert.Empty(t, formattedMessage)
			} else {
				assert.Nil(t, formatErr)
				assert.NotEmpty(t, formattedMessage)
				assert.Equal(t, testCase.expectedString, string(formattedMessage))
			}
		})
	}

}

func TestHashMessage(t *testing.T) {
	assert.True(t, IsAleoContextInitialized())

	aleoCtx, err := GetAleoContext()
	assert.Nil(t, err)
	assert.NotNil(t, aleoCtx)

	testCases := []struct {
		message       []byte
		expectedHash  []byte
		expectedError bool
		name          string
	}{
		{
			message:       bytes.Repeat([]byte{1}, 16),
			expectedHash:  []byte{0x8a, 0x55, 0x2d, 0x99, 0xb2, 0xa4, 0x57, 0x58, 0x79, 0x8a, 0x48, 0x68, 0xb1, 0xc3, 0x35, 0x30},
			expectedError: false,
			name:          "16 bytes",
		},
		{
			message:       bytes.Repeat([]byte{1}, 100),
			expectedHash:  []byte{0x4c, 0x59, 0x99, 0x2c, 0x83, 0xe5, 0x18, 0x13, 0xe2, 0x5c, 0xc6, 0xe0, 0xbd, 0x6f, 0x14, 0x25},
			expectedError: false,
			name:          "100 bytes",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			formattedMessage, formatErr := aleoCtx.FormatMessage(testCase.message, 1)
			assert.Nil(t, formatErr)
			assert.NotEmpty(t, formattedMessage)

			hash, hashErr := aleoCtx.HashMessage(formattedMessage)
			assert.Nil(t, hashErr)
			assert.NotEmpty(t, hash)

			assert.Equal(t, testCase.expectedHash, hash)
		})
	}
}

func TestAleoContext_ParallelExecutionFormatMessage(t *testing.T) {
	aleoCtx, err := GetAleoContext()
	if err != nil {
		t.Fatalf("Failed to get Aleo context: %v", err)
	}
	t.Logf("Aleo context: %v", aleoCtx)

	wg := sync.WaitGroup{}
	wg.Add(1000)

	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			formattedMessage, formatErr := aleoCtx.FormatMessage(bytes.Repeat([]byte{1}, 16), 8)
			assert.Nil(t, formatErr)
			assert.NotEmpty(t, formattedMessage)
		}()
	}

	wg.Wait()
}

func TestAleoContext_ParallelExecutionHashMessage(t *testing.T) {
	aleoCtx, err := GetAleoContext()
	if err != nil {
		t.Fatalf("Failed to get Aleo context: %v", err)
	}
	t.Logf("Aleo context: %v", aleoCtx)

	// var mu sync.Mutex

	wg := sync.WaitGroup{}
	wg.Add(1000)
	
	log.SetOutput(io.Discard)      // disable logger
    defer log.SetOutput(os.Stderr) // restore if needed

	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			// mu.Lock()
			_, _ = aleoCtx.HashMessage(bytes.Repeat([]byte{1}, 16))
			// mu.Unlock()
		}()
	}

	wg.Wait()
}


func TestAleoContext_ShutdownAleoContext(t *testing.T) {
	assert.True(t, IsAleoContextInitialized())

	ShutdownAleoContext()
	assert.False(t, IsAleoContextInitialized())
}