package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/goplan/backend/internal/config"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(cfg *config.DatabaseConfig) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = time.Hour

	maxIdleTime, err := time.ParseDuration(cfg.MaxIdleTime)
	if err != nil {
		maxIdleTime = 30 * time.Minute
	}
	poolConfig.MaxConnIdleTime = maxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) Health(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Stats returns the connection pool statistics.
func (db *DB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}

// PoolStats holds connection pool statistics for monitoring and health checks.
type PoolStats struct {
	AcquireCount         int64  `json:"acquire_count"`
	AcquireDurationMs    int64  `json:"acquire_duration_ms"`
	AcquiredConns        int32  `json:"acquired_conns"`
	CanceledAcquireCount int64  `json:"canceled_acquire_count"`
	ConstructingConns    int32  `json:"constructing_conns"`
	EmptyAcquireCount    int64  `json:"empty_acquire_count"`
	IdleConns            int32  `json:"idle_conns"`
	MaxConns             int32  `json:"max_conns"`
	TotalConns           int32  `json:"total_conns"`
	NewConnsCount        int64  `json:"new_conns_count"`
	MaxLifetimeDestroy   int64  `json:"max_lifetime_destroy_count"`
	MaxIdleDestroy       int64  `json:"max_idle_destroy_count"`
}

// PoolStatsSummary returns a structured snapshot of the connection pool statistics.
// This can be used by health endpoints or monitoring dashboards.
func (db *DB) PoolStatsSummary() *PoolStats {
	s := db.Pool.Stat()
	return &PoolStats{
		AcquireCount:         s.AcquireCount(),
		AcquireDurationMs:    s.AcquireDuration().Milliseconds(),
		AcquiredConns:        s.AcquiredConns(),
		CanceledAcquireCount: s.CanceledAcquireCount(),
		ConstructingConns:    s.ConstructingConns(),
		EmptyAcquireCount:    s.EmptyAcquireCount(),
		IdleConns:            s.IdleConns(),
		MaxConns:             s.MaxConns(),
		TotalConns:           s.TotalConns(),
		NewConnsCount:        s.NewConnsCount(),
		MaxLifetimeDestroy:   s.MaxLifetimeDestroyCount(),
		MaxIdleDestroy:       s.MaxIdleDestroyCount(),
	}
}
