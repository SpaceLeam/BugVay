package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kokuroshesh/bugvay/internal/database"
	"github.com/kokuroshesh/bugvay/internal/queue"
)

type EndpointService struct {
	pg *database.PostgresDB
	ch *database.ClickHouseDB
	q  *queue.Client
}

type Endpoint struct {
	ID           int       `json:"id"`
	AssetID      int       `json:"asset_id"`
	URL          string    `json:"url"`
	CanonicalURL string    `json:"canonical_url"`
	Hash         string    `json:"hash"`
	Crawled      bool      `json:"crawled"`
	DiscoveredBy []string  `json:"discovered_by"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewEndpointService(pg *database.PostgresDB, ch *database.ClickHouseDB, q *queue.Client) *EndpointService {
	return &EndpointService{pg: pg, ch: ch, q: q}
}

func (s *EndpointService) CreateEndpoint(ctx context.Context, assetID int, rawURL, source string) (*Endpoint, error) {
	canonical := CanonicalizeURL(rawURL)
	hash := HashURL(canonical)

	var endpoint Endpoint
	err := s.pg.Pool.QueryRow(ctx, `
		INSERT INTO endpoints (asset_id, url, canonical_url, hash, discovered_by)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (hash) DO UPDATE 
		SET discovered_by = array_append(endpoints.discovered_by, $5)
		RETURNING id, asset_id, url, canonical_url, hash, crawled, discovered_by, created_at
	`, assetID, rawURL, canonical, hash, []string{source}).Scan(
		&endpoint.ID, &endpoint.AssetID, &endpoint.URL, &endpoint.CanonicalURL,
		&endpoint.Hash, &endpoint.Crawled, &endpoint.DiscoveredBy, &endpoint.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("insert endpoint: %w", err)
	}

	return &endpoint, nil
}

func (s *EndpointService) GetEndpoint(ctx context.Context, id int) (*Endpoint, error) {
	var e Endpoint
	err := s.pg.Pool.QueryRow(ctx, `
		SELECT id, asset_id, url, canonical_url, hash, crawled, discovered_by, created_at
		FROM endpoints WHERE id = $1
	`, id).Scan(&e.ID, &e.AssetID, &e.URL, &e.CanonicalURL, &e.Hash, &e.Crawled, &e.DiscoveredBy, &e.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("endpoint not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query endpoint: %w", err)
	}

	return &e, nil
}

func (s *EndpointService) ListEndpoints(ctx context.Context, assetID int, limit, offset int) ([]Endpoint, error) {
	query := `
		SELECT id, asset_id, url, canonical_url, hash, crawled, discovered_by, created_at
		FROM endpoints
	`
	args := []interface{}{}

	if assetID > 0 {
		query += " WHERE asset_id = $1"
		args = append(args, assetID)
	}

	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprint(len(args)+1) + " OFFSET $" + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.pg.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query endpoints: %w", err)
	}
	defer rows.Close()

	var endpoints []Endpoint
	for rows.Next() {
		var e Endpoint
		if err := rows.Scan(&e.ID, &e.AssetID, &e.URL, &e.CanonicalURL, &e.Hash, &e.Crawled, &e.DiscoveredBy, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		endpoints = append(endpoints, e)
	}

	return endpoints, nil
}

// CanonicalizeURL normalizes URLs for deduplication
func CanonicalizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Remove fragment
	u.Fragment = ""

	// Sort query parameters
	if u.RawQuery != "" {
		params := u.Query()
		keys := make([]string, 0, len(params))
		for k := range params {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		newParams := url.Values{}
		for _, k := range keys {
			newParams[k] = params[k]
		}
		u.RawQuery = newParams.Encode()
	}

	// Normalize path
	u.Path = strings.TrimSuffix(u.Path, "/")

	return u.String()
}

// HashURL generates consistent hash for URL deduplication
func HashURL(canonicalURL string) string {
	h := sha256.New()
	h.Write([]byte(canonicalURL))
	return hex.EncodeToString(h.Sum(nil))[:16]
}
