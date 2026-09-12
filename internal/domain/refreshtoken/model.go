package refreshtoken

import (
	"context"
	"time"

	"card_manager/api_service/ent"
)

type RefreshToken struct {
	ID        int64
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	DeviceID  *string
	UserAgent *string
	IP        *string
	CreatedAt time.Time
}

func FromEnt(t *ent.RefreshToken) *RefreshToken {
	if t == nil {
		return nil
	}
	return &RefreshToken{
		ID:        t.ID,
		UserID:    t.UserID,
		TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt,
		RevokedAt: t.RevokedAt,
		DeviceID:  t.DeviceID,
		UserAgent: t.UserAgent,
		IP:        t.IP,
		CreatedAt: t.CreatedAt,
	}
}

type CreateInput struct {
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
	DeviceID  *string
	UserAgent *string
	IP        *string
}

type Repository interface {
	Create(ctx context.Context, in CreateInput) (*RefreshToken, error)
	GetByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id int64) error
}
