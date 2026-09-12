package smscode

import (
	"context"
	"time"

	"card_manager/api_service/ent"
)

type SmsCode struct {
	ID        int64
	Phone     string
	Scene     string
	CodeHash  string
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

func FromEnt(c *ent.SmsCode) *SmsCode {
	if c == nil {
		return nil
	}
	return &SmsCode{
		ID:        c.ID,
		Phone:     c.Phone,
		Scene:     c.Scene,
		CodeHash:  c.CodeHash,
		ExpiresAt: c.ExpiresAt,
		UsedAt:    c.UsedAt,
		CreatedAt: c.CreatedAt,
	}
}

type CreateInput struct {
	Phone     string
	Scene     string
	CodeHash  string
	ExpiresAt time.Time
}

type Repository interface {
	Create(ctx context.Context, in CreateInput) (*SmsCode, error)
	FindLatestValid(ctx context.Context, phone, scene string) (*SmsCode, error)
	MarkUsed(ctx context.Context, id int64) error
}
