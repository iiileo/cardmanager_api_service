package service

import (
	"context"

	"card_manager/api_service/internal/api/dto"
	domainbiz "card_manager/api_service/internal/domain/biztype"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
)

type BizTypeService interface {
	List(ctx context.Context) (*dto.BizTypeListResponse, error)
	ValidateCode(ctx context.Context, code string) error
}

type bizTypeService struct {
	repo domainbiz.Repository
	log  *logger.Logger
}

func NewBizTypeService(repo domainbiz.Repository, log *logger.Logger) BizTypeService {
	return &bizTypeService{repo: repo, log: log}
}

func (s *bizTypeService) List(ctx context.Context) (*dto.BizTypeListResponse, error) {
	list, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	out := make([]*dto.BizTypeItem, 0, len(list))
	for _, item := range list {
		out = append(out, &dto.BizTypeItem{
			Code: item.Code,
			Name: item.Name,
		})
	}
	return &dto.BizTypeListResponse{List: out}, nil
}

func (s *bizTypeService) ValidateCode(ctx context.Context, code string) error {
	if code == "" {
		return nil
	}
	b, err := s.repo.GetByCode(ctx, code)
	if err != nil {
		return ierr.Internal(err)
	}
	if b == nil {
		return ierr.Validation("业态不存在或已停用")
	}
	return nil
}
