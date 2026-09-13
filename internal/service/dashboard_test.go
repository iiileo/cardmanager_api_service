package service

import (
	"context"
	"testing"
	"time"

	"card_manager/api_service/internal/config"
	domaindash "card_manager/api_service/internal/domain/dashboard"
	"card_manager/api_service/internal/logger"
)

type memDashRepo struct {
	stats *domaindash.HomeStats
}

func (m *memDashRepo) HomeStats(_ context.Context, _ int64, _, _ time.Time) (*domaindash.HomeStats, error) {
	return m.stats, nil
}

func TestDashboardService_HomeStats(t *testing.T) {
	svc := NewDashboardService(
		&memDashRepo{stats: &domaindash.HomeStats{
			TodayRecharge:   50000,
			TodayConsume:    12000,
			TodayNewMembers: 2,
			StoreBalance:    70000,
			TodayTxnCount:   5,
		}},
		okStoreSvc{},
		logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}}),
	)
	resp, err := svc.HomeStats(context.Background(), 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if resp.TodayRecharge != 50000 || resp.TodayConsume != 12000 ||
		resp.TodayNewMembers != 2 || resp.StoreBalance != 70000 || resp.TodayTxnCount != 5 {
		t.Fatalf("unexpected stats: %+v", resp)
	}
}
