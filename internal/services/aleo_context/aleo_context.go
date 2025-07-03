package aleo_context

import (
	"fmt"
	"log"
	"sync"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	aleo "github.com/venture23-aleo/aleo-utils-go"
)

type AleoPublicContext interface {
	GetSession() aleo.Session
	GetPublicKey() string
	Sign(message []byte) (string, error)
}

type AleoContext struct {
	Session    aleo.Session
	privateKey string
	PublicKey  string
	Close      func()
}

func (a *AleoContext) GetSession() aleo.Session {
	return a.Session
}

func (a *AleoContext) GetPublicKey() string {
	return a.PublicKey
}

func (a *AleoContext) Sign(message []byte) (string, error) {
	return a.Session.Sign(a.privateKey, message)
}

func (a *AleoContext) String() string {
	return fmt.Sprintf("AleoContext{Session: <hidden>, PublicKey: %s}", a.PublicKey)
}

// NewAleoContext creates a new Aleo context with a session and a private key.
func newAleoContext() (*AleoContext, error) {
	// Create a new wrapper.
	wrapper, closeFn, err := aleo.NewWrapper()
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

// AleoContextManager manages the singleton Aleo context
type AleoContextManager struct {
	context AleoPublicContext
	once    sync.Once
	mu      sync.RWMutex
	initialized bool
}

// Global singleton instance
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
		log.Println("Aleo context initialized successfully")
	})
	
	if initErr != nil {
		return nil, appErrors.NewAppErrorWithDetails(appErrors.ErrAleoContext, initErr.Error())
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
		log.Println("Aleo context initialized successfully")
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