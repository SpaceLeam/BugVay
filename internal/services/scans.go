package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kokuroshesh/bugvay/internal/database"
	"github.com/kokuroshesh/bugvay/internal/queue"
)

type ScanService struct {
	pg *database.PostgresDB
	ch *database.ClickHouseDB
	q  *queue.Client
}

type ScanRequest struct {
	EndpointIDs []int    `json:"endpoint_ids"`
	Scanners    []string `json:"scanners"`
	Concurrency int      `json:"concurrency"`
	RateLimit   int      `json:"rate_limit"`
}

type Scan struct {
	ID          string    `json:"id"`
	ProgramID   int       `json:"program_id"`
	Status      string    `json:"status"`
	Scanners    []string  `json:"scanners"`
	JobsTotal   int       `json:"jobs_total"`
	JobsSuccess int       `json:"jobs_success"`
	JobsFailed  int       `json:"jobs_failed"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewScanService(pg *database.PostgresDB, ch *database.ClickHouseDB, q *queue.Client) *ScanService {
	return &ScanService{pg: pg, ch: ch, q: q}
}

func (s *ScanService) CreateScan(ctx context.Context, req *ScanRequest) (*Scan, error) {
	// Validate scanners
	validScanners := map[string]bool{
		"xss": true, "sqli": true, "lfi": true, "redirect": true,
	}

	for _, scanner := range req.Scanners {
		if !validScanners[scanner] {
			return nil, fmt.Errorf("invalid scanner: %s", scanner)
		}
	}

	// Enqueue jobs
	jobIDs := []string{}
	for _, endpointID := range req.EndpointIDs {
		for _, scanner := range req.Scanners {
			payload, _ := json.Marshal(map[string]interface{}{
				"endpoint_id": endpointID,
				"scanner":     scanner,
			})

			jobID, err := s.q.EnqueueScan(ctx, scanner, endpointID, payload)
			if err != nil {
				return nil, fmt.Errorf("enqueue job: %w", err)
			}
			jobIDs = append(jobIDs, jobID)
		}
	}

	scan := &Scan{
		ID:        fmt.Sprintf("scan_%s", generateScanID()),
		Status:    "running",
		Scanners:  req.Scanners,
		JobsTotal: len(jobIDs),
		CreatedAt: time.Now(),
	}

	return scan, nil
}

// generateScanID creates a short UUID for scan identification
func generateScanID() string {
	// Use timestamp + random for shorter IDs (16 chars)
	return fmt.Sprintf("%d%04x", time.Now().Unix(), time.Now().Nanosecond()%0xFFFF)
}

func (s *ScanService) GetScanStatus(ctx context.Context, scanID string) (*Scan, error) {
	// TODO: Query job status from Asynq/Redis
	// For now, return mock data
	return &Scan{
		ID:          scanID,
		Status:      "running",
		JobsTotal:   100,
		JobsSuccess: 75,
		JobsFailed:  5,
	}, nil
}
