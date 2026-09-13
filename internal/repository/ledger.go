package repository

import (
	"context"
	"fmt"
	"strings"

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
			SetCardType(in.CardType).
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

func (r *ledgerRepository) applyListFilter(q *ent.LedgerEntryQuery, f domainledger.ListFilter) *ent.LedgerEntryQuery {
	q = q.Where(entledger.StoreIDEQ(f.StoreID))
	types := f.Types
	if len(types) == 0 && f.Type != "" {
		types = []string{f.Type}
	}
	if len(types) == 1 {
		q = q.Where(entledger.TypeEQ(types[0]))
	} else if len(types) > 1 {
		q = q.Where(entledger.TypeIn(types...))
	}
	if f.CardType != "" {
		q = q.Where(entledger.CardTypeEQ(f.CardType))
	}
	if f.MemberID != nil {
		q = q.Where(entledger.MemberIDEQ(*f.MemberID))
	}
	if f.CardID != nil {
		q = q.Where(entledger.CardIDEQ(*f.CardID))
	}
	if f.From != nil {
		q = q.Where(entledger.CreatedAtGTE(*f.From))
	}
	if f.To != nil {
		q = q.Where(entledger.CreatedAtLTE(*f.To))
	}
	return q
}

func (r *ledgerRepository) List(ctx context.Context, f domainledger.ListFilter) ([]*domainledger.Entry, int, error) {
	q := r.applyListFilter(r.client.Ent().LedgerEntry.Query(), f)
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

func (r *ledgerRepository) Summarize(ctx context.Context, f domainledger.SummaryFilter) (*domainledger.Summary, error) {
	where, args := buildSummaryWhere(f)

	typeSQL := `
SELECT type, card_type,
       COUNT(*)::int,
       COALESCE(SUM(CASE WHEN amount IS NULL THEN 0 ELSE ABS(amount) END), 0)::int,
       COALESCE(SUM(CASE WHEN times IS NULL THEN 0 ELSE ABS(times) END), 0)::int
FROM ledger_entries
WHERE ` + where + `
GROUP BY type, card_type
ORDER BY type, card_type`

	rows, err := r.client.DB().QueryContext(ctx, typeSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("summarize by type: %w", err)
	}
	defer rows.Close()

	sum := &domainledger.Summary{ByType: make([]domainledger.TypeBucket, 0)}
	for rows.Next() {
		var b domainledger.TypeBucket
		if err := rows.Scan(&b.Type, &b.CardType, &b.Count, &b.Amount, &b.Times); err != nil {
			return nil, err
		}
		sum.ByType = append(sum.ByType, b)
		sum.TotalCount += b.Count
		sum.TotalAmount += b.Amount
		sum.TotalTimes += b.Times
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// 套餐项目明细统计（仅当筛选包含 consume_pack 或未限定类型时）
	wantPack := len(f.Types) == 0
	for _, t := range f.Types {
		if t == domainledger.TypeConsumePack {
			wantPack = true
			break
		}
	}
	if wantPack {
		packWhere, packArgs := buildSummaryWherePrefixed(f, "e")
		packSQL := `
SELECT i.product_item_id, i.name_snapshot,
       COALESCE(SUM(i.times), 0)::int,
       COUNT(*)::int
FROM ledger_entry_items i
INNER JOIN ledger_entries e ON e.id = i.ledger_id
WHERE e.type = $` + fmt.Sprintf("%d", len(packArgs)+1) + ` AND ` + packWhere + `
GROUP BY i.product_item_id, i.name_snapshot
ORDER BY COALESCE(SUM(i.times), 0) DESC, i.product_item_id`

		packArgs = append(packArgs, domainledger.TypeConsumePack)
		prows, err := r.client.DB().QueryContext(ctx, packSQL, packArgs...)
		if err != nil {
			return nil, fmt.Errorf("summarize pack items: %w", err)
		}
		defer prows.Close()
		sum.ByPackItem = make([]domainledger.PackItemBucket, 0)
		for prows.Next() {
			var b domainledger.PackItemBucket
			if err := prows.Scan(&b.ProductItemID, &b.Name, &b.Times, &b.Count); err != nil {
				return nil, err
			}
			sum.ByPackItem = append(sum.ByPackItem, b)
		}
		if err := prows.Err(); err != nil {
			return nil, err
		}
	}
	return sum, nil
}

func buildSummaryWhere(f domainledger.SummaryFilter) (string, []any) {
	parts := []string{"store_id = $1"}
	args := []any{f.StoreID}
	n := 2
	if len(f.Types) == 1 {
		parts = append(parts, fmt.Sprintf("type = $%d", n))
		args = append(args, f.Types[0])
		n++
	} else if len(f.Types) > 1 {
		ph := make([]string, 0, len(f.Types))
		for _, t := range f.Types {
			ph = append(ph, fmt.Sprintf("$%d", n))
			args = append(args, t)
			n++
		}
		parts = append(parts, "type IN ("+strings.Join(ph, ",")+")" )
	}
	if f.CardType != "" {
		parts = append(parts, fmt.Sprintf("card_type = $%d", n))
		args = append(args, f.CardType)
		n++
	}
	if f.MemberID != nil {
		parts = append(parts, fmt.Sprintf("member_id = $%d", n))
		args = append(args, *f.MemberID)
		n++
	}
	if f.CardID != nil {
		parts = append(parts, fmt.Sprintf("card_id = $%d", n))
		args = append(args, *f.CardID)
		n++
	}
	if f.From != nil {
		parts = append(parts, fmt.Sprintf("created_at >= $%d", n))
		args = append(args, *f.From)
		n++
	}
	if f.To != nil {
		parts = append(parts, fmt.Sprintf("created_at <= $%d", n))
		args = append(args, *f.To)
		n++
	}
	_ = n
	return strings.Join(parts, " AND "), args
}

func buildSummaryWherePrefixed(f domainledger.SummaryFilter, prefix string) (string, []any) {
	p := prefix + "."
	parts := []string{p + "store_id = $1"}
	args := []any{f.StoreID}
	n := 2
	if len(f.Types) == 1 {
		parts = append(parts, fmt.Sprintf("%stype = $%d", p, n))
		args = append(args, f.Types[0])
		n++
	} else if len(f.Types) > 1 {
		ph := make([]string, 0, len(f.Types))
		for _, t := range f.Types {
			ph = append(ph, fmt.Sprintf("$%d", n))
			args = append(args, t)
			n++
		}
		parts = append(parts, p+"type IN ("+strings.Join(ph, ",")+")")
	}
	if f.CardType != "" {
		parts = append(parts, fmt.Sprintf("%scard_type = $%d", p, n))
		args = append(args, f.CardType)
		n++
	}
	if f.MemberID != nil {
		parts = append(parts, fmt.Sprintf("%smember_id = $%d", p, n))
		args = append(args, *f.MemberID)
		n++
	}
	if f.CardID != nil {
		parts = append(parts, fmt.Sprintf("%scard_id = $%d", p, n))
		args = append(args, *f.CardID)
		n++
	}
	if f.From != nil {
		parts = append(parts, fmt.Sprintf("%screated_at >= $%d", p, n))
		args = append(args, *f.From)
		n++
	}
	if f.To != nil {
		parts = append(parts, fmt.Sprintf("%screated_at <= $%d", p, n))
		args = append(args, *f.To)
		n++
	}
	_ = n
	return strings.Join(parts, " AND "), args
}
