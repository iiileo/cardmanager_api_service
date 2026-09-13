package repository

import (
	"context"
	"strings"

	"card_manager/api_service/ent"
	entmember "card_manager/api_service/ent/member"
	domainmember "card_manager/api_service/internal/domain/member"
	"card_manager/api_service/internal/postgres"
)

type memberRepository struct {
	client *postgres.Client
}

func NewMemberRepository(client *postgres.Client) domainmember.Repository {
	return &memberRepository{client: client}
}

func (r *memberRepository) GetByID(ctx context.Context, id int64) (*domainmember.Member, error) {
	m, err := r.client.Ent().Member.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainmember.FromEnt(m), nil
}

func (r *memberRepository) GetByStorePhone(ctx context.Context, storeID int64, phone string) (*domainmember.Member, error) {
	m, err := r.client.Ent().Member.Query().
		Where(entmember.StoreIDEQ(storeID), entmember.PhoneEQ(phone)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainmember.FromEnt(m), nil
}

func (r *memberRepository) Create(ctx context.Context, in domainmember.CreateInput) (*domainmember.Member, error) {
	b := r.client.Ent().Member.Create().
		SetStoreID(in.StoreID).
		SetName(in.Name).
		SetPhone(in.Phone)
	if in.Source != nil {
		b.SetSource(*in.Source)
	}
	m, err := b.Save(ctx)
	if err != nil {
		return nil, err
	}
	return domainmember.FromEnt(m), nil
}

func (r *memberRepository) ListByStore(ctx context.Context, storeID int64, q string, limit, offset int) ([]*domainmember.Member, int, error) {
	query := r.client.Ent().Member.Query().Where(entmember.StoreIDEQ(storeID))
	q = strings.TrimSpace(q)
	if q != "" {
		query = query.Where(entmember.Or(
			entmember.NameContainsFold(q),
			entmember.PhoneContains(q),
		))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	list, err := query.
		Order(ent.Desc(entmember.FieldID)).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*domainmember.Member, 0, len(list))
	for _, m := range list {
		out = append(out, domainmember.FromEnt(m))
	}
	return out, total, nil
}
