package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/kokuroshesh/bugvay/internal/config"
)

type ClickHouseDB struct {
	Conn driver.Conn
}

func NewClickHouse(cfg *config.ClickHouseConfig) (*ClickHouseDB, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},
		TLS: &tls.Config{
			InsecureSkipVerify: true,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout:      time.Second * 10,
		MaxOpenConns:     10,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Hour,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	})
	if err != nil {
		return nil, fmt.Errorf("open connection: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ping ClickHouse: %w", err)
	}

	return &ClickHouseDB{Conn: conn}, nil
}

func (db *ClickHouseDB) Close() error {
	return db.Conn.Close()
}

func (db *ClickHouseDB) Health(ctx context.Context) error {
	return db.Conn.Ping(ctx)
}
