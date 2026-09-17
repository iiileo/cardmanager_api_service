package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"card_manager/api_service/internal/api/dto"
	domaindash "card_manager/api_service/internal/domain/dashboard"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
)

type DashboardService interface {
	HomeStats(ctx context.Context, userID, storeID int64, month string) (*dto.HomeStatsResponse, error)
	StatsOverview(ctx context.Context, userID, storeID int64, days int) (*dto.StatsOverviewResponse, error)
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

func (s *dashboardService) StatsOverview(ctx context.Context, userID, storeID int64, days int) (*dto.StatsOverviewResponse, error) {
	if _, err := s.storeSvc.RequireActiveMember(ctx, userID, storeID); err != nil {
		return nil, err
	}
	if days != 7 && days != 30 {
		return nil, ierr.Validation("days 仅支持 7 或 30")
	}

	curFrom, curTo, prevFrom, prevTo := resolveOverviewRanges(days, time.Now().In(time.Local))
	raw, err := s.dash.StatsOverview(ctx, domaindash.StatsOverviewQuery{
		StoreID: storeID,
		CurFrom: curFrom, CurTo: curTo,
		PrevFrom: prevFrom, PrevTo: prevTo,
	})
	if err != nil {
		return nil, ierr.Internal(err)
	}

	activeRate := 0
	if raw.ActiveMembers > 0 {
		activeRate = raw.RepeatCustomers * 100 / raw.ActiveMembers
	}

	summary := dto.StatsOverviewSummary{
		RechargeAmount:  raw.Cur.RechargeAmount,
		ConsumeAmount:   raw.Cur.ConsumeAmount,
		NewOpens:        raw.Cur.NewOpens,
		NewOpensDelta:   raw.Cur.NewOpens - raw.Prev.NewOpens,
		RepeatCustomers: raw.RepeatCustomers,
		ActiveRatePct:   activeRate,
		RechargeChangePct: pctChange(raw.Cur.RechargeAmount, raw.Prev.RechargeAmount),
		ConsumeChangePct:  pctChange(raw.Cur.ConsumeAmount, raw.Prev.ConsumeAmount),
	}

	daily := make([]*dto.StatsDailyRechargePoint, 0, len(raw.DailyRecharge))
	for _, p := range raw.DailyRecharge {
		daily = append(daily, &dto.StatsDailyRechargePoint{
			Date:   p.Date.Format("2006-01-02"),
			Amount: p.Amount,
		})
	}

	mixItems := make([]*dto.StatsCardTypeMixItem, 0, len(raw.CardTypeMix))
	totalCards := 0
	for _, row := range raw.CardTypeMix {
		totalCards += row.Count
	}
	for _, row := range raw.CardTypeMix {
		pct := 0
		if totalCards > 0 {
			pct = row.Count * 100 / totalCards
		}
		mixItems = append(mixItems, &dto.StatsCardTypeMixItem{
			CardType: row.CardType,
			Count:    row.Count,
			Percent:  pct,
		})
	}

	rank := make([]*dto.StatsRechargeRankItem, 0, len(raw.RechargeRank))
	for i, row := range raw.RechargeRank {
		rank = append(rank, &dto.StatsRechargeRankItem{
			Rank:     i + 1,
			MemberID: fmt.Sprintf("%d", row.MemberID),
			Name:     row.Name,
			Amount:   row.Amount,
		})
	}

	return &dto.StatsOverviewResponse{
		RangeDays:     days,
		From:          curFrom.Format("2006-01-02"),
		To:            curTo.Format("2006-01-02"),
		Summary:       summary,
		DailyRecharge: daily,
		CardTypeMix: dto.StatsCardTypeMix{
			TotalHolders: raw.TotalCardHolders,
			Items:        mixItems,
		},
		RechargeRank: rank,
	}, nil
}

// resolveOverviewRanges 近 N 天（含今日）及上一段等长窗口。
func resolveOverviewRanges(days int, now time.Time) (curFrom, curTo, prevFrom, prevTo time.Time) {
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	curTo = todayStart.Add(24*time.Hour - time.Nanosecond)
	curFrom = todayStart.AddDate(0, 0, -(days - 1))
	prevTo = curFrom.Add(-time.Nanosecond)
	prevFrom = time.Date(prevTo.Year(), prevTo.Month(), prevTo.Day(), 0, 0, 0, 0, time.Local).
		AddDate(0, 0, -(days - 1))
	return curFrom, curTo, prevFrom, prevTo
}

func pctChange(cur, prev int) *int {
	if prev <= 0 {
		if cur <= 0 {
			return nil
		}
		v := 100
		return &v
	}
	v := (cur - prev) * 100 / prev
	return &v
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
