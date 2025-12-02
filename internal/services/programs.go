package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kokuroshesh/bugvay/internal/database"
)

type ProgramService struct {
	pg *database.PostgresDB
}

type Program struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateProgramRequest struct {
	Name string `json:"name" binding:"required"`
}

func NewProgramService(pg *database.PostgresDB) *ProgramService {
	return &ProgramService{pg: pg}
}

func (s *ProgramService) CreateProgram(ctx context.Context, req *CreateProgramRequest) (*Program, error) {
	var program Program
	err := s.pg.Pool.QueryRow(ctx, `
		INSERT INTO programs (name) VALUES ($1)
		RETURNING id, name, created_at
	`, req.Name).Scan(&program.ID, &program.Name, &program.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("create program: %w", err)
	}

	return &program, nil
}

func (s *ProgramService) GetProgram(ctx context.Context, id int) (*Program, error) {
	var p Program
	err := s.pg.Pool.QueryRow(ctx, `
		SELECT id, name, created_at FROM programs WHERE id = $1
	`, id).Scan(&p.ID, &p.Name, &p.CreatedAt)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("program not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query program: %w", err)
	}

	return &p, nil
}

func (s *ProgramService) ListPrograms(ctx context.Context, limit, offset int) ([]Program, error) {
	rows, err := s.pg.Pool.Query(ctx, `
		SELECT id, name, created_at FROM programs
		ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query programs: %w", err)
	}
	defer rows.Close()

	var programs []Program
	for rows.Next() {
		var p Program
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		programs = append(programs, p)
	}

	return programs, nil
}
