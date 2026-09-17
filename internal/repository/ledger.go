package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"card_manager/api_service/ent"
	entledger "card_manager/api_service/ent/ledgerentry"
	domainledger "card_manager/api_service/internal/domain/ledger"
	domainmember "card_manager/api_service/internal/domain/member"
	domaincard "card_manager/api_service/internal/domain/membercard"
	domainuser "card_manager/api_service/internal/domain/user"
	"card_manager/api_service/internal/postgres"
	"golang.org/x/sync/errgroup"
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
		WithMember().
		WithCard().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	entry := ledgerEntryFromEnt(e)
	if u, err := r.client.Ent().User.Get(ctx, e.OperatorID); err == nil {
		entry.Operator = domainuser.FromEnt(u)
	} else if !ent.IsNotFound(err) {
		return nil, err
	}
	return entry, nil
}

func (r *ledgerRepository) List(ctx context.Context, f domainledger.ListFilter) ([]*domainledger.Entry, int, error) {
	// 单条 JOIN SQL + 批量 items，减少远程 PG 往返。
	limit := f.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	sf := summaryFilterFromList(f)
	where, args := buildSummaryWherePrefixed(sf, "le")
	listSQL := `
SELECT
  le.id, le.store_id, le.member_id, le.card_id, le.type, le.card_type,
  le.amount, le.times, le.item_name, le.balance_after, le.times_after,
  le.remark, le.operator_id, le.created_at,
  m.name, m.phone,
  mc.type, mc.name_snapshot, mc.balance, mc.remain_times, mc.status, mc.valid_to,
  u.nickname
FROM ledger_entries le
LEFT JOIN members m ON m.id = le.member_id
LEFT JOIN member_cards mc ON mc.id = le.card_id
LEFT JOIN users u ON u.id = le.operator_id
WHERE ` + where + `
ORDER BY le.created_at DESC, le.id DESC
LIMIT $` + fmt.Sprintf("%d", len(args)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(args)+2)

	listArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.client.DB().QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list ledger entries: %w", err)
	}
	defer rows.Close()

	entries := make([]*domainledger.Entry, 0, limit)
	ledgerIDs := make([]int64, 0, limit)
	for rows.Next() {
		entry, err := scanLedgerListRow(rows)
		if err != nil {
			return nil, 0, err
		}
		entries = append(entries, entry)
		ledgerIDs = append(ledgerIDs, entry.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	total := offset + len(entries)
	var itemsByLedger map[int64][]*domainledger.EntryItem
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		itemsByLedger, err = r.loadItemsByLedgerIDs(gctx, ledgerIDs)
		return err
	})
	if len(entries) == limit {
		g.Go(func() error {
			cnt, err := r.countList(gctx, sf)
			if err != nil {
				return err
			}
			total = cnt
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, 0, err
	}
	for _, e := range entries {
		if items := itemsByLedger[e.ID]; len(items) > 0 {
			e.Items = items
		}
	}
	return entries, total, nil
}

func (r *ledgerRepository) countList(ctx context.Context, f domainledger.SummaryFilter) (int, error) {
	where, args := buildSummaryWherePrefixed(f, "le")
	countSQL := `SELECT COUNT(*)::int FROM ledger_entries le WHERE ` + where
	var total int
	if err := r.client.DB().QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count ledger entries: %w", err)
	}
	return total, nil
}

func (r *ledgerRepository) loadItemsByLedgerIDs(ctx context.Context, ledgerIDs []int64) (map[int64][]*domainledger.EntryItem, error) {
	out := make(map[int64][]*domainledger.EntryItem)
	if len(ledgerIDs) == 0 {
		return out, nil
	}
	ph := make([]string, len(ledgerIDs))
	args := make([]any, len(ledgerIDs))
	for i, id := range ledgerIDs {
		ph[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	itemSQL := `
SELECT id, ledger_id, item_balance_id, product_item_id, name_snapshot, times, times_after
FROM ledger_entry_items
WHERE ledger_id IN (` + strings.Join(ph, ",") + `)
ORDER BY ledger_id, id`
	rows, err := r.client.DB().QueryContext(ctx, itemSQL, args...)
	if err != nil {
		return nil, fmt.Errorf("list ledger items: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var it domainledger.EntryItem
		if err := rows.Scan(&it.ID, &it.LedgerID, &it.ItemBalanceID, &it.ProductItemID, &it.NameSnapshot, &it.Times, &it.TimesAfter); err != nil {
			return nil, err
		}
		out[it.LedgerID] = append(out[it.LedgerID], &it)
	}
	return out, rows.Err()
}

func scanLedgerListRow(rows *sql.Rows) (*domainledger.Entry, error) {
	var (
		entry                                   domainledger.Entry
		amount, times, balanceAfter, timesAfter sql.NullInt64
		itemName, remark                          sql.NullString
		memberName, memberPhone                   sql.NullString
		cardType, cardName, cardStatus            sql.NullString
		cardBalance                               sql.NullInt64
		cardRemainTimes                           sql.NullInt64
		cardValidTo                               sql.NullTime
		operatorNickname                          sql.NullString
	)
	if err := rows.Scan(
		&entry.ID, &entry.StoreID, &entry.MemberID, &entry.CardID, &entry.Type, &entry.CardType,
		&amount, &times, &itemName, &balanceAfter, &timesAfter,
		&remark, &entry.OperatorID, &entry.CreatedAt,
		&memberName, &memberPhone,
		&cardType, &cardName, &cardBalance, &cardRemainTimes, &cardStatus, &cardValidTo,
		&operatorNickname,
	); err != nil {
		return nil, err
	}
	if amount.Valid {
		v := int(amount.Int64)
		entry.Amount = &v
	}
	if times.Valid {
		v := int(times.Int64)
		entry.Times = &v
	}
	if itemName.Valid {
		v := itemName.String
		entry.ItemName = &v
	}
	if balanceAfter.Valid {
		v := int(balanceAfter.Int64)
		entry.BalanceAfter = &v
	}
	if timesAfter.Valid {
		v := int(timesAfter.Int64)
		entry.TimesAfter = &v
	}
	if remark.Valid {
		v := remark.String
		entry.Remark = &v
	}
	if memberName.Valid {
		entry.Member = &domainmember.Member{
			ID: entry.MemberID, StoreID: entry.StoreID,
			Name: memberName.String, Phone: memberPhone.String,
		}
	}
	if cardType.Valid {
		card := &domaincard.Card{
			ID: entry.CardID, StoreID: entry.StoreID, MemberID: entry.MemberID,
			Type: cardType.String, NameSnapshot: cardName.String, Status: cardStatus.String,
		}
		if cardBalance.Valid {
			card.Balance = int(cardBalance.Int64)
		}
		if cardRemainTimes.Valid {
			v := int(cardRemainTimes.Int64)
			card.RemainTimes = &v
		}
		if cardValidTo.Valid {
			t := cardValidTo.Time
			card.ValidTo = &t
		}
		entry.Card = card
	}
	if operatorNickname.Valid {
		entry.Operator = &domainuser.User{
			ID: entry.OperatorID, Nickname: operatorNickname.String,
		}
	}
	return &entry, nil
}

func summaryFilterFromList(f domainledger.ListFilter) domainledger.SummaryFilter {
	types := f.Types
	if len(types) == 0 && f.Type != "" {
		types = []string{f.Type}
	}
	return domainledger.SummaryFilter{
		StoreID:  f.StoreID,
		Types:    types,
		CardType: f.CardType,
		MemberID: f.MemberID,
		CardID:   f.CardID,
		From:     f.From,
		To:       f.To,
	}
}

func ledgerEntryFromEnt(e *ent.LedgerEntry) *domainledger.Entry {
	entry := domainledger.FromEnt(e, nil)
	if m := e.Edges.Member; m != nil {
		entry.Member = domainmember.FromEnt(m)
	}
	if c := e.Edges.Card; c != nil {
		entry.Card = domaincard.FromEnt(c, nil)
	}
	return entry
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
