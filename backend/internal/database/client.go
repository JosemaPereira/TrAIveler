// Package database provides PostgreSQL connection pool management using pgx/v5.
// It handles connection lifecycle, health checks, and graceful shutdown.
//
// This package follows Dependency Injection principles with an interface-first
// approach, allowing easy swapping of implementations for testing or different
// database providers.
package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Client defines the interface for database operations.
//
// Example usage:
//
//	var dbClient database.Client
//	dbClient, err := database.NewClient(ctx, databaseURL)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer dbClient.Close()
//
//	// Health check
//	if err := dbClient.Ping(ctx); err != nil {
//	    log.Fatal(err)
//	}
//
//	// Execute queries
//	pool := dbClient.Pool()
//	row := pool.QueryRow(ctx, "SELECT version()")
type Client interface {
	// Ping verifies the database connection is healthy by executing SELECT 1.
	Ping(ctx context.Context) error

	// Close gracefully shuts down the connection pool.
	Close() error

	// Pool returns the underlying pgxpool.Pool for executing queries.
	Pool() *pgxpool.Pool
}

// pgxClient is the concrete implementation of the Client interface using pgx/v5.
type pgxClient struct {
	pool   *pgxpool.Pool
	closed bool
}

// NewClient creates a new database client with connection pooling.
// It attempts to connect with retry logic (3 attempts, 2-second delays).
// Connection pool settings:
//   - MinConns: 5 (minimum idle connections)
//   - MaxConns: 25 (maximum concurrent connections)
//   - MaxConnLifetime: 1 hour (connection reuse limit)
//   - MaxConnIdleTime: 30 minutes (idle connection timeout)
//
// Returns an error if connection fails after all retries.
// The returned Client interface allows for easy mocking and implementation swapping.
func NewClient(ctx context.Context, databaseURL string) (Client, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL cannot be empty")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MinConns = 5
	config.MaxConns = 25
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	const maxRetries = 3
	const retryDelay = 2 * time.Second

	var pool *pgxpool.Pool
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			lastErr = err
			slog.WarnContext(ctx, "database connection pool creation failed",
				"attempt", attempt,
				"max_retries", maxRetries,
				"error", err.Error(),
			)

			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("context canceled during connection retry: %w", ctx.Err())
				case <-time.After(retryDelay):
				}
			}
			continue
		}

		client := &pgxClient{pool: pool}
		err = client.Ping(ctx)
		if err == nil {
			slog.InfoContext(ctx, "database connection established",
				"attempt", attempt,
				"min_conns", config.MinConns,
				"max_conns", config.MaxConns,
			)

			return client, nil
		}

		pool.Close()
		lastErr = err
		slog.WarnContext(ctx, "database connection verification failed",
			"attempt", attempt,
			"max_retries", maxRetries,
			"error", err.Error(),
		)

		// Don't sleep after the last attempt
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context canceled during connection retry: %w", ctx.Err())
			case <-time.After(retryDelay):
			}
		}
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, lastErr)
}

// Ping verifies the database connection is healthy by executing SELECT 1.
// It respects the provided context for timeout/cancellation.
func (c *pgxClient) Ping(ctx context.Context) error {
	if c.closed {
		return fmt.Errorf("connection pool is closed")
	}

	var result int
	err := c.pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("ping returned unexpected result: %d", result)
	}

	return nil
}

// Close gracefully shuts down the connection pool, closing all connections.
// It should be called during application shutdown to prevent connection leaks.
func (c *pgxClient) Close() error {
	if c.pool == nil || c.closed {
		return nil
	}

	c.pool.Close()
	c.closed = true
	slog.Info("database connection pool closed")

	return nil
}

// Pool returns the underlying pgxpool.Pool for direct access when needed.
// This allows repositories to execute queries using the pool.
func (c *pgxClient) Pool() *pgxpool.Pool {
	return c.pool
}
