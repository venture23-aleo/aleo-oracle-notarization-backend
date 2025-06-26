package configs

import (
	"os"
	"strings"
)

// WHITELISTED_DOMAINS is the list of whitelisted domains. It's a comma separated list of domains.
var WHITELISTED_DOMAINS = []string(strings.Split(os.Getenv("WHITELISTED_DOMAINS"), ","))
