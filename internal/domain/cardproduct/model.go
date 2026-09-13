package cardproduct

import (
	"context"
	"time"

	"card_manager/api_service/ent"
)

const (
	TypeValue = "value"
	TypeCount = "count"
	TypePack  = "pack"

	StatusActive   int8 = 1
	StatusDisabled int8 = 0
)

type Item struct {
	ID        int64
	ProductID int64
	Name      string
	Times     int
	Sort      int
}

type Product struct {
	ID          int64
	StoreID     int64
	Type        string
	Name        string
	Price       int
	Times       *int
	ValidMonths *int
	Status      int8
	Items       []*Item
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func FromEnt(p *ent.CardProduct, items []*ent.CardProductItem) *Product {
	if p == nil {
		return nil
	}
	out := &Product{
		ID:          p.ID,
		StoreID:     p.StoreID,
		Type:        p.Type,
		Name:        p.Name,
		Price:       p.Price,
		Times:       p.Times,
		ValidMonths: p.ValidMonths,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	if len(items) > 0 {
		out.Items = make([]*Item, 0, len(items))
		for _, it := range items {
			out.Items = append(out.Items, ItemFromEnt(it))
		}
	}
	return out
}

func ItemFromEnt(it *ent.CardProductItem) *Item {
	if it == nil {
		return nil
	}
	return &Item{
		ID:        it.ID,
		ProductID: it.ProductID,
		Name:      it.Name,
		Times:     it.Times,
		Sort:      it.Sort,
	}
}

type ItemInput struct {
	Name  string
	Times int
	Sort  int
}

type CreateInput struct {
	StoreID     int64
	Type        string
	Name        string
	Price       int
	Times       *int
	ValidMonths *int
	Items       []ItemInput
}

type Repository interface {
	ListByStore(ctx context.Context, storeID int64, typ string) ([]*Product, error)
	GetByID(ctx context.Context, id int64) (*Product, error)
	CountActiveByStoreType(ctx context.Context, storeID int64, typ string) (int, error)
	Create(ctx context.Context, in CreateInput) (*Product, error)
	Delete(ctx context.Context, id int64) error
}
