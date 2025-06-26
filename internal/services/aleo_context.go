package services

import (
	"fmt"

	aleo "github.com/zkportal/aleo-utils-go"
)

type AleoPublicContext interface {
    GetSession() aleo.Session
    GetPublicKey() string
	Sign(message []byte) (string, error)
}

type AleoContext struct {
	Session aleo.Session
	privateKey string
	PublicKey  string
	Close    func() 
}

func (a *AleoContext) GetSession() aleo.Session {
	return a.Session
}

func (a *AleoContext) GetPublicKey() string {
	return a.PublicKey
}

func (a *AleoContext) Sign(message []byte) (string,error) {
	return a.Session.Sign(a.privateKey, message)
}

func (a *AleoContext) String() string {
	return fmt.Sprintf("AleoContext{Session: <hidden>, PublicKey: %s}", a.PublicKey)
}

// NewAleoContext creates a new Aleo context with a session and a private key.
func NewAleoContext() (*AleoContext, error) {

	// Create a new wrapper.
	wrapper, closeFn, err := aleo.NewWrapper()
	if err != nil {
		return nil, err
	}

	// Create a new session.
	s, err := wrapper.NewSession()
	if err != nil {
		return nil,err
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
		Close: closeFn,
	}, nil		
}