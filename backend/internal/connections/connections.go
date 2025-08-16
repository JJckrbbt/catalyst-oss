package connections

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Client holds the database connection pool.
type Client struct {
	Pool *pgxpool.Pool
}

// ConnectDB establishes a connection to the PostgreSQL database.
func ConnectDB(databaseURL string, logger *slog.Logger) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	logger.Info("Database connection established")
	return &Client{Pool: pool}, nil
}

// Close gracefully closes the database connection pool.
func (c *Client) Close() {
	c.Pool.Close()
}

// Ping verifies the connection to the database is still alive.
func (c *Client) Ping() error {
	return c.Pool.Ping(context.Background())
}
