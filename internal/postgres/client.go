package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"card_manager/api_service/ent"
	"card_manager/api_service/internal/config"
	"card_manager/api_service/internal/logger"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/lib/pq"
	"go.uber.org/fx"
)

type Client struct {
	ent *ent.Client
	db  *sql.DB
}

func NewEntClient(lc fx.Lifecycle, cfg *config.Config, log *logger.Logger) (*Client, error) {
	db, err := sql.Open("postgres", cfg.Postgres.DSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	entClient := ent.NewClient(ent.Driver(drv))

	c := &Client{ent: entClient, db: db}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info(ctx, "running ent schema migrate",
				"host", cfg.Postgres.Host,
				"dbname", cfg.Postgres.DBName,
			)
			if err := c.ent.Schema.Create(ctx); err != nil {
				return fmt.Errorf("ent migrate: %w", err)
			}
			log.Info(ctx, "postgres connected and migrated")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info(ctx, "closing postgres")
			return c.Close()
		},
	})
	return c, nil
}

func (c *Client) Ent() *ent.Client {
	if c == nil {
		return nil
	}
	return c.ent
}

func (c *Client) DB() *sql.DB {
	if c == nil {
		return nil
	}
	return c.db
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	var err error
	if c.ent != nil {
		err = c.ent.Close()
	}
	if c.db != nil {
		if e := c.db.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}
