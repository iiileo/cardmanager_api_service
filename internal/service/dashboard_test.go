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
	todayFrom, todayTo time.Time
	monthFrom, monthTo time.Time
	stats              *domaindash.HomeStats
}

func (m *memDashRepo) HomeStats(_ context.Context, _ int64, todayFrom, todayTo, monthFrom, monthTo time.Time) (*domaindash.HomeStats, error) {
	m.todayFrom, m.todayTo = todayFrom, todayTo
	m.monthFrom, m.monthTo = monthFrom, monthTo
	return m.stats, nil
}

func (m *memDashRepo) StatsOverview(_ context.Context, _ domaindash.StatsOverviewQuery) (*domaindash.StatsOverview, error) {
	return &domaindash.StatsOverview{}, nil
}

func TestDashboardService_HomeStats(t *testing.T) {
	repo := &memDashRepo{stats: &domaindash.HomeStats{
		Today: domaindash.PeriodStats{
			Recharge: 50000, Consume: 12000, NewMembers: 2, TxnCount: 5,
		},
		Month: domaindash.PeriodStats{
			Recharge: 120000, Consume: 45000, NewMembers: 8, TxnCount: 22,
		},
		StoreBalance: 70000,
	}}
	svc := NewDashboardService(
		repo,
		okStoreSvc{},
		logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}}),
	)

	resp, err := svc.HomeStats(context.Background(), 1, 2, "")
	if err != nil {
		t.Fatal(err)
	}
	if resp.TodayRecharge != 50000 || resp.TodayConsume != 12000 ||
		resp.TodayNewMembers != 2 || resp.TodayTxnCount != 5 {
		t.Fatalf("unexpected today stats: %+v", resp)
	}
	if resp.MonthRecharge != 120000 || resp.MonthConsume != 45000 ||
		resp.MonthNewMembers != 8 || resp.MonthTxnCount != 22 {
		t.Fatalf("unexpected month stats: %+v", resp)
	}
	if resp.StoreBalance != 70000 {
		t.Fatalf("unexpected balance: %d", resp.StoreBalance)
	}

	now := time.Now().In(time.Local)
	wantMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	if !repo.monthFrom.Equal(wantMonthStart) {
		t.Fatalf("monthFrom=%v want=%v", repo.monthFrom, wantMonthStart)
	}
}

func TestDashboardService_HomeStats_SpecificMonth(t *testing.T) {
	repo := &memDashRepo{stats: &domaindash.HomeStats{}}
	svc := NewDashboardService(
		repo,
		okStoreSvc{},
		logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}}),
	)

	if _, err := svc.HomeStats(context.Background(), 1, 2, "2026-03"); err != nil {
		t.Fatal(err)
	}
	wantStart := time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local)
	wantEnd := time.Date(2026, 3, 31, 23, 59, 59, 999999999, time.Local)
	if !repo.monthFrom.Equal(wantStart) || !repo.monthTo.Equal(wantEnd) {
		t.Fatalf("month range=%v~%v want=%v~%v", repo.monthFrom, repo.monthTo, wantStart, wantEnd)
	}
}

func TestDashboardService_HomeStats_InvalidMonth(t *testing.T) {
	svc := NewDashboardService(
		&memDashRepo{stats: &domaindash.HomeStats{}},
		okStoreSvc{},
		logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}}),
	)
	if _, err := svc.HomeStats(context.Background(), 1, 2, "2026/03"); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestDashboardService_StatsOverview_Days(t *testing.T) {
	svc := NewDashboardService(
		&memDashRepo{stats: &domaindash.HomeStats{}},
		okStoreSvc{},
		logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}}),
	)
	if _, err := svc.StatsOverview(context.Background(), 1, 2, 7); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StatsOverview(context.Background(), 1, 2, 14); err == nil {
		t.Fatal("expected validation error for days=14")
	}
}

func TestResolveMonthRange(t *testing.T) {
	now := time.Date(2026, 9, 14, 15, 0, 0, 0, time.Local)

	from, to, err := resolveMonthRange("", now)
	if err != nil {
		t.Fatal(err)
	}
	if from.Day() != 1 || to.Day() != 14 {
		t.Fatalf("current month range=%v~%v", from, to)
	}

	from, to, err = resolveMonthRange("2025-08", now)
	if err != nil {
		t.Fatal(err)
	}
	if from != time.Date(2025, 8, 1, 0, 0, 0, 0, time.Local) {
		t.Fatalf("from=%v", from)
	}
	if to.Day() != 31 || to.Month() != 8 {
		t.Fatalf("to=%v", to)
	}
}
