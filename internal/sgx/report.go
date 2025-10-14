package sgx

import (
	"bytes"
	"encoding/binary"
	"os"

	appErrors "github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/errors"
	"github.com/venture23-aleo/aleo-oracle-notarization-backend/internal/logger"
)

// SGX report body is structured as follows:
/*
| Field        | Offset | Size (bytes) |
|--------------|--------|--------------|
| CPUSVN       | 0      | 16           |
| MISCSELECT   | 16     | 4            |
| RESERVED1    | 20     | 28           |
| ATTRIBUTES   | 48     | 16           |
|   ├─ FLAGS   | 48     | 8            |
|   └─ XFRM    | 56     | 8            |
| MRENCLAVE    | 64     | 32           |
| RESERVED2    | 96     | 32           |
| MRSIGNER     | 128    | 32           |
| RESERVED3    | 160    | 96           |
| ISVPRODID    | 256    | 2            |
| ISVSVN       | 258    | 2            |
| RESERVED4    | 260    | 60           |
| REPORTDATA   | 320    | 64           |
|--------------|--------|--------------|
| **Total**    | —      | **384 bytes** |

*/

type Attributes struct {
	Flags uint64
	Xfrm  uint64
}

type ReportBody struct {
	CPUSVN     [16]byte
	MiscSelect [4]byte
	_          [28]byte // Reserved1
	Attributes Attributes
	MREnclave  [32]byte
	_          [32]byte // Reserved2
	MRSigner   [32]byte
	_          [96]byte // Reserved3
	ISVProdID  [2]byte
	ISVSVN     [2]byte
	_          [60]byte // Reserved4
	ReportData [64]byte
}

type SGXReport struct {
	Body  ReportBody
	KeyID [32]byte
	MAC   [16]byte
}

// GenerateSGXReport generates the SGX report.
// Refer to https://gramine.readthedocs.io/en/stable/attestation.html#low-level-dev-attestation-interface
func GenerateSGXReport() ([]byte, *appErrors.AppError) {

	enclaveLock.Lock()
	defer enclaveLock.Unlock()

	// Read the target info from target info path
	targetInfo, err := os.ReadFile(gramineAttestationPaths.MyTargetInfoPath)
	if err != nil {
		logger.Error("Error reading target info: ", "error", err)
		return nil, appErrors.ErrReadingTargetInfo
	}

	// Write the target info to the target info path
	if err := os.WriteFile(gramineAttestationPaths.TargetInfoPath, targetInfo, SGXFilePermissions); err != nil {
		logger.Error("Error writing target info: ", "error", err)
		return nil, appErrors.ErrWritingTargetInfo
	}

	// Create and write report data
	reportData := make([]byte, SGXReportDataSize)
	if err := os.WriteFile(gramineAttestationPaths.UserReportDataPath, reportData, SGXFilePermissions); err != nil {
		logger.Error("Error writing report data: ", "error", err)
		return nil, appErrors.ErrWritingReportData
	}

	// Read the report from the report path
	report, err := os.ReadFile(gramineAttestationPaths.ReportPath)
	if err != nil {
		logger.Error("Error reading report: ", "error", err)
		return nil, appErrors.ErrReadingReport
	}

	logger.Debug("Attestation environment prepared successfully")
	return report, nil
}

func ParseSGXReport(rawSgxReport []byte) (*SGXReport, *appErrors.AppError) {
	if len(rawSgxReport) != SGXReportSize {
		return nil, appErrors.ErrInvalidSGXReportSize
	}

	var report SGXReport
	reader := bytes.NewReader(rawSgxReport)

	err := binary.Read(reader, binary.LittleEndian, &report)
	if err != nil {
		return nil, appErrors.ErrParsingSGXReport
	}

	return &report, nil
}
