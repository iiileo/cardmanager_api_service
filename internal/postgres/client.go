package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"card_manager/api_service/internal/config"
	"card_manager/api_service/internal/logger"

	_ "github.com/lib/pq"
)

// Client is a thin SQL wrapper. Ent wiring will replace/extend this when schemas exist.
type Client struct {
	db *sql.DB
}

type IClient interface {
	DB() *sql.DB
	Ping(ctx context.Context) error
	Close() error
}

func NewClient(cfg *config.Config, log *logger.Logger) (IClient, error) {
	db, err := sql.Open("postgres", cfg.Postgres.DSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	log.Info(context.Background(), "postgres connected",
		"host", cfg.Postgres.Host,
		"dbname", cfg.Postgres.DBName,
	)
	return &Client{db: db}, nil
}

func (c *Client) DB() *sql.DB {
	if c == nil {
		return nil
	}
	return c.db
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.db == nil {
		return fmt.Errorf("postgres client is nil")
	}
	return c.db.PingContext(ctx)
}

func (c *Client) Close() error {
	if c == nil || c.db == nil {
		return nil
	}
	return c.db.Close()
}
