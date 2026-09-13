package ledger

import (
	"context"
	"time"

	"card_manager/api_service/ent"
)

const (
	TypeOpen         = "open"
	TypeRecharge     = "recharge"
	TypeConsumeValue = "consume_value"
	TypeConsumeCount = "consume_count"
	TypeConsumePack  = "consume_pack"
)

func ConsumeTypes() []string {
	return []string{TypeConsumeValue, TypeConsumeCount, TypeConsumePack}
}

func IsConsumeType(typ string) bool {
	switch typ {
	case TypeConsumeValue, TypeConsumeCount, TypeConsumePack:
		return true
	default:
		return false
	}
}

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
	CardType     string
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
		CardType:     e.CardType,
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
	CardType     string
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
	StoreID  int64
	Types    []string
	Type     string // 兼容单类型；与 Types 二选一，Types 优先
	CardType string
	MemberID *int64
	CardID   *int64
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}

type TypeBucket struct {
	Type     string
	CardType string
	Count    int
	Amount   int // 金额合计（充值为正；储值消费为正的绝对值）
	Times    int // 次数合计（正数）
}

type PackItemBucket struct {
	ProductItemID int64
	Name          string
	Times         int
	Count         int
}

type Summary struct {
	TotalCount  int
	TotalAmount int
	TotalTimes  int
	ByType      []TypeBucket
	ByPackItem  []PackItemBucket
}

type SummaryFilter struct {
	StoreID  int64
	Types    []string
	CardType string
	MemberID *int64
	CardID   *int64
	From     *time.Time
	To       *time.Time
}

type Repository interface {
	Create(ctx context.Context, in CreateInput) (*Entry, error)
	GetByID(ctx context.Context, id int64) (*Entry, error)
	List(ctx context.Context, f ListFilter) ([]*Entry, int, error)
	Summarize(ctx context.Context, f SummaryFilter) (*Summary, error)
}
