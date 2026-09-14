package service

import (
	"context"
	"strings"

	"card_manager/api_service/internal/api/dto"
	domainnotify "card_manager/api_service/internal/domain/notifysetting"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
)

type NotifySettingService interface {
	List(ctx context.Context, userID, storeID int64) (*dto.NotifySettingListResponse, error)
	Get(ctx context.Context, userID, storeID int64, event string) (*dto.NotifySettingResponse, error)
	Update(ctx context.Context, userID, storeID int64, event string, req dto.UpdateNotifySettingRequest) (*dto.NotifySettingResponse, error)
}

type notifySettingService struct {
	settings domainnotify.Repository
	storeSvc StoreService
	log      *logger.Logger
}

func NewNotifySettingService(
	settings domainnotify.Repository,
	storeSvc StoreService,
	log *logger.Logger,
) NotifySettingService {
	return &notifySettingService{settings: settings, storeSvc: storeSvc, log: log}
}

func (s *notifySettingService) List(ctx context.Context, userID, storeID int64) (*dto.NotifySettingListResponse, error) {
	if _, err := s.storeSvc.RequireOwner(ctx, userID, storeID); err != nil {
		return nil, err
	}
	if err := s.settings.EnsureDefaults(ctx, storeID); err != nil {
		return nil, ierr.Internal(err)
	}
	list, err := s.settings.ListByStore(ctx, storeID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	out := make([]*dto.NotifySettingResponse, 0, len(list))
	for _, item := range list {
		out = append(out, toNotifySettingDTO(item))
	}
	return &dto.NotifySettingListResponse{List: out}, nil
}

func (s *notifySettingService) Get(ctx context.Context, userID, storeID int64, event string) (*dto.NotifySettingResponse, error) {
	event = strings.TrimSpace(event)
	if !domainnotify.ValidEvent(event) {
		return nil, ierr.Validation("通知类型不正确")
	}
	if _, err := s.storeSvc.RequireOwner(ctx, userID, storeID); err != nil {
		return nil, err
	}
	if err := s.settings.EnsureDefaults(ctx, storeID); err != nil {
		return nil, ierr.Internal(err)
	}
	item, err := s.settings.GetByStoreEvent(ctx, storeID, event)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	if item == nil {
		return nil, ierr.NotFound("通知配置不存在")
	}
	return toNotifySettingDTO(item), nil
}

func (s *notifySettingService) Update(ctx context.Context, userID, storeID int64, event string, req dto.UpdateNotifySettingRequest) (*dto.NotifySettingResponse, error) {
	event = strings.TrimSpace(event)
	if !domainnotify.ValidEvent(event) {
		return nil, ierr.Validation("通知类型不正确")
	}
	if !req.Enabled {
		// 关闭通知时不校验渠道
	} else if !req.Wechat && !req.App {
		return nil, ierr.Validation("请至少选择一种通知方式")
	}
	if _, err := s.storeSvc.RequireOwner(ctx, userID, storeID); err != nil {
		return nil, err
	}
	item, err := s.settings.Upsert(ctx, storeID, event, domainnotify.UpsertInput{
		Enabled:    req.Enabled,
		NotifyBoss: req.NotifyBoss,
		Wechat:     req.Wechat,
		App:        req.App,
	})
	if err != nil {
		return nil, ierr.Internal(err)
	}
	s.log.Info(ctx, "notify setting updated", "store_id", storeID, "event", event)
	return toNotifySettingDTO(item), nil
}

func toNotifySettingDTO(s *domainnotify.Setting) *dto.NotifySettingResponse {
	return &dto.NotifySettingResponse{
		Event:      s.Event,
		Enabled:    s.Enabled,
		NotifyBoss: s.NotifyBoss,
		Channels: &dto.NotifyChannelsResponse{
			Wechat: s.Wechat,
			App:    s.App,
		},
	}
}
