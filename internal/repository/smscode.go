package repository

import (
	"context"
	"time"

	"card_manager/api_service/ent"
	entsmscode "card_manager/api_service/ent/smscode"
	domainsms "card_manager/api_service/internal/domain/smscode"
	"card_manager/api_service/internal/postgres"
)

type smsCodeRepository struct {
	client *postgres.Client
}

func NewSmsCodeRepository(client *postgres.Client) domainsms.Repository {
	return &smsCodeRepository{client: client}
}

func (r *smsCodeRepository) Create(ctx context.Context, in domainsms.CreateInput) (*domainsms.SmsCode, error) {
	c, err := r.client.Ent().SmsCode.Create().
		SetPhone(in.Phone).
		SetScene(in.Scene).
		SetCodeHash(in.CodeHash).
		SetExpiresAt(in.ExpiresAt).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return domainsms.FromEnt(c), nil
}

func (r *smsCodeRepository) FindLatestValid(ctx context.Context, phone, scene string) (*domainsms.SmsCode, error) {
	now := time.Now().UTC()
	c, err := r.client.Ent().SmsCode.Query().
		Where(
			entsmscode.PhoneEQ(phone),
			entsmscode.SceneEQ(scene),
			entsmscode.ExpiresAtGT(now),
			entsmscode.UsedAtIsNil(),
		).
		Order(ent.Desc(entsmscode.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainsms.FromEnt(c), nil
}

func (r *smsCodeRepository) MarkUsed(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	return r.client.Ent().SmsCode.UpdateOneID(id).SetUsedAt(now).Exec(ctx)
}
