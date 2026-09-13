package service

import (
	"context"
	"time"

	"card_manager/api_service/internal/api/dto"
	domaindash "card_manager/api_service/internal/domain/dashboard"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
)

type DashboardService interface {
	HomeStats(ctx context.Context, userID, storeID int64) (*dto.HomeStatsResponse, error)
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

func (s *dashboardService) HomeStats(ctx context.Context, userID, storeID int64) (*dto.HomeStatsResponse, error) {
	if _, err := s.storeSvc.RequireActiveMember(ctx, userID, storeID); err != nil {
		return nil, err
	}

	now := time.Now().In(time.Local)
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	from := day
	to := day.Add(24*time.Hour - time.Nanosecond)

	stats, err := s.dash.HomeStats(ctx, storeID, from, to)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	return &dto.HomeStatsResponse{
		TodayRecharge:   stats.TodayRecharge,
		TodayConsume:    stats.TodayConsume,
		TodayNewMembers: stats.TodayNewMembers,
		StoreBalance:    stats.StoreBalance,
		TodayTxnCount:   stats.TodayTxnCount,
	}, nil
}
