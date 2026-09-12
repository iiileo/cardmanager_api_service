package biztype

import (
	"context"
	"time"

	"card_manager/api_service/ent"
)

const StatusActive int8 = 1

type BizType struct {
	ID        int64
	Code      string
	Name      string
	Sort      int
	Status    int8
	CreatedAt time.Time
	UpdatedAt time.Time
}

func FromEnt(b *ent.BizType) *BizType {
	if b == nil {
		return nil
	}
	return &BizType{
		ID:        b.ID,
		Code:      b.Code,
		Name:      b.Name,
		Sort:      b.Sort,
		Status:    b.Status,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
}

type Repository interface {
	ListActive(ctx context.Context) ([]*BizType, error)
	GetByCode(ctx context.Context, code string) (*BizType, error)
	EnsureDefaults(ctx context.Context) error
}
