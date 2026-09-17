package repository

import (
	"context"
	"strings"

	"card_manager/api_service/ent"
	entmember "card_manager/api_service/ent/member"
	"card_manager/api_service/ent/predicate"
	domainmember "card_manager/api_service/internal/domain/member"
	"card_manager/api_service/internal/pkg/pinyinutil"
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

func (r *memberRepository) MapByIDs(ctx context.Context, ids []int64) (map[int64]*domainmember.Member, error) {
	out := make(map[int64]*domainmember.Member, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.client.Ent().Member.Query().Where(entmember.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, m := range rows {
		out[m.ID] = domainmember.FromEnt(m)
	}
	return out, nil
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
	full, initials := pinyinutil.NameKeys(in.Name)
	b := r.client.Ent().Member.Create().
		SetStoreID(in.StoreID).
		SetName(in.Name).
		SetNamePinyin(full).
		SetNameInitials(initials).
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
	raw := strings.TrimSpace(q)
	if raw != "" {
		key := pinyinutil.NormalizeQuery(raw)
		preds := []predicate.Member{
			entmember.NameContainsFold(raw),
			entmember.PhoneContains(raw),
		}
		if key != "" {
			preds = append(preds,
				entmember.NamePinyinContainsFold(key),
				entmember.NameInitialsContainsFold(key),
			)
		}
		query = query.Where(entmember.Or(preds...))
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

// BackfillNamePinyin 为缺失拼音字段的会员补写全拼/首拼。
func (r *memberRepository) BackfillNamePinyin(ctx context.Context) (int, error) {
	list, err := r.client.Ent().Member.Query().
		Where(entmember.Or(
			entmember.NamePinyinEQ(""),
			entmember.NameInitialsEQ(""),
		)).
		All(ctx)
	if err != nil {
		return 0, err
	}
	n := 0
	for _, m := range list {
		full, initials := pinyinutil.NameKeys(m.Name)
		if full == m.NamePinyin && initials == m.NameInitials {
			continue
		}
		if _, err := r.client.Ent().Member.UpdateOneID(m.ID).
			SetNamePinyin(full).
			SetNameInitials(initials).
			Save(ctx); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
