package accesslog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"card_manager/api_service/internal/config"
	"card_manager/api_service/internal/logger"
	"go.uber.org/fx"
)

// Store 写入文件并保留最近记录在内存中供页面查询。
type Store struct {
	enabled   bool
	filePath  string
	maxMemory int
	maxBody   int
	skipPaths []string
	mu        sync.RWMutex
	entries   []Entry
	file      *os.File
}

func NewStore(lc fx.Lifecycle, cfg *config.Config, log *logger.Logger) (*Store, error) {
	al := cfg.AccessLog
	s := &Store{
		enabled:   al.Enabled,
		filePath:  al.FilePath,
		maxMemory: al.MaxMemory,
		maxBody:   al.MaxBodyBytes,
		skipPaths: append([]string(nil), al.SkipPaths...),
	}
	if s.maxMemory <= 0 {
		s.maxMemory = 200
	}
	if s.maxBody <= 0 {
		s.maxBody = 8192
	}
	if !s.enabled {
		return s, nil
	}
	if s.filePath == "" {
		s.filePath = "data/access.jsonl"
	}

	if err := os.MkdirAll(filepath.Dir(s.filePath), 0o755); err != nil {
		return nil, fmt.Errorf("access log mkdir: %w", err)
	}
	f, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("access log open file: %w", err)
	}
	s.file = f

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info(ctx, "closing access log file")
			if s.file != nil {
				return s.file.Close()
			}
			return nil
		},
	})
	log.Info(context.Background(), "access log enabled", "file", s.filePath)
	return s, nil
}

func (s *Store) Enabled() bool {
	return s != nil && s.enabled
}

func (s *Store) MaxBodyBytes() int {
	if s == nil || s.maxBody <= 0 {
		return 8192
	}
	return s.maxBody
}

func (s *Store) SkipPaths() []string {
	if s == nil {
		return nil
	}
	return s.skipPaths
}

func (s *Store) Append(e Entry) error {
	if !s.Enabled() {
		return nil
	}
	s.mu.Lock()
	s.entries = append(s.entries, e)
	if len(s.entries) > s.maxMemory {
		s.entries = s.entries[len(s.entries)-s.maxMemory:]
	}
	s.mu.Unlock()

	line, err := json.Marshal(e)
	if err != nil {
		return err
	}
	if s.file != nil {
		_, err = s.file.Write(append(line, '\n'))
	}
	return err
}

func (s *Store) List(limit int) []Entry {
	if !s.Enabled() {
		return nil
	}
	if limit <= 0 {
		limit = 100
	}
	s.mu.RLock()
	n := len(s.entries)
	if n == 0 {
		s.mu.RUnlock()
		return nil
	}
	start := 0
	if n > limit {
		start = n - limit
	}
	out := append([]Entry(nil), s.entries[start:]...)
	s.mu.RUnlock()
	// 最新在前
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}
