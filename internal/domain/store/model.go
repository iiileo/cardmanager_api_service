package store

import (
	"context"
	"time"

	"card_manager/api_service/ent"
)

const (
	RoleOwner = "owner"
	RoleStaff = "staff"

	MemberActive   = "active"
	MemberPending  = "pending"
	MemberRejected = "rejected"
	MemberLeft     = "left"

	StoreStatusNormal int8 = 1
)

type Store struct {
	ID          int64
	Name        string
	City        string
	Address     *string
	OpenTime    string
	CloseTime   string
	BizType     *string
	InviteCode  string
	OwnerUserID int64
	Status      int8
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func FromEnt(s *ent.Store) *Store {
	if s == nil {
		return nil
	}
	return &Store{
		ID:          s.ID,
		Name:        s.Name,
		City:        s.City,
		Address:     s.Address,
		OpenTime:    s.OpenTime,
		CloseTime:   s.CloseTime,
		BizType:     s.BizType,
		InviteCode:  s.InviteCode,
		OwnerUserID: s.OwnerUserID,
		Status:      s.Status,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

type CreateInput struct {
	Name        string
	City        string
	Address     *string
	OpenTime    string
	CloseTime   string
	BizType     *string
	InviteCode  string
	OwnerUserID int64
}

type UpdateInput struct {
	Name      *string
	City      *string
	Address   *string
	OpenTime  *string
	CloseTime *string
	BizType   *string
}

type Repository interface {
	Create(ctx context.Context, in CreateInput) (*Store, error)
	// CreateWithOwner 在同一事务中创建门店并写入老板成员。
	CreateWithOwner(ctx context.Context, in CreateInput) (*Store, error)
	GetByID(ctx context.Context, id int64) (*Store, error)
	GetByInviteCode(ctx context.Context, code string) (*Store, error)
	Update(ctx context.Context, id int64, in UpdateInput) (*Store, error)
	UpdateInviteCode(ctx context.Context, id int64, code string) (*Store, error)
	InviteCodeExists(ctx context.Context, code string) (bool, error)
}
