package member

import (
	"context"
	"time"

	"card_manager/api_service/ent"
)

type Member struct {
	ID        int64
	StoreID   int64
	Name      string
	Phone     string
	Source    *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func FromEnt(m *ent.Member) *Member {
	if m == nil {
		return nil
	}
	return &Member{
		ID:        m.ID,
		StoreID:   m.StoreID,
		Name:      m.Name,
		Phone:     m.Phone,
		Source:    m.Source,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type CreateInput struct {
	StoreID int64
	Name    string
	Phone   string
	Source  *string
}

type Repository interface {
	GetByID(ctx context.Context, id int64) (*Member, error)
	GetByStorePhone(ctx context.Context, storeID int64, phone string) (*Member, error)
	Create(ctx context.Context, in CreateInput) (*Member, error)
	ListByStore(ctx context.Context, storeID int64, q string, limit, offset int) ([]*Member, int, error)
}
