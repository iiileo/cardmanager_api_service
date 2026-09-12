package storemember

import (
	"context"
	"time"

	"card_manager/api_service/ent"
)

type Member struct {
	ID          int64
	StoreID     int64
	UserID      int64
	Role        string
	Status      string
	DisplayName *string
	JoinedAt    *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func FromEnt(m *ent.StoreMember) *Member {
	if m == nil {
		return nil
	}
	return &Member{
		ID:          m.ID,
		StoreID:     m.StoreID,
		UserID:      m.UserID,
		Role:        m.Role,
		Status:      m.Status,
		DisplayName: m.DisplayName,
		JoinedAt:    m.JoinedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

type CreateInput struct {
	StoreID     int64
	UserID      int64
	Role        string
	Status      string
	DisplayName *string
	JoinedAt    *time.Time
}

type Repository interface {
	Create(ctx context.Context, in CreateInput) (*Member, error)
	GetByStoreUser(ctx context.Context, storeID, userID int64) (*Member, error)
	ListByUser(ctx context.Context, userID int64) ([]*Member, error)
	ListByStore(ctx context.Context, storeID int64, statuses []string) ([]*Member, error)
	CountActiveByStore(ctx context.Context, storeID int64) (int, error)
	UpdateStatus(ctx context.Context, id int64, status string, joinedAt *time.Time, displayName *string) (*Member, error)
	GetByID(ctx context.Context, id int64) (*Member, error)
}
