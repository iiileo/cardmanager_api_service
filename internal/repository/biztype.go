package repository

import (
	"context"

	"card_manager/api_service/ent"
	entbiz "card_manager/api_service/ent/biztype"
	domainbiz "card_manager/api_service/internal/domain/biztype"
	"card_manager/api_service/internal/postgres"
)

type bizTypeRepository struct {
	client *postgres.Client
}

func NewBizTypeRepository(client *postgres.Client) domainbiz.Repository {
	return &bizTypeRepository{client: client}
}

func (r *bizTypeRepository) ListActive(ctx context.Context) ([]*domainbiz.BizType, error) {
	list, err := r.client.Ent().BizType.Query().
		Where(entbiz.StatusEQ(domainbiz.StatusActive)).
		Order(ent.Asc(entbiz.FieldSort), ent.Asc(entbiz.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domainbiz.BizType, 0, len(list))
	for _, item := range list {
		out = append(out, domainbiz.FromEnt(item))
	}
	return out, nil
}

func (r *bizTypeRepository) GetByCode(ctx context.Context, code string) (*domainbiz.BizType, error) {
	b, err := r.client.Ent().BizType.Query().
		Where(entbiz.CodeEQ(code), entbiz.StatusEQ(domainbiz.StatusActive)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainbiz.FromEnt(b), nil
}

func (r *bizTypeRepository) EnsureDefaults(ctx context.Context) error {
	defaults := []struct {
		Code string
		Name string
		Sort int
	}{
		{Code: "tea", Name: "茶饮", Sort: 10},
		{Code: "beauty", Name: "美业", Sort: 20},
		{Code: "retail", Name: "零售", Sort: 30},
	}
	for _, d := range defaults {
		exists, err := r.client.Ent().BizType.Query().Where(entbiz.CodeEQ(d.Code)).Exist(ctx)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if _, err := r.client.Ent().BizType.Create().
			SetCode(d.Code).
			SetName(d.Name).
			SetSort(d.Sort).
			SetStatus(domainbiz.StatusActive).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}
