package repository

import (
	"context"
	"fmt"

	"card_manager/api_service/ent"
	"card_manager/api_service/ent/user"
	domainuser "card_manager/api_service/internal/domain/user"
	"card_manager/api_service/internal/postgres"
)

type userRepository struct {
	client *postgres.Client
}

func NewUserRepository(client *postgres.Client) domainuser.Repository {
	return &userRepository{client: client}
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domainuser.User, error) {
	u, err := r.client.Ent().User.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainuser.FromEnt(u), nil
}

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*domainuser.User, error) {
	u, err := r.client.Ent().User.Query().Where(user.PhoneEQ(phone)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainuser.FromEnt(u), nil
}

func (r *userRepository) Create(ctx context.Context, phone, nickname string) (*domainuser.User, error) {
	if nickname == "" {
		nickname = maskPhone(phone)
	}
	u, err := r.client.Ent().User.Create().
		SetPhone(phone).
		SetNickname(nickname).
		SetStatus(1).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return domainuser.FromEnt(u), nil
}

func (r *userRepository) UpdateNickname(ctx context.Context, id int64, nickname string) (*domainuser.User, error) {
	u, err := r.client.Ent().User.UpdateOneID(id).SetNickname(nickname).Save(ctx)
	if err != nil {
		return nil, err
	}
	return domainuser.FromEnt(u), nil
}

func (r *userRepository) UpdatePhone(ctx context.Context, id int64, phone string) (*domainuser.User, error) {
	u, err := r.client.Ent().User.UpdateOneID(id).SetPhone(phone).Save(ctx)
	if err != nil {
		return nil, err
	}
	return domainuser.FromEnt(u), nil
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return fmt.Sprintf("%s****%s", phone[:3], phone[len(phone)-4:])
}
