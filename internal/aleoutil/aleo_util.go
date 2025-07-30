package aleo

import (
	"fmt"
	"sync"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
	aleoUtils "github.com/venture23-aleo/aleo-utils-go"
)

type AleoPublicContext interface {
	GetSession() aleoUtils.Session       // Get the Aleo session.
	GetPublicKey() string                // Get the Aleo public key.
	Sign(message []byte) (string, error) // Sign a message.
}

type AleoContext struct {
	Session    aleoUtils.Session // The Aleo session.
	privateKey string            // The Aleo private key.
	PublicKey  string            // The Aleo public key.
	Close      func()            // The Aleo close function.
}

// GetSession returns the Aleo session.
func (a *AleoContext) GetSession() aleoUtils.Session {
	return a.Session
}

// GetPublicKey returns the Aleo public key.
func (a *AleoContext) GetPublicKey() string {
	return a.PublicKey
}

// Sign signs a message.
func (a *AleoContext) Sign(message []byte) (string, error) {
	return a.Session.Sign(a.privateKey, message)
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
		Session:    s,
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
	var initErr error
	m.once.Do(func() {
		// Initialize once
		aleoCtx, err := newAleoContext()
		if err != nil {
			initErr = err
			return
		}
		m.context = aleoCtx
		m.initialized = true
		logger.Debug("Aleo context initialized successfully")
	})

	if initErr != nil {
		return nil, appErrors.ErrAleoContext.WithDetails(initErr.Error())
	}
	return m.context, nil
}

// Initialize explicitly initializes the Aleo context
func InitAleoContext() error {
	var initErr error
	aleoManager.once.Do(func() {
		aleoCtx, err := newAleoContext()
		if err != nil {
			initErr = err
			return
		}
		aleoManager.context = aleoCtx
		aleoManager.initialized = true
		logger.Debug("Aleo context initialized successfully")
	})
	return initErr
}

// ShutdownAleoContext properly closes the Aleo context
func ShutdownAleoContext() error {
	aleoManager.mu.Lock()
	defer aleoManager.mu.Unlock()

	if aleoManager.initialized && aleoManager.context != nil {
		if aleoCtx, ok := aleoManager.context.(*AleoContext); ok && aleoCtx.Close != nil {
			aleoCtx.Close()
		}
		aleoManager.initialized = false
	}
	return nil
}

// IsInitialized checks if the Aleo context has been initialized
func IsAleoContextInitialized() bool {
	aleoManager.mu.RLock()
	defer aleoManager.mu.RUnlock()
	return aleoManager.initialized
}
