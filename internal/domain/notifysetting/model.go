package notifysetting

import (
	"context"
	"time"

	"card_manager/api_service/ent"
)

const (
	EventOpen     = "open"
	EventRecharge = "recharge"
	EventConsume  = "consume"
	EventCount    = "count"
)

var AllEvents = []string{EventOpen, EventRecharge, EventConsume, EventCount}

type Setting struct {
	ID         int64
	StoreID    int64
	Event      string
	Enabled    bool
	NotifyBoss bool
	Wechat     bool
	App        bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func FromEnt(s *ent.StoreNotifySetting) *Setting {
	if s == nil {
		return nil
	}
	return &Setting{
		ID:         s.ID,
		StoreID:    s.StoreID,
		Event:      s.Event,
		Enabled:    s.Enabled,
		NotifyBoss: s.NotifyBoss,
		Wechat:     s.Wechat,
		App:        s.App,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

type UpsertInput struct {
	Enabled    bool
	NotifyBoss bool
	Wechat     bool
	App        bool
}

type Repository interface {
	ListByStore(ctx context.Context, storeID int64) ([]*Setting, error)
	GetByStoreEvent(ctx context.Context, storeID int64, event string) (*Setting, error)
	EnsureDefaults(ctx context.Context, storeID int64) error
	Upsert(ctx context.Context, storeID int64, event string, in UpsertInput) (*Setting, error)
}

func ValidEvent(event string) bool {
	switch event {
	case EventOpen, EventRecharge, EventConsume, EventCount:
		return true
	default:
		return false
	}
}
