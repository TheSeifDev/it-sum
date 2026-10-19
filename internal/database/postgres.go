package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig holds database pool configuration parameters.
type PoolConfig struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

// DefaultPoolConfig returns sensible defaults suitable for Neon and free-tier hosting.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxConns:          5,
		MinConns:          0,
		MaxConnLifetime:   1 * time.Hour,
		MaxConnIdleTime:   5 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
		ConnectTimeout:    5 * time.Second,
	}
}

// NewPool initializes a reusable PostgreSQL connection pool with sensible limits for Neon.
// It verifies connectivity with a bounded ping and returns the connection pool.
// On failure, it returns a sanitized error that does not leak connection strings or credentials.
func NewPool(ctx context.Context, databaseURL string, customCfg ...PoolConfig) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, errors.New("database URL cannot be empty")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		// Do not include err or databaseURL directly to prevent leaking credentials
		return nil, errors.New("invalid database URL format")
	}

	cfg := DefaultPoolConfig()
	if len(customCfg) > 0 {
		cfg = customCfg[0]
	}

	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = cfg.MaxConnLifetime
	config.MaxConnIdleTime = cfg.MaxConnIdleTime
	config.HealthCheckPeriod = cfg.HealthCheckPeriod
	config.ConnConfig.ConnectTimeout = cfg.ConnectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errors.New("failed to initialize database pool")
	}

	// Verify initial connectivity with a bounded timeout
	pingCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, errors.New("failed to connect to database: ping unsuccessful")
	}

	return pool, nil
}

// Connect is a convenience wrapper around NewPool.
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	return NewPool(ctx, databaseURL)
}
