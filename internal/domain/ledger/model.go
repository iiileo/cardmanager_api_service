package ledger

import (
	"context"
	"time"

	"card_manager/api_service/ent"
)

const (
	TypeOpen          = "open"
	TypeRecharge      = "recharge"
	TypeConsumeValue  = "consume_value"
	TypeConsumeCount  = "consume_count"
	TypeConsumePack   = "consume_pack"
)

type EntryItem struct {
	ID            int64
	LedgerID      int64
	ItemBalanceID int64
	ProductItemID int64
	NameSnapshot  string
	Times         int
	TimesAfter    int
}

type Entry struct {
	ID           int64
	StoreID      int64
	MemberID     int64
	CardID       int64
	Type         string
	Amount       *int
	Times        *int
	ItemName     *string
	BalanceAfter *int
	TimesAfter   *int
	Remark       *string
	OperatorID   int64
	Items        []*EntryItem
	CreatedAt    time.Time
}

func ItemFromEnt(it *ent.LedgerEntryItem) *EntryItem {
	if it == nil {
		return nil
	}
	return &EntryItem{
		ID:            it.ID,
		LedgerID:      it.LedgerID,
		ItemBalanceID: it.ItemBalanceID,
		ProductItemID: it.ProductItemID,
		NameSnapshot:  it.NameSnapshot,
		Times:         it.Times,
		TimesAfter:    it.TimesAfter,
	}
}

func FromEnt(e *ent.LedgerEntry, items []*ent.LedgerEntryItem) *Entry {
	if e == nil {
		return nil
	}
	out := &Entry{
		ID:           e.ID,
		StoreID:      e.StoreID,
		MemberID:     e.MemberID,
		CardID:       e.CardID,
		Type:         e.Type,
		Amount:       e.Amount,
		Times:        e.Times,
		ItemName:     e.ItemName,
		BalanceAfter: e.BalanceAfter,
		TimesAfter:   e.TimesAfter,
		Remark:       e.Remark,
		OperatorID:   e.OperatorID,
		CreatedAt:    e.CreatedAt,
	}
	src := items
	if len(src) == 0 {
		src = e.Edges.Items
	}
	if len(src) > 0 {
		out.Items = make([]*EntryItem, 0, len(src))
		for _, it := range src {
			out.Items = append(out.Items, ItemFromEnt(it))
		}
	}
	return out
}

type ItemInput struct {
	ItemBalanceID int64
	ProductItemID int64
	NameSnapshot  string
	Times         int
	TimesAfter    int
}

type CreateInput struct {
	StoreID      int64
	MemberID     int64
	CardID       int64
	Type         string
	Amount       *int
	Times        *int
	ItemName     *string
	BalanceAfter *int
	TimesAfter   *int
	Remark       *string
	OperatorID   int64
	Items        []ItemInput
}

type ListFilter struct {
	StoreID int64
	Type    string
	From    *time.Time
	To      *time.Time
	Limit   int
	Offset  int
}

type Repository interface {
	Create(ctx context.Context, in CreateInput) (*Entry, error)
	GetByID(ctx context.Context, id int64) (*Entry, error)
	List(ctx context.Context, f ListFilter) ([]*Entry, int, error)
}
