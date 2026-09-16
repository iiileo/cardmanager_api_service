package repository

import (
	"context"

	"card_manager/api_service/ent"
	entbalance "card_manager/api_service/ent/carditembalance"
	entcard "card_manager/api_service/ent/membercard"
	domaincard "card_manager/api_service/internal/domain/membercard"
	"card_manager/api_service/internal/postgres"
)

type memberCardRepository struct {
	client *postgres.Client
}

func NewMemberCardRepository(client *postgres.Client) domaincard.Repository {
	return &memberCardRepository{client: client}
}

func (r *memberCardRepository) withItems(q *ent.MemberCardQuery) *ent.MemberCardQuery {
	return q.WithItemBalances(func(iq *ent.CardItemBalanceQuery) {
		iq.Order(ent.Asc(entbalance.FieldID))
	})
}

func (r *memberCardRepository) ListByIDs(ctx context.Context, ids []int64) (map[int64]*domaincard.Card, error) {
	out := make(map[int64]*domaincard.Card)
	if len(ids) == 0 {
		return out, nil
	}
	list, err := r.client.Ent().MemberCard.Query().Where(entcard.IDIn(ids...)).All(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range list {
		out[c.ID] = domaincard.FromEnt(c, nil)
	}
	return out, nil
}

func (r *memberCardRepository) GetByID(ctx context.Context, id int64) (*domaincard.Card, error) {
	c, err := r.withItems(r.client.Ent().MemberCard.Query().Where(entcard.IDEQ(id))).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domaincard.FromEnt(c, nil), nil
}

func (r *memberCardRepository) GetByMemberAndType(ctx context.Context, storeID, memberID int64, typ string) (*domaincard.Card, error) {
	c, err := r.withItems(r.client.Ent().MemberCard.Query().Where(
		entcard.StoreIDEQ(storeID),
		entcard.MemberIDEQ(memberID),
		entcard.TypeEQ(typ),
	)).Order(ent.Desc(entcard.FieldID)).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domaincard.FromEnt(c, nil), nil
}

func (r *memberCardRepository) ListByMember(ctx context.Context, storeID, memberID int64) ([]*domaincard.Card, error) {
	list, err := r.withItems(r.client.Ent().MemberCard.Query().Where(
		entcard.StoreIDEQ(storeID),
		entcard.MemberIDEQ(memberID),
	)).Order(ent.Desc(entcard.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domaincard.Card, 0, len(list))
	for _, c := range list {
		out = append(out, domaincard.FromEnt(c, nil))
	}
	return out, nil
}
