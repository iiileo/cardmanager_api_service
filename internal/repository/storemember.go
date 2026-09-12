package repository

import (
	"context"
	"time"

	"card_manager/api_service/ent"
	entmember "card_manager/api_service/ent/storemember"
	"card_manager/api_service/internal/domain/store"
	domainmember "card_manager/api_service/internal/domain/storemember"
	"card_manager/api_service/internal/postgres"
)

type storeMemberRepository struct {
	client *postgres.Client
}

func NewStoreMemberRepository(client *postgres.Client) domainmember.Repository {
	return &storeMemberRepository{client: client}
}

func (r *storeMemberRepository) Create(ctx context.Context, in domainmember.CreateInput) (*domainmember.Member, error) {
	b := r.client.Ent().StoreMember.Create().
		SetStoreID(in.StoreID).
		SetUserID(in.UserID).
		SetRole(in.Role).
		SetStatus(in.Status)
	if in.DisplayName != nil {
		b.SetDisplayName(*in.DisplayName)
	}
	if in.JoinedAt != nil {
		b.SetJoinedAt(*in.JoinedAt)
	}
	m, err := b.Save(ctx)
	if err != nil {
		return nil, err
	}
	return domainmember.FromEnt(m), nil
}

func (r *storeMemberRepository) GetByStoreUser(ctx context.Context, storeID, userID int64) (*domainmember.Member, error) {
	m, err := r.client.Ent().StoreMember.Query().
		Where(entmember.StoreIDEQ(storeID), entmember.UserIDEQ(userID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainmember.FromEnt(m), nil
}

func (r *storeMemberRepository) GetByID(ctx context.Context, id int64) (*domainmember.Member, error) {
	m, err := r.client.Ent().StoreMember.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainmember.FromEnt(m), nil
}

func (r *storeMemberRepository) ListByUser(ctx context.Context, userID int64) ([]*domainmember.Member, error) {
	list, err := r.client.Ent().StoreMember.Query().
		Where(
			entmember.UserIDEQ(userID),
			entmember.StatusIn(store.MemberActive, store.MemberPending, store.MemberRejected),
		).
		Order(ent.Desc(entmember.FieldUpdatedAt)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domainmember.Member, 0, len(list))
	for _, m := range list {
		out = append(out, domainmember.FromEnt(m))
	}
	return out, nil
}

func (r *storeMemberRepository) ListByStore(ctx context.Context, storeID int64, statuses []string) ([]*domainmember.Member, error) {
	q := r.client.Ent().StoreMember.Query().Where(entmember.StoreIDEQ(storeID))
	if len(statuses) > 0 {
		q = q.Where(entmember.StatusIn(statuses...))
	}
	list, err := q.Order(ent.Asc(entmember.FieldCreatedAt)).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domainmember.Member, 0, len(list))
	for _, m := range list {
		out = append(out, domainmember.FromEnt(m))
	}
	return out, nil
}

func (r *storeMemberRepository) CountActiveByStore(ctx context.Context, storeID int64) (int, error) {
	return r.client.Ent().StoreMember.Query().
		Where(entmember.StoreIDEQ(storeID), entmember.StatusEQ(store.MemberActive)).
		Count(ctx)
}

func (r *storeMemberRepository) UpdateStatus(ctx context.Context, id int64, status string, joinedAt *time.Time, displayName *string) (*domainmember.Member, error) {
	b := r.client.Ent().StoreMember.UpdateOneID(id).SetStatus(status)
	if joinedAt != nil {
		b.SetJoinedAt(*joinedAt)
	}
	if displayName != nil {
		b.SetDisplayName(*displayName)
	}
	m, err := b.Save(ctx)
	if err != nil {
		return nil, err
	}
	return domainmember.FromEnt(m), nil
}
