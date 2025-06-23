package configs

import (
	"os"
	"strings"
)

// PrivateKey and PublicKey holds key pair generated at the runtime by Aleo Wrapper
var PrivateKey, PublicKey string

// WHITELISTED_DOMAINS is the list of whitelisted domains. It's a comma separated list of domains.
var WHITELISTED_DOMAINS = []string(strings.Split(os.Getenv("WHITELISTED_DOMAINS"), ","))
