package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"card_manager/api_service/internal/config"
)

type Logger struct {
	base *slog.Logger
}

func NewLogger(cfg *config.Config) *Logger {
	level := slog.LevelInfo
	switch strings.ToLower(cfg.Logging.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	return &Logger{base: slog.New(handler)}
}

func (l *Logger) Info(ctx context.Context, msg string, keysAndValues ...any) {
	l.base.InfoContext(ctx, msg, keysAndValues...)
}

func (l *Logger) Warn(ctx context.Context, msg string, keysAndValues ...any) {
	l.base.WarnContext(ctx, msg, keysAndValues...)
}

func (l *Logger) Error(ctx context.Context, msg string, keysAndValues ...any) {
	l.base.ErrorContext(ctx, msg, keysAndValues...)
}

func (l *Logger) Debug(ctx context.Context, msg string, keysAndValues ...any) {
	l.base.DebugContext(ctx, msg, keysAndValues...)
}
