package membercard

import (
	"context"
	"time"

	"card_manager/api_service/ent"
)

const (
	TypeValue = "value"
	TypeCount = "count"
	TypePack  = "pack"

	StatusActive    = "active"
	StatusExpired   = "expired"
	StatusExhausted = "exhausted"
)

type ItemBalance struct {
	ID            int64
	CardID        int64
	ProductItemID int64
	NameSnapshot  string
	RemainTimes   int
}

type Card struct {
	ID           int64
	StoreID      int64
	MemberID     int64
	ProductID    int64
	Type         string
	NameSnapshot string
	Balance      int
	RemainTimes  *int
	ValidFrom    *time.Time
	ValidTo      *time.Time
	Status       string
	OpenedBy     int64
	Items        []*ItemBalance
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func ItemFromEnt(it *ent.CardItemBalance) *ItemBalance {
	if it == nil {
		return nil
	}
	return &ItemBalance{
		ID:            it.ID,
		CardID:        it.CardID,
		ProductItemID: it.ProductItemID,
		NameSnapshot:  it.NameSnapshot,
		RemainTimes:   it.RemainTimes,
	}
}

func FromEnt(c *ent.MemberCard, items []*ent.CardItemBalance) *Card {
	if c == nil {
		return nil
	}
	out := &Card{
		ID:           c.ID,
		StoreID:      c.StoreID,
		MemberID:     c.MemberID,
		ProductID:    c.ProductID,
		Type:         c.Type,
		NameSnapshot: c.NameSnapshot,
		Balance:      c.Balance,
		RemainTimes:  c.RemainTimes,
		ValidFrom:    c.ValidFrom,
		ValidTo:      c.ValidTo,
		Status:       c.Status,
		OpenedBy:     c.OpenedBy,
		CreatedAt:    c.CreatedAt,
		UpdatedAt:    c.UpdatedAt,
	}
	if len(items) > 0 {
		out.Items = make([]*ItemBalance, 0, len(items))
		for _, it := range items {
			out.Items = append(out.Items, ItemFromEnt(it))
		}
	} else if c.Edges.ItemBalances != nil {
		out.Items = make([]*ItemBalance, 0, len(c.Edges.ItemBalances))
		for _, it := range c.Edges.ItemBalances {
			out.Items = append(out.Items, ItemFromEnt(it))
		}
	}
	return out
}

type ItemBalanceInput struct {
	ProductItemID int64
	NameSnapshot  string
	RemainTimes   int
}

type CreateInput struct {
	StoreID      int64
	MemberID     int64
	ProductID    int64
	Type         string
	NameSnapshot string
	Balance      int
	RemainTimes  *int
	ValidFrom    *time.Time
	ValidTo      *time.Time
	OpenedBy     int64
	Items        []ItemBalanceInput
}

type PackDeductItem struct {
	ItemBalanceID int64
	Times         int
}

type Repository interface {
	GetByID(ctx context.Context, id int64) (*Card, error)
	GetByMemberAndType(ctx context.Context, storeID, memberID int64, typ string) (*Card, error)
	ListByMember(ctx context.Context, storeID, memberID int64) ([]*Card, error)
}
