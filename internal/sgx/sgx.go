package sgx

import (
	"sync"
)

// SGX Constants
const (
	// Debug flag mask
	DebugFlagMask = 0x02
	// Report body size
	SGXReportBodySize = 384

	// SGX report size
	SGXReportSize = 432

	// Report data size
	SGXReportDataSize = 64

	// File permissions
	SGXFilePermissions = 0644
)

// enclaveLock is the lock for the enclave to generate the sgx report and quote.
var enclaveLock sync.Mutex

// GetEnclaveLock returns the enclave lock.
func GetEnclaveLock() *sync.Mutex {
	return &enclaveLock
}

type AttestationPaths struct {
	MyTargetInfoPath    string
	TargetInfoPath      string
	UserReportDataPath  string
	ReportPath          string
	AttestationTypePath string
	QuotePath           string
}

// gramineAttestationPaths holds the default Gramine attestation paths.
var gramineAttestationPaths = AttestationPaths{
	MyTargetInfoPath:    "/dev/attestation/my_target_info",
	TargetInfoPath:      "/dev/attestation/target_info",
	UserReportDataPath:  "/dev/attestation/user_report_data",
	ReportPath:          "/dev/attestation/report",
	AttestationTypePath: "/dev/attestation/attestation_type",
	QuotePath:           "/dev/attestation/quote",
}

// GetGramineAttestationPaths returns the Gramine attestation paths.
func GetGramineAttestationPaths() AttestationPaths {
	return gramineAttestationPaths
}
