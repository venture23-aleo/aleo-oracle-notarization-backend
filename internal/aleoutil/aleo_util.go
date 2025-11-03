package aleo

import (
	"fmt"
	"sync"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	aleoUtils "github.com/venture23-aleo/aleo-utils-go"
)

type AleoPublicContext interface {
	HashMessage(message []byte) ([]byte, error) // Hash a message.
	HashMessageToString(message []byte) (string, error) // Hash a message and return a string.
	FormatMessage(message []byte, chunkSize int) ([]byte, error) // Format a message.
	GetPublicKey() string                // Get the Aleo public key.
	Sign(message []byte) (string, error) // Sign a message.
}

type AleoContext struct {
	sessionLock sync.RWMutex      // Lock for thread safety.
	session    aleoUtils.Session // The Aleo session.
	privateKey []byte            // The Aleo private key.
	PublicKey  string            // The Aleo public key.
	Close      func()            // The Aleo close function.
}

// FormatMessage formats a message.
func (a *AleoContext) FormatMessage(message []byte, chunkSize int) ([]byte, error) {
	a.sessionLock.Lock()
	defer a.sessionLock.Unlock()
	return a.session.FormatMessage(message, chunkSize)
}

// HashMessage hashes a message.
func (a *AleoContext) HashMessage(message []byte) ([]byte, error) {
	a.sessionLock.Lock()
	defer a.sessionLock.Unlock()
	return a.session.HashMessage(message)
}

func (a *AleoContext) HashMessageToString(message []byte) (string, error) {
	a.sessionLock.Lock()
	defer a.sessionLock.Unlock()
	return a.session.HashMessageToString(message)
}

// GetPublicKey returns the Aleo public key.
func (a *AleoContext) GetPublicKey() string {
	return a.PublicKey
}

// Sign signs a message.
func (a *AleoContext) Sign(message []byte) (string, error) {
	if IsAleoContextInitialized() {
		a.sessionLock.Lock()
		defer a.sessionLock.Unlock()
		return a.session.Sign(a.privateKey, message)
	}
	return "", appErrors.ErrAleoContext.WithDetails("Aleo context is not initialized")
}

// String returns a string representation of the Aleo context.
func (a *AleoContext) String() string {
	return fmt.Sprintf("AleoContext{Session: <hidden>, PublicKey: %s}", a.PublicKey)
}

// newAleoContext creates a new Aleo context with a session and a private key.
func newAleoContext() (*AleoContext, error) {
	// Create a new wrapper.
	wrapper, closeFn, err := aleoUtils.NewWrapper()
	if err != nil {
		return nil, err
	}

	// Create a new session.
	s, err := wrapper.NewSession()
	if err != nil {
		return nil, err
	}

	// Generate a new private key.
	privKey, address, err := s.NewPrivateKey()
	if err != nil {
		return nil, err
	}

	return &AleoContext{
		sessionLock: sync.RWMutex{},
		session:    s,
		privateKey: privKey,
		PublicKey:  address,
		Close:      closeFn,
	}, nil
}

// AleoContextManager manages the singleton Aleo context.
type AleoContextManager struct {
	context     AleoPublicContext // The Aleo context.
	once        sync.Once         // Once for lazy initialization.
	mu          sync.RWMutex      // Mutex for thread safety.
	initialized bool              // Whether the Aleo context has been initialized.
}

// Global singleton instance.
var aleoManager = &AleoContextManager{}

// GetAleoContext returns the singleton Aleo context, initializing it if needed
func GetAleoContext() (AleoPublicContext, *appErrors.AppError) {
	return aleoManager.GetAleoContext()
}

// GetAleoContext returns the singleton Aleo context, initializing it if needed
func (m *AleoContextManager) GetAleoContext() (AleoPublicContext, *appErrors.AppError) {
	m.mu.RLock()
   	if m.initialized && m.context != nil {
        ctx := m.context
        m.mu.RUnlock()
        return ctx, nil
    }
    m.mu.RUnlock()
    
	var initErr error
	m.once.Do(func() {
		// Initialize once
		aleoCtx, err := newAleoContext()
		if err != nil {
			initErr = err
			return
		}
		m.mu.Lock()
		defer m.mu.Unlock()
		m.context = aleoCtx
		m.initialized = true
		logger.Debug("Aleo context initialized successfully")
	})

	if initErr != nil {
		return nil, appErrors.ErrAleoContext.WithDetails(initErr.Error())
	}

	m.mu.RLock()
	ctx := m.context
	m.mu.RUnlock()

	return ctx, nil
}

// Initialize explicitly initializes the Aleo context
func InitAleoContext() error {
    _, err := GetAleoContext()
	if err != nil {
		return err
	}
	return nil
}


// ShutdownAleoContext properly closes the Aleo context
func ShutdownAleoContext() error {
    m := aleoManager
    m.mu.Lock()
    defer m.mu.Unlock()

    if m.initialized && m.context != nil {
        if aleoCtx, ok := m.context.(*AleoContext); ok && aleoCtx.Close != nil {
            aleoCtx.Close()
        }
        m.context = nil
        m.initialized = false
    }
    return nil

}

// IsInitialized checks if the Aleo context has been initialized
func IsAleoContextInitialized() bool {
   	m := aleoManager
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.initialized
}
