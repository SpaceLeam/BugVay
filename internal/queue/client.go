package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/kokuroshesh/bugvay/internal/config"
)

const (
	TypeScanXSS      = "scan:xss"
	TypeScanSQLi     = "scan:sqli"
	TypeScanLFI      = "scan:lfi"
	TypeScanRedirect = "scan:redirect"
)

type Client struct {
	*asynq.Client
}

func NewClient(cfg *config.RedisConfig) *Client {
	client := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	return &Client{Client: client}
}

type ScanPayload struct {
	EndpointID int    `json:"endpoint_id"`
	Scanner    string `json:"scanner"`
	URL        string `json:"url"`
}

func (c *Client) EnqueueScan(ctx context.Context, scanner string, endpointID int, payload []byte) (string, error) {
	taskType := ""
	switch scanner {
	case "xss":
		taskType = TypeScanXSS
	case "sqli":
		taskType = TypeScanSQLi
	case "lfi":
		taskType = TypeScanLFI
	case "redirect":
		taskType = TypeScanRedirect
	default:
		return "", fmt.Errorf("unknown scanner: %s", scanner)
	}

	task := asynq.NewTask(taskType, payload,
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
		asynq.Queue("default"),
	)

	info, err := c.Enqueue(task, asynq.ProcessIn(1*time.Second))
	if err != nil {
		return "", fmt.Errorf("enqueue task: %w", err)
	}

	return info.ID, nil
}

func (c *Client) GetJobStatus(ctx context.Context, jobID string) (string, error) {
	// TODO: Query Asynq inspector for job status
	return "running", nil
}

func NewScanPayload(endpointID int, scanner, url string) ([]byte, error) {
	return json.Marshal(ScanPayload{
		EndpointID: endpointID,
		Scanner:    scanner,
		URL:        url,
	})
}
