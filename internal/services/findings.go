package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kokuroshesh/bugvay/internal/database"
)

type FindingService struct {
	pg *database.PostgresDB
	ch *database.ClickHouseDB
}

type Finding struct {
	ID         int                    `json:"id"`
	EndpointID int                    `json:"endpoint_id"`
	Scanner    string                 `json:"scanner"`
	Severity   string                 `json:"severity"`
	CWE        int                    `json:"cwe,omitempty"`
	Evidence   map[string]interface{} `json:"evidence"`
	Proof      string                 `json:"proof"`
	Status     string                 `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
}

type TriageRequest struct {
	Status        string `json:"status"`
	FalsePositive bool   `json:"false_positive"`
}

func NewFindingService(pg *database.PostgresDB, ch *database.ClickHouseDB) *FindingService {
	return &FindingService{pg: pg, ch: ch}
}

func (s *FindingService) CreateFinding(ctx context.Context, f *Finding) error {
	_, err := s.pg.Pool.Exec(ctx, `
		INSERT INTO findings (endpoint_id, scanner, severity, cwe, evidence, proof, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, f.EndpointID, f.Scanner, f.Severity, f.CWE, f.Evidence, f.Proof, f.Status)

	return err
}

func (s *FindingService) GetFinding(ctx context.Context, id int) (*Finding, error) {
	var f Finding
	err := s.pg.Pool.QueryRow(ctx, `
		SELECT id, endpoint_id, scanner, severity, cwe, evidence, proof, status, created_at
		FROM findings WHERE id = $1
	`, id).Scan(&f.ID, &f.EndpointID, &f.Scanner, &f.Severity, &f.CWE, &f.Evidence, &f.Proof, &f.Status, &f.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("finding not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query finding: %w", err)
	}

	return &f, nil
}

func (s *FindingService) ListFindings(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]Finding, error) {
	query := `
		SELECT id, endpoint_id, scanner, severity, COALESCE(cwe, 0), evidence, proof, status, created_at
		FROM findings
		WHERE 1=1
	`
	args := []interface{}{}
	argPos := 1

	if severity, ok := filters["severity"].(string); ok && severity != "" {
		query += fmt.Sprintf(" AND severity = $%d", argPos)
		args = append(args, severity)
		argPos++
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, status)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := s.pg.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query findings: %w", err)
	}
	defer rows.Close()

	var findings []Finding
	for rows.Next() {
		var f Finding
		if err := rows.Scan(&f.ID, &f.EndpointID, &f.Scanner, &f.Severity, &f.CWE, &f.Evidence, &f.Proof, &f.Status, &f.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		findings = append(findings, f)
	}

	return findings, nil
}

func (s *FindingService) TriageFinding(ctx context.Context, id int, req *TriageRequest) error {
	_, err := s.pg.Pool.Exec(ctx, `
		UPDATE findings
		SET status = $1, false_positive = $2, resolved_at = CASE WHEN $1 = 'closed' THEN NOW() ELSE NULL END
		WHERE id = $3
	`, req.Status, req.FalsePositive, id)

	return err
}
