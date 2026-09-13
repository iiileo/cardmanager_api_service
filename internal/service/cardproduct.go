package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"card_manager/api_service/internal/api/dto"
	domainproduct "card_manager/api_service/internal/domain/cardproduct"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
)

type CardProductService interface {
	List(ctx context.Context, userID, storeID int64, typ string) (*dto.CardProductListResponse, error)
	Create(ctx context.Context, userID, storeID int64, req dto.CreateCardProductRequest) (*dto.CardProductResponse, error)
	Delete(ctx context.Context, userID, storeID, productID int64) error
}

type cardProductService struct {
	products domainproduct.Repository
	storeSvc StoreService
	log      *logger.Logger
}

func NewCardProductService(
	products domainproduct.Repository,
	storeSvc StoreService,
	log *logger.Logger,
) CardProductService {
	return &cardProductService{products: products, storeSvc: storeSvc, log: log}
}

func (s *cardProductService) List(ctx context.Context, userID, storeID int64, typ string) (*dto.CardProductListResponse, error) {
	if _, err := s.storeSvc.RequireActiveMember(ctx, userID, storeID); err != nil {
		return nil, err
	}
	typ = strings.TrimSpace(typ)
	if typ != "" && typ != domainproduct.TypeValue && typ != domainproduct.TypeCount && typ != domainproduct.TypePack {
		return nil, ierr.Validation("卡种类型不正确")
	}
	list, err := s.products.ListByStore(ctx, storeID, typ)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	out := make([]*dto.CardProductResponse, 0, len(list))
	for _, p := range list {
		out = append(out, toCardProductDTO(p))
	}
	return &dto.CardProductListResponse{List: out}, nil
}

func (s *cardProductService) Create(ctx context.Context, userID, storeID int64, req dto.CreateCardProductRequest) (*dto.CardProductResponse, error) {
	if _, err := s.storeSvc.RequireOwner(ctx, userID, storeID); err != nil {
		return nil, err
	}

	typ := strings.TrimSpace(req.Type)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ierr.Validation("请填写卡种名称")
	}
	if typ != domainproduct.TypeValue && typ != domainproduct.TypeCount && typ != domainproduct.TypePack {
		return nil, ierr.Validation("卡种类型不正确")
	}

	if typ == domainproduct.TypeValue || typ == domainproduct.TypeCount {
		n, err := s.products.CountActiveByStoreType(ctx, storeID, typ)
		if err != nil {
			return nil, ierr.Internal(err)
		}
		if n > 0 {
			if typ == domainproduct.TypeValue {
				return nil, ierr.Conflict("储值卡只能创建一个")
			}
			return nil, ierr.Conflict("次卡只能创建一个")
		}
	}

	price := 0
	if req.Price != nil {
		if *req.Price < 0 {
			return nil, ierr.Validation("售价不能为负数")
		}
		price = *req.Price
	}

	var times *int
	var validMonths *int
	var items []domainproduct.ItemInput

	switch typ {
	case domainproduct.TypeValue:
		// 储值卡：无次数、不限期
		if req.Times != nil || (req.ValidMonths != nil && *req.ValidMonths > 0) || len(req.Items) > 0 {
			return nil, ierr.Validation("储值卡无需次数、有效期或套餐项目")
		}
	case domainproduct.TypeCount:
		if req.Times == nil || *req.Times <= 0 {
			return nil, ierr.Validation("请填写次卡次数")
		}
		if price <= 0 {
			return nil, ierr.Validation("请填写次卡售价")
		}
		if len(req.Items) > 0 {
			return nil, ierr.Validation("次卡不支持套餐项目")
		}
		t := *req.Times
		times = &t
		validMonths = normalizeValidMonths(req.ValidMonths)
	case domainproduct.TypePack:
		if price <= 0 {
			return nil, ierr.Validation("请填写套餐卡售价")
		}
		if len(req.Items) == 0 {
			return nil, ierr.Validation("请至少添加一个套餐项目")
		}
		items = make([]domainproduct.ItemInput, 0, len(req.Items))
		for i, it := range req.Items {
			n := strings.TrimSpace(it.Name)
			if n == "" {
				return nil, ierr.Validation("套餐项目名称不能为空")
			}
			if it.Times <= 0 {
				return nil, ierr.Validation("套餐项目次数须大于 0")
			}
			items = append(items, domainproduct.ItemInput{
				Name:  n,
				Times: it.Times,
				Sort:  i + 1,
			})
		}
		validMonths = normalizeValidMonths(req.ValidMonths)
	}

	p, err := s.products.Create(ctx, domainproduct.CreateInput{
		StoreID:     storeID,
		Type:        typ,
		Name:        name,
		Price:       price,
		Times:       times,
		ValidMonths: validMonths,
		Items:       items,
	})
	if err != nil {
		return nil, ierr.Internal(err)
	}
	s.log.Info(ctx, "card product created", "store_id", storeID, "product_id", p.ID, "type", typ)
	return toCardProductDTO(p), nil
}

func (s *cardProductService) Delete(ctx context.Context, userID, storeID, productID int64) error {
	if _, err := s.storeSvc.RequireOwner(ctx, userID, storeID); err != nil {
		return err
	}
	p, err := s.products.GetByID(ctx, productID)
	if err != nil {
		return ierr.Internal(err)
	}
	if p == nil || p.StoreID != storeID {
		return ierr.NotFound("卡种不存在")
	}
	if err := s.products.Delete(ctx, productID); err != nil {
		return ierr.Internal(err)
	}
	s.log.Info(ctx, "card product deleted", "store_id", storeID, "product_id", productID)
	return nil
}

func normalizeValidMonths(v *int) *int {
	if v == nil || *v <= 0 {
		return nil
	}
	x := *v
	return &x
}

func toCardProductDTO(p *domainproduct.Product) *dto.CardProductResponse {
	resp := &dto.CardProductResponse{
		ID:          fmt.Sprintf("%d", p.ID),
		Type:        p.Type,
		Name:        p.Name,
		Price:       p.Price,
		Times:       p.Times,
		ValidMonths: p.ValidMonths,
		Status:      p.Status,
	}
	if len(p.Items) > 0 {
		resp.Items = make([]*dto.CardProductItemResponse, 0, len(p.Items))
		for _, it := range p.Items {
			resp.Items = append(resp.Items, &dto.CardProductItemResponse{
				ID:    fmt.Sprintf("%d", it.ID),
				Name:  it.Name,
				Times: it.Times,
				Sort:  it.Sort,
			})
		}
	}
	return resp
}

func ParseProductID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, ierr.Validation("卡种 ID 不正确")
	}
	return id, nil
}