package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/kokuroshesh/bugvay/internal/config"
	"github.com/kokuroshesh/bugvay/internal/database"
	"github.com/kokuroshesh/bugvay/internal/httpclient"
	"github.com/kokuroshesh/bugvay/internal/queue"
	"github.com/kokuroshesh/bugvay/internal/scanners"
	"github.com/kokuroshesh/bugvay/internal/scanners/xss"
	"github.com/kokuroshesh/bugvay/internal/services"
)

type Worker struct {
	server          *asynq.Server
	mux             *asynq.ServeMux
	pg              *database.PostgresDB
	ch              *database.ClickHouseDB
	httpClient      *httpclient.Scanner
	findingService  *services.FindingService
	endpointService *services.EndpointService
}

func NewWorker(cfg *config.Config, pg *database.PostgresDB, ch *database.ClickHouseDB) *Worker {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     cfg.Redis.Addr(),
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		},
		asynq.Config{
			Concurrency: cfg.Worker.Concurrency,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	httpClient := httpclient.NewScanner(cfg.Worker.RateLimit, time.Duration(cfg.Scanner.Timeout)*time.Second)
	findingService := services.NewFindingService(pg, ch)
	endpointService := services.NewEndpointService(pg, ch, nil)

	w := &Worker{
		server:          srv,
		mux:             asynq.NewServeMux(),
		pg:              pg,
		ch:              ch,
		httpClient:      httpClient,
		findingService:  findingService,
		endpointService: endpointService,
	}

	w.registerHandlers()
	return w
}

func (w *Worker) registerHandlers() {
	w.mux.HandleFunc(queue.TypeScanXSS, w.handleXSSScan)
	w.mux.HandleFunc(queue.TypeScanSQLi, w.handleSQLiScan)
	w.mux.HandleFunc(queue.TypeScanLFI, w.handleLFIScan)
	w.mux.HandleFunc(queue.TypeScanRedirect, w.handleRedirectScan)
}

func (w *Worker) handleXSSScan(ctx context.Context, task *asynq.Task) error {
	var payload queue.ScanPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Get endpoint details
	endpoint, err := w.endpointService.GetEndpoint(ctx, payload.EndpointID)
	if err != nil {
		return err
	}

	// Run XSS scanner
	scanner := xss.New(w.httpClient)
	result, err := scanner.Scan(ctx, &scanners.ScanInput{
		EndpointID: payload.EndpointID,
		URL:        endpoint.URL,
		Method:     "GET",
	})
	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Save finding if vulnerable
	if result.Vulnerable {
		finding := &services.Finding{
			EndpointID: payload.EndpointID,
			Scanner:    "xss",
			Severity:   result.Severity,
			CWE:        result.CWE,
			Evidence:   result.Evidence,
			Proof:      result.Proof,
			Status:     "new",
		}

		if err := w.findingService.CreateFinding(ctx, finding); err != nil {
			log.Printf("Failed to save finding: %v", err)
		}
	}

	return nil
}

func (w *Worker) handleSQLiScan(ctx context.Context, task *asynq.Task) error {
	// TODO: Implement SQLi scanner
	log.Println("SQLi scan not implemented yet")
	return nil
}

func (w *Worker) handleLFIScan(ctx context.Context, task *asynq.Task) error {
	// TODO: Implement LFI scanner
	log.Println("LFI scan not implemented yet")
	return nil
}

func (w *Worker) handleRedirectScan(ctx context.Context, task *asynq.Task) error {
	// TODO: Implement redirect scanner
	log.Println("Redirect scan not implemented yet")
	return nil
}

func (w *Worker) Run() error {
	return w.server.Run(w.mux)
}

func (w *Worker) Shutdown() {
	w.server.Shutdown()
}
