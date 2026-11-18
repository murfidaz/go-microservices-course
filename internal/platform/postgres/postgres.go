// Package postgres service terhubung ke PostgreSQL melalui sebuah pgx connection pool.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool membuat sebuah connection pool untuk database dalam dsn, dan mengecek apakah database dapat dihubungi sebelum dikembalikan.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse dsn: %w", err)
	}

	// ukuran Pool untuk minimal dan maksimal perlu diperhatikan
	cfg.MinConns = 2
	cfg.MaxConns = 10

	// daur ulang koneksi sekarang dan nanti, jadi load akan rebalce ketika terjadi failover.
	// jitter memberhentkan setiap koneksi dari reconnection pada saat yang bersamaan.
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnLifetimeJitter = 5 * time.Minute
	cfg.HealthCheckPeriod = 30 * time.Second
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return pool, nil
}
