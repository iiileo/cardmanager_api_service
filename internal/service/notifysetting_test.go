package service

import (
	"context"
	"fmt"
	"testing"

	"card_manager/api_service/internal/api/dto"
	"card_manager/api_service/internal/config"
	domainnotify "card_manager/api_service/internal/domain/notifysetting"
	"card_manager/api_service/internal/logger"
)

type memNotifyRepo struct {
	byKey map[string]*domainnotify.Setting
	seq   int64
}

func notifyKey(storeID int64, event string) string {
	return fmt.Sprintf("%d:%s", storeID, event)
}

func (m *memNotifyRepo) ListByStore(_ context.Context, storeID int64) ([]*domainnotify.Setting, error) {
	var out []*domainnotify.Setting
	for _, s := range m.byKey {
		if s.StoreID == storeID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (m *memNotifyRepo) GetByStoreEvent(_ context.Context, storeID int64, event string) (*domainnotify.Setting, error) {
	return m.byKey[notifyKey(storeID, event)], nil
}

func (m *memNotifyRepo) EnsureDefaults(_ context.Context, storeID int64) error {
	for _, event := range domainnotify.AllEvents {
		key := notifyKey(storeID, event)
		if m.byKey[key] != nil {
			continue
		}
		m.seq++
		m.byKey[key] = &domainnotify.Setting{
			ID: m.seq, StoreID: storeID, Event: event,
			Enabled: true, NotifyBoss: true, Wechat: true, App: false,
		}
	}
	return nil
}

func (m *memNotifyRepo) Upsert(_ context.Context, storeID int64, event string, in domainnotify.UpsertInput) (*domainnotify.Setting, error) {
	key := notifyKey(storeID, event)
	if s := m.byKey[key]; s != nil {
		s.Enabled = in.Enabled
		s.NotifyBoss = in.NotifyBoss
		s.Wechat = in.Wechat
		s.App = in.App
		return s, nil
	}
	m.seq++
	s := &domainnotify.Setting{
		ID: m.seq, StoreID: storeID, Event: event,
		Enabled: in.Enabled, NotifyBoss: in.NotifyBoss, Wechat: in.Wechat, App: in.App,
	}
	m.byKey[key] = s
	return s, nil
}

type ownerStoreSvc struct{ okStoreSvc }

func TestNotifySettingService_ListAndUpdate(t *testing.T) {
	log := logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}})
	repo := &memNotifyRepo{byKey: map[string]*domainnotify.Setting{}}
	svc := NewNotifySettingService(repo, ownerStoreSvc{}, log)

	list, err := svc.List(context.Background(), 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.List) != 4 {
		t.Fatalf("want 4 defaults, got %d", len(list.List))
	}

	updated, err := svc.Update(context.Background(), 1, 2, domainnotify.EventOpen, dto.UpdateNotifySettingRequest{
		Enabled: true, NotifyBoss: false, Wechat: true, App: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Channels.App || updated.NotifyBoss {
		t.Fatalf("unexpected update: %+v", updated)
	}
}

func TestNotifySettingService_UpdateValidation(t *testing.T) {
	svc := NewNotifySettingService(&memNotifyRepo{byKey: map[string]*domainnotify.Setting{}}, ownerStoreSvc{}, logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}}))
	if _, err := svc.Update(context.Background(), 1, 2, "bad", dto.UpdateNotifySettingRequest{}); err == nil {
		t.Fatal("expected invalid event")
	}
	if _, err := svc.Update(context.Background(), 1, 2, domainnotify.EventOpen, dto.UpdateNotifySettingRequest{
		Enabled: true, Wechat: false, App: false,
	}); err == nil {
		t.Fatal("expected channel validation")
	}
}
