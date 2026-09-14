package repository

import (
	"context"

	"card_manager/api_service/ent"
	entnotify "card_manager/api_service/ent/storenotifysetting"
	domainnotify "card_manager/api_service/internal/domain/notifysetting"
	"card_manager/api_service/internal/postgres"
)

type notifySettingRepository struct {
	client *postgres.Client
}

func NewNotifySettingRepository(client *postgres.Client) domainnotify.Repository {
	return &notifySettingRepository{client: client}
}

func (r *notifySettingRepository) ListByStore(ctx context.Context, storeID int64) ([]*domainnotify.Setting, error) {
	list, err := r.client.Ent().StoreNotifySetting.Query().
		Where(entnotify.StoreIDEQ(storeID)).
		Order(ent.Asc(entnotify.FieldEvent)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domainnotify.Setting, 0, len(list))
	for _, s := range list {
		out = append(out, domainnotify.FromEnt(s))
	}
	return out, nil
}

func (r *notifySettingRepository) GetByStoreEvent(ctx context.Context, storeID int64, event string) (*domainnotify.Setting, error) {
	s, err := r.client.Ent().StoreNotifySetting.Query().
		Where(entnotify.StoreIDEQ(storeID), entnotify.EventEQ(event)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainnotify.FromEnt(s), nil
}

func (r *notifySettingRepository) EnsureDefaults(ctx context.Context, storeID int64) error {
	for _, event := range domainnotify.AllEvents {
		exists, err := r.client.Ent().StoreNotifySetting.Query().
			Where(entnotify.StoreIDEQ(storeID), entnotify.EventEQ(event)).
			Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if _, err := r.client.Ent().StoreNotifySetting.Create().
			SetStoreID(storeID).
			SetEvent(event).
			SetEnabled(true).
			SetNotifyBoss(true).
			SetWechat(true).
			SetApp(false).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *notifySettingRepository) Upsert(ctx context.Context, storeID int64, event string, in domainnotify.UpsertInput) (*domainnotify.Setting, error) {
	existing, err := r.GetByStoreEvent(ctx, storeID, event)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		s, err := r.client.Ent().StoreNotifySetting.Create().
			SetStoreID(storeID).
			SetEvent(event).
			SetEnabled(in.Enabled).
			SetNotifyBoss(in.NotifyBoss).
			SetWechat(in.Wechat).
			SetApp(in.App).
			Save(ctx)
		if err != nil {
			return nil, err
		}
		return domainnotify.FromEnt(s), nil
	}
	s, err := r.client.Ent().StoreNotifySetting.UpdateOneID(existing.ID).
		SetEnabled(in.Enabled).
		SetNotifyBoss(in.NotifyBoss).
		SetWechat(in.Wechat).
		SetApp(in.App).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return domainnotify.FromEnt(s), nil
}
