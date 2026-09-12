package repository

import (
	"context"
	"time"

	"card_manager/api_service/ent"
	entreftoken "card_manager/api_service/ent/refreshtoken"
	domainrt "card_manager/api_service/internal/domain/refreshtoken"
	"card_manager/api_service/internal/postgres"
)

type refreshTokenRepository struct {
	client *postgres.Client
}

func NewRefreshTokenRepository(client *postgres.Client) domainrt.Repository {
	return &refreshTokenRepository{client: client}
}

func (r *refreshTokenRepository) Create(ctx context.Context, in domainrt.CreateInput) (*domainrt.RefreshToken, error) {
	b := r.client.Ent().RefreshToken.Create().
		SetUserID(in.UserID).
		SetTokenHash(in.TokenHash).
		SetExpiresAt(in.ExpiresAt)
	if in.DeviceID != nil {
		b.SetDeviceID(*in.DeviceID)
	}
	if in.UserAgent != nil {
		b.SetUserAgent(*in.UserAgent)
	}
	if in.IP != nil {
		b.SetIP(*in.IP)
	}
	t, err := b.Save(ctx)
	if err != nil {
		return nil, err
	}
	return domainrt.FromEnt(t), nil
}

func (r *refreshTokenRepository) GetByHash(ctx context.Context, hash string) (*domainrt.RefreshToken, error) {
	t, err := r.client.Ent().RefreshToken.Query().Where(entreftoken.TokenHashEQ(hash)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainrt.FromEnt(t), nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	return r.client.Ent().RefreshToken.UpdateOneID(id).SetRevokedAt(now).Exec(ctx)
}
