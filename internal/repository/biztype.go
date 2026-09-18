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
	// 面向服务行业：老板/员工用 App 记次卡/储值，客户无需操作 App。
	defaults := []struct {
		Code   string
		Name   string
		Sort   int
		Status int8
	}{
		{Code: "beauty", Name: "美业", Sort: 10, Status: domainbiz.StatusActive},         // 美发美甲护肤纹绣等
		{Code: "spa", Name: "养生保健", Sort: 20, Status: domainbiz.StatusActive},       // 推拿理疗足疗按摩
		{Code: "fitness", Name: "健身运动", Sort: 30, Status: domainbiz.StatusActive},   // 瑜伽普拉提健身房
		{Code: "education", Name: "教育培训", Sort: 40, Status: domainbiz.StatusActive}, // 兴趣班驾校早教
		{Code: "pet", Name: "宠物服务", Sort: 50, Status: domainbiz.StatusActive},       // 洗护美容寄养
		{Code: "auto", Name: "汽车服务", Sort: 60, Status: domainbiz.StatusActive},      // 洗车美容保养
		{Code: "photo", Name: "摄影写真", Sort: 70, Status: domainbiz.StatusActive},
		{Code: "other", Name: "其他服务", Sort: 80, Status: domainbiz.StatusActive},
		// 旧业态：停用（不再出现在列表；已有门店的 biz_type 字符串保留）
		{Code: "tea", Name: "茶饮", Sort: 900, Status: domainbiz.StatusInactive},
		{Code: "retail", Name: "零售", Sort: 910, Status: domainbiz.StatusInactive},
	}
	for _, d := range defaults {
		existing, err := r.client.Ent().BizType.Query().Where(entbiz.CodeEQ(d.Code)).Only(ctx)
		if err != nil {
			if !ent.IsNotFound(err) {
				return err
			}
			if _, err := r.client.Ent().BizType.Create().
				SetCode(d.Code).
				SetName(d.Name).
				SetSort(d.Sort).
				SetStatus(d.Status).
				Save(ctx); err != nil {
				return err
			}
			continue
		}
		if _, err := existing.Update().
			SetName(d.Name).
			SetSort(d.Sort).
			SetStatus(d.Status).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}
