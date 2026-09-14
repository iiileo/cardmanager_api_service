package service

import (
	"context"
	"strings"
	"time"

	"card_manager/api_service/internal/api/dto"
	domaindash "card_manager/api_service/internal/domain/dashboard"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
)

type DashboardService interface {
	HomeStats(ctx context.Context, userID, storeID int64, month string) (*dto.HomeStatsResponse, error)
}

type dashboardService struct {
	dash     domaindash.Repository
	storeSvc StoreService
	log      *logger.Logger
}

func NewDashboardService(
	dash domaindash.Repository,
	storeSvc StoreService,
	log *logger.Logger,
) DashboardService {
	return &dashboardService{dash: dash, storeSvc: storeSvc, log: log}
}

func (s *dashboardService) HomeStats(ctx context.Context, userID, storeID int64, month string) (*dto.HomeStatsResponse, error) {
	if _, err := s.storeSvc.RequireActiveMember(ctx, userID, storeID); err != nil {
		return nil, err
	}

	now := time.Now().In(time.Local)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	todayEnd := todayStart.Add(24*time.Hour - time.Nanosecond)

	monthStart, monthEnd, err := resolveMonthRange(strings.TrimSpace(month), now)
	if err != nil {
		return nil, err
	}

	stats, err := s.dash.HomeStats(ctx, storeID, todayStart, todayEnd, monthStart, monthEnd)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	return &dto.HomeStatsResponse{
		TodayRecharge:   stats.Today.Recharge,
		TodayConsume:    stats.Today.Consume,
		TodayNewMembers: stats.Today.NewMembers,
		TodayTxnCount:   stats.Today.TxnCount,
		MonthRecharge:   stats.Month.Recharge,
		MonthConsume:    stats.Month.Consume,
		MonthNewMembers: stats.Month.NewMembers,
		MonthTxnCount:   stats.Month.TxnCount,
		StoreBalance:    stats.StoreBalance,
	}, nil
}

// resolveMonthRange 解析按月查询范围。
// month 为空时：当月 1 日 0 点 ~ 今日结束（与流水页「本月」一致）。
// month=YYYY-MM 时：该自然月整月。
func resolveMonthRange(month string, now time.Time) (from, to time.Time, err error) {
	if month == "" {
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).
			Add(24*time.Hour - time.Nanosecond)
		return start, end, nil
	}

	t, parseErr := time.ParseInLocation("2006-01", month, time.Local)
	if parseErr != nil {
		return time.Time{}, time.Time{}, ierr.Validation("month 格式应为 YYYY-MM")
	}
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return start, end, nil
}
