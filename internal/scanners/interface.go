package scanners

import (
	"context"
)

// Scanner interface for all scanner modules
type Scanner interface {
	Scan(ctx context.Context, input *ScanInput) (*ScanResult, error)
	Name() string
}

type ScanInput struct {
	EndpointID int
	URL        string
	Method     string
	Headers    map[string]string
	Body       string
}

type ScanResult struct {
	Vulnerable bool
	Severity   string
	CWE        int
	Evidence   map[string]interface{}
	Proof      string
	Confidence float64
}
