package configs

import (
	"os"
	"strings"
)

var PrivateKey, PublicKey string

var WHITELISTED_DOMAINS = []string(strings.Split(os.Getenv("WHITELISTED_DOMAINS"), ","))
