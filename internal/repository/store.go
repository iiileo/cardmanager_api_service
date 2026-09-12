package repository

import (
	"context"
	"time"

	"card_manager/api_service/ent"
	entstore "card_manager/api_service/ent/store"
	domainstore "card_manager/api_service/internal/domain/store"
	"card_manager/api_service/internal/postgres"
)

type storeRepository struct {
	client *postgres.Client
}

func NewStoreRepository(client *postgres.Client) domainstore.Repository {
	return &storeRepository{client: client}
}

func (r *storeRepository) Create(ctx context.Context, in domainstore.CreateInput) (*domainstore.Store, error) {
	s, err := buildStoreCreate(r.client.Ent().Store.Create(), in).Save(ctx)
	if err != nil {
		return nil, err
	}
	return domainstore.FromEnt(s), nil
}

func (r *storeRepository) CreateWithOwner(ctx context.Context, in domainstore.CreateInput) (*domainstore.Store, error) {
	var out *domainstore.Store
	err := r.client.WithTx(ctx, func(tx *ent.Tx) error {
		s, err := buildStoreCreate(tx.Store.Create(), in).Save(ctx)
		if err != nil {
			return err
		}
		now := time.Now().UTC()
		if _, err := tx.StoreMember.Create().
			SetStoreID(s.ID).
			SetUserID(in.OwnerUserID).
			SetRole(domainstore.RoleOwner).
			SetStatus(domainstore.MemberActive).
			SetJoinedAt(now).
			Save(ctx); err != nil {
			return err
		}
		out = domainstore.FromEnt(s)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func buildStoreCreate(b *ent.StoreCreate, in domainstore.CreateInput) *ent.StoreCreate {
	b = b.
		SetName(in.Name).
		SetCity(in.City).
		SetOpenTime(in.OpenTime).
		SetCloseTime(in.CloseTime).
		SetInviteCode(in.InviteCode).
		SetOwnerUserID(in.OwnerUserID).
		SetStatus(domainstore.StoreStatusNormal)
	if in.Address != nil {
		b.SetAddress(*in.Address)
	}
	if in.BizType != nil {
		b.SetBizType(*in.BizType)
	}
	return b
}

func (r *storeRepository) GetByID(ctx context.Context, id int64) (*domainstore.Store, error) {
	s, err := r.client.Ent().Store.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainstore.FromEnt(s), nil
}

func (r *storeRepository) GetByInviteCode(ctx context.Context, code string) (*domainstore.Store, error) {
	s, err := r.client.Ent().Store.Query().Where(entstore.InviteCodeEQ(code)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainstore.FromEnt(s), nil
}

func (r *storeRepository) Update(ctx context.Context, id int64, in domainstore.UpdateInput) (*domainstore.Store, error) {
	b := r.client.Ent().Store.UpdateOneID(id)
	if in.Name != nil {
		b.SetName(*in.Name)
	}
	if in.City != nil {
		b.SetCity(*in.City)
	}
	if in.Address != nil {
		b.SetAddress(*in.Address)
	}
	if in.OpenTime != nil {
		b.SetOpenTime(*in.OpenTime)
	}
	if in.CloseTime != nil {
		b.SetCloseTime(*in.CloseTime)
	}
	if in.BizType != nil {
		b.SetBizType(*in.BizType)
	}
	s, err := b.Save(ctx)
	if err != nil {
		return nil, err
	}
	return domainstore.FromEnt(s), nil
}

func (r *storeRepository) UpdateInviteCode(ctx context.Context, id int64, code string) (*domainstore.Store, error) {
	s, err := r.client.Ent().Store.UpdateOneID(id).SetInviteCode(code).Save(ctx)
	if err != nil {
		return nil, err
	}
	return domainstore.FromEnt(s), nil
}

func (r *storeRepository) InviteCodeExists(ctx context.Context, code string) (bool, error) {
	return r.client.Ent().Store.Query().Where(entstore.InviteCodeEQ(code)).Exist(ctx)
}
