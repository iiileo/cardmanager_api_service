package user

import (
	"context"
	"time"

	"card_manager/api_service/ent"
)

type User struct {
	ID        int64
	Phone     string
	Nickname  string
	Status    int8
	CreatedAt time.Time
	UpdatedAt time.Time
}

func FromEnt(u *ent.User) *User {
	if u == nil {
		return nil
	}
	return &User{
		ID:        u.ID,
		Phone:     u.Phone,
		Nickname:  u.Nickname,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

type Repository interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	Create(ctx context.Context, phone, nickname string) (*User, error)
	UpdateNickname(ctx context.Context, id int64, nickname string) (*User, error)
}
