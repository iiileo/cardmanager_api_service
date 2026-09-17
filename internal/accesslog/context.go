package accesslog

import (
	"context"
	"sync"
	"time"
)

type ctxKey int

const recorderKey ctxKey = 1

// SQLRecord 单条 SQL 执行记录。
type SQLRecord struct {
	Query      string  `json:"query"`
	Args       string  `json:"args,omitempty"`
	DurationMs float64 `json:"duration_ms"`
	Error      string  `json:"error,omitempty"`
}

// Recorder 绑定在请求 context 上，收集该请求内的 SQL。
type Recorder struct {
	mu   sync.Mutex
	sqls []SQLRecord
}

func WithRecorder(ctx context.Context) (context.Context, *Recorder) {
	rec := &Recorder{}
	return context.WithValue(ctx, recorderKey, rec), rec
}

func RecorderFromContext(ctx context.Context) *Recorder {
	if rec, ok := ctx.Value(recorderKey).(*Recorder); ok {
		return rec
	}
	return nil
}

func (r *Recorder) AddSQL(query, args string, d time.Duration, err error) {
	if r == nil {
		return
	}
	rec := SQLRecord{
		Query:      query,
		Args:       args,
		DurationMs: float64(d.Microseconds()) / 1000,
	}
	if err != nil {
		rec.Error = err.Error()
	}
	r.mu.Lock()
	r.sqls = append(r.sqls, rec)
	r.mu.Unlock()
}

func (r *Recorder) Snapshot() []SQLRecord {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	out := append([]SQLRecord(nil), r.sqls...)
	r.mu.Unlock()
	return out
}
