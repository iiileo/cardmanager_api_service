package repository

import (
	"context"

	"card_manager/api_service/ent"
	entledger "card_manager/api_service/ent/ledgerentry"
	domainledger "card_manager/api_service/internal/domain/ledger"
	"card_manager/api_service/internal/postgres"
)

type ledgerRepository struct {
	client *postgres.Client
}

func NewLedgerRepository(client *postgres.Client) domainledger.Repository {
	return &ledgerRepository{client: client}
}

func (r *ledgerRepository) Create(ctx context.Context, in domainledger.CreateInput) (*domainledger.Entry, error) {
	var out *domainledger.Entry
	err := r.client.WithTx(ctx, func(tx *ent.Tx) error {
		b := tx.LedgerEntry.Create().
			SetStoreID(in.StoreID).
			SetMemberID(in.MemberID).
			SetCardID(in.CardID).
			SetType(in.Type).
			SetOperatorID(in.OperatorID)
		if in.Amount != nil {
			b.SetAmount(*in.Amount)
		}
		if in.Times != nil {
			b.SetTimes(*in.Times)
		}
		if in.ItemName != nil {
			b.SetItemName(*in.ItemName)
		}
		if in.BalanceAfter != nil {
			b.SetBalanceAfter(*in.BalanceAfter)
		}
		if in.TimesAfter != nil {
			b.SetTimesAfter(*in.TimesAfter)
		}
		if in.Remark != nil {
			b.SetRemark(*in.Remark)
		}
		e, err := b.Save(ctx)
		if err != nil {
			return err
		}
		created := make([]*ent.LedgerEntryItem, 0, len(in.Items))
		for _, it := range in.Items {
			row, err := tx.LedgerEntryItem.Create().
				SetLedgerID(e.ID).
				SetItemBalanceID(it.ItemBalanceID).
				SetProductItemID(it.ProductItemID).
				SetNameSnapshot(it.NameSnapshot).
				SetTimes(it.Times).
				SetTimesAfter(it.TimesAfter).
				Save(ctx)
			if err != nil {
				return err
			}
			created = append(created, row)
		}
		out = domainledger.FromEnt(e, created)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *ledgerRepository) GetByID(ctx context.Context, id int64) (*domainledger.Entry, error) {
	e, err := r.client.Ent().LedgerEntry.Query().
		Where(entledger.IDEQ(id)).
		WithItems().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainledger.FromEnt(e, nil), nil
}

func (r *ledgerRepository) List(ctx context.Context, f domainledger.ListFilter) ([]*domainledger.Entry, int, error) {
	q := r.client.Ent().LedgerEntry.Query().Where(entledger.StoreIDEQ(f.StoreID))
	if f.Type != "" {
		q = q.Where(entledger.TypeEQ(f.Type))
	}
	if f.From != nil {
		q = q.Where(entledger.CreatedAtGTE(*f.From))
	}
	if f.To != nil {
		q = q.Where(entledger.CreatedAtLTE(*f.To))
	}
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	list, err := q.WithItems().
		Order(ent.Desc(entledger.FieldCreatedAt), ent.Desc(entledger.FieldID)).
		Limit(limit).
		Offset(f.Offset).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}
	out := make([]*domainledger.Entry, 0, len(list))
	for _, e := range list {
		out = append(out, domainledger.FromEnt(e, nil))
	}
	return out, total, nil
}
