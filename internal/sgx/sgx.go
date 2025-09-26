package sgx

import (
	"fmt"
	"sync"
	"syscall"
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

	// Attestation type
	AttestationType = "dcap"

	// https://github.com/intel/SGXDataCenterAttestationPrimitives/blob/a46ee8ab10569962c5cd7397b4babd4a47431976/QuoteVerification/QvE/Include/sgx_qve_def.h#L95
	QuoteMinSize = 1020

	ModeRead               = "read"
	ModeWrite              = "write"
	GraminePseudoFilesRoot = "/dev"
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

func EnforceSGXStartup() error {
	for _, path := range []string{
		gramineAttestationPaths.MyTargetInfoPath,
		gramineAttestationPaths.TargetInfoPath,
		gramineAttestationPaths.UserReportDataPath,
		gramineAttestationPaths.QuotePath,
		gramineAttestationPaths.ReportPath,
		gramineAttestationPaths.AttestationTypePath,
	} {
		fd, err := SecureOpenFile(GraminePseudoFilesRoot, path, ModeRead)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}
		syscall.Close(fd)
	}

	attType, err := SecureReadFile(GraminePseudoFilesRoot, gramineAttestationPaths.AttestationTypePath)
	if err != nil {
		return fmt.Errorf("reading attestation_type: %w", err)
	}

	if string(attType) != AttestationType {
		return fmt.Errorf("attestation type is not %s", AttestationType)
	}

	sgxReport, reportErr := GenerateSGXReport()
	if reportErr != nil {
		return fmt.Errorf("generating SGX report: %w", reportErr)
	}

	parsedSgxReport, reportErr := ParseSGXReport(sgxReport)
	if reportErr != nil {
		return fmt.Errorf("parsing SGX report: %w", reportErr)
	}

	isDebug := parsedSgxReport.Body.Attributes.Flags&DebugFlagMask == DebugFlagMask

	if isDebug {
		return fmt.Errorf("SGX is in debug mode")
	}

	_, quoteErr := GenerateQuote([]byte("test"))
	if quoteErr != nil {
		return fmt.Errorf("generating quote: %w", quoteErr)
	}

	return nil
}

// GetGramineAttestationPaths returns the Gramine attestation paths.
func GetGramineAttestationPaths() AttestationPaths {
	return gramineAttestationPaths
}
