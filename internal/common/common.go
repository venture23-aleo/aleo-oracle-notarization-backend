package common

import "sync"

// quoteLock is the lock for the quote.
var quoteLock sync.Mutex

func GetQuoteLock() *sync.Mutex {
	return &quoteLock
}
