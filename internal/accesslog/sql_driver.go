package accesslog

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
)

// WrapDriver 包装 ent dialect.Driver，将 SQL 与耗时写入请求 context 中的 Recorder。
func WrapDriver(d dialect.Driver) dialect.Driver {
	return &loggingDriver{Driver: d}
}

type loggingDriver struct {
	dialect.Driver
}

func (d *loggingDriver) Exec(ctx context.Context, query string, args, v any) error {
	start := time.Now()
	err := d.Driver.Exec(ctx, query, args, v)
	record(ctx, query, args, start, err)
	return err
}

func (d *loggingDriver) Query(ctx context.Context, query string, args, v any) error {
	start := time.Now()
	err := d.Driver.Query(ctx, query, args, v)
	record(ctx, query, args, start, err)
	return err
}

func (d *loggingDriver) Tx(ctx context.Context) (dialect.Tx, error) {
	tx, err := d.Driver.Tx(ctx)
	if err != nil {
		return nil, err
	}
	return &loggingTx{tx: tx, ctx: ctx}, nil
}

type loggingTx struct {
	tx  dialect.Tx
	ctx context.Context
}

func (t *loggingTx) Exec(ctx context.Context, query string, args, v any) error {
	start := time.Now()
	err := t.tx.Exec(ctx, query, args, v)
	record(t.ctx, query, args, start, err)
	return err
}

func (t *loggingTx) Query(ctx context.Context, query string, args, v any) error {
	start := time.Now()
	err := t.tx.Query(ctx, query, args, v)
	record(t.ctx, query, args, start, err)
	return err
}

func (t *loggingTx) Commit() error   { return t.tx.Commit() }
func (t *loggingTx) Rollback() error { return t.tx.Rollback() }

func record(ctx context.Context, query string, args any, start time.Time, err error) {
	rec := RecorderFromContext(ctx)
	if rec == nil {
		return
	}
	rec.AddSQL(query, fmt.Sprint(args), time.Since(start), err)
}
