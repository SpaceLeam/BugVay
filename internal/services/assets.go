package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kokuroshesh/bugvay/internal/database"
)

type AssetService struct {
	pg *database.PostgresDB
}

type Asset struct {
	ID        int       `json:"id"`
	ProgramID int       `json:"program_id"`
	Domain    string    `json:"domain"`
	Type      string    `json:"type"`
	Origin    string    `json:"origin"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAssetRequest struct {
	ProgramID int    `json:"program_id" binding:"required"`
	Domain    string `json:"domain" binding:"required"`
	Type      string `json:"type" binding:"required"` // subdomain, wildcard, url
	Origin    string `json:"origin"`
}

func NewAssetService(pg *database.PostgresDB) *AssetService {
	return &AssetService{pg: pg}
}

func (s *AssetService) CreateAsset(ctx context.Context, req *CreateAssetRequest) (*Asset, error) {
	// Validate type
	validTypes := map[string]bool{
		"subdomain": true,
		"wildcard":  true,
		"url":       true,
	}
	if !validTypes[req.Type] {
		return nil, fmt.Errorf("invalid asset type: %s (must be: subdomain, wildcard, url)", req.Type)
	}

	var asset Asset
	err := s.pg.Pool.QueryRow(ctx, `
		INSERT INTO assets (program_id, domain, type, origin)
		VALUES ($1, $2, $3, $4)
		RETURNING id, program_id, domain, type, origin, created_at
	`, req.ProgramID, req.Domain, req.Type, req.Origin).Scan(
		&asset.ID, &asset.ProgramID, &asset.Domain, &asset.Type, &asset.Origin, &asset.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("create asset: %w", err)
	}

	return &asset, nil
}

func (s *AssetService) GetAsset(ctx context.Context, id int) (*Asset, error) {
	var a Asset
	err := s.pg.Pool.QueryRow(ctx, `
		SELECT id, program_id, domain, type, origin, created_at
		FROM assets WHERE id = $1
	`, id).Scan(&a.ID, &a.ProgramID, &a.Domain, &a.Type, &a.Origin, &a.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("asset not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query asset: %w", err)
	}

	return &a, nil
}

func (s *AssetService) ListAssets(ctx context.Context, programID int, limit, offset int) ([]Asset, error) {
	query := `
		SELECT id, program_id, domain, type, origin, created_at
		FROM assets
	`
	args := []interface{}{}

	if programID > 0 {
		query += " WHERE program_id = $1"
		args = append(args, programID)
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprint(len(args)+1) + " OFFSET $" + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.pg.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query assets: %w", err)
	}
	defer rows.Close()

	var assets []Asset
	for rows.Next() {
		var a Asset
		if err := rows.Scan(&a.ID, &a.ProgramID, &a.Domain, &a.Type, &a.Origin, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		assets = append(assets, a)
	}

	return assets, nil
}

func (s *AssetService) DeleteAsset(ctx context.Context, id int) error {
	result, err := s.pg.Pool.Exec(ctx, "DELETE FROM assets WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete asset: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("asset not found")
	}

	return nil
}
