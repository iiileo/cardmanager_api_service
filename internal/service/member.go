package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

"card_manager/api_service/ent"
	"card_manager/api_service/internal/api/dto"
	domainproduct "card_manager/api_service/internal/domain/cardproduct"
	domainledger "card_manager/api_service/internal/domain/ledger"
	domainmember "card_manager/api_service/internal/domain/member"
	domaincard "card_manager/api_service/internal/domain/membercard"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/repository"
)

type MemberService interface {
	List(ctx context.Context, userID, storeID int64, q string, page, pageSize int) (*dto.MemberListResponse, error)
	Get(ctx context.Context, userID, storeID, memberID int64) (*dto.MemberDetailResponse, error)
	OpenCard(ctx context.Context, userID, storeID int64, req dto.OpenCardRequest) (*dto.OpenCardResponse, error)
}

type memberService struct {
	members  domainmember.Repository
	cards    domaincard.Repository
	products domainproduct.Repository
	txns     repository.CardTxnRepository
	storeSvc StoreService
	log      *logger.Logger
}

func NewMemberService(
	members domainmember.Repository,
	cards domaincard.Repository,
	products domainproduct.Repository,
	txns repository.CardTxnRepository,
	storeSvc StoreService,
	log *logger.Logger,
) MemberService {
	return &memberService{
		members: members, cards: cards, products: products,
		txns: txns, storeSvc: storeSvc, log: log,
	}
}

func (s *memberService) List(ctx context.Context, userID, storeID int64, q string, page, pageSize int) (*dto.MemberListResponse, error) {
	if _, err := s.storeSvc.RequireActiveMember(ctx, userID, storeID); err != nil {
		return nil, err
	}
	page, pageSize = normalizePage(page, pageSize)
	list, total, err := s.members.ListByStore(ctx, storeID, q, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	out := make([]*dto.MemberListItem, 0, len(list))
	for _, m := range list {
		out = append(out, &dto.MemberListItem{
			ID: fmt.Sprintf("%d", m.ID), Name: m.Name, Phone: m.Phone,
		})
	}
	pg := dto.NewListPage(page, pageSize, total)
	return &dto.MemberListResponse{
		List: out, Total: total,
		Page: pg.Page, PageSize: pg.PageSize, HasMore: pg.HasMore,
	}, nil
}

func (s *memberService) Get(ctx context.Context, userID, storeID, memberID int64) (*dto.MemberDetailResponse, error) {
	if _, err := s.storeSvc.RequireActiveMember(ctx, userID, storeID); err != nil {
		return nil, err
	}
	m, err := s.members.GetByID(ctx, memberID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	if m == nil || m.StoreID != storeID {
		return nil, ierr.NotFound("会员不存在")
	}
	cards, err := s.cards.ListByMember(ctx, storeID, memberID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	briefs := make([]*dto.MemberCardBrief, 0, len(cards))
	for _, c := range cards {
		briefs = append(briefs, toCardBrief(c))
	}
	return &dto.MemberDetailResponse{
		ID: fmt.Sprintf("%d", m.ID), Name: m.Name, Phone: m.Phone, Cards: briefs,
	}, nil
}

func (s *memberService) OpenCard(ctx context.Context, userID, storeID int64, req dto.OpenCardRequest) (*dto.OpenCardResponse, error) {
	if _, err := s.storeSvc.RequireActiveMember(ctx, userID, storeID); err != nil {
		return nil, err
	}
	productID, err := ParseProductID(req.ProductID)
	if err != nil {
		return nil, err
	}
	product, err := s.products.GetByID(ctx, productID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	if product == nil || product.StoreID != storeID || product.Status != domainproduct.StatusActive {
		return nil, ierr.NotFound("卡种不存在或已停用")
	}

	var m *domainmember.Member
	if req.MemberID != nil && strings.TrimSpace(*req.MemberID) != "" {
		mid, err := ParseCustomerID(*req.MemberID)
		if err != nil {
			return nil, err
		}
		m, err = s.members.GetByID(ctx, mid)
		if err != nil {
			return nil, ierr.Internal(err)
		}
		if m == nil || m.StoreID != storeID {
			return nil, ierr.NotFound("会员不存在")
		}
	} else {
		name := strings.TrimSpace(req.Name)
		phone := strings.TrimSpace(req.Phone)
		if name == "" || phone == "" {
			return nil, ierr.Validation("请填写会员姓名和手机号")
		}
		existing, err := s.members.GetByStorePhone(ctx, storeID, phone)
		if err != nil {
			return nil, ierr.Internal(err)
		}
		if existing != nil {
			m = existing
		} else {
			var source *string
			if src := strings.TrimSpace(req.Source); src != "" {
				source = &src
			}
			m, err = s.members.Create(ctx, domainmember.CreateInput{
				StoreID: storeID, Name: name, Phone: phone, Source: source,
			})
			if err != nil {
				return nil, ierr.Internal(err)
			}
		}
	}

	// 储值卡、次卡每人每店仅一张：已存在则直接返回
	if product.Type == domainproduct.TypeValue || product.Type == domainproduct.TypeCount {
		existing, err := s.cards.GetByMemberAndType(ctx, storeID, m.ID, product.Type)
		if err != nil {
			return nil, ierr.Internal(err)
		}
		if existing != nil {
			s.log.Info(ctx, "open card idempotent", "store_id", storeID, "member_id", m.ID, "card_id", existing.ID, "type", product.Type)
			return &dto.OpenCardResponse{
				Member:   &dto.MemberBrief{ID: fmt.Sprintf("%d", m.ID), Name: m.Name, Phone: m.Phone},
				Card:     toCardDetail(existing, m),
				LedgerID: "",
			}, nil
		}
	}

	cardIn, err := buildOpenCardInput(storeID, m.ID, userID, product)
	if err != nil {
		return nil, err
	}
	zero := 0
	ledgerIn := domainledger.CreateInput{
		Type:       domainledger.TypeOpen,
		CardType:   product.Type,
		OperatorID: userID,
	}
	switch product.Type {
	case domainproduct.TypeValue:
		bal := cardIn.Balance
		ledgerIn.Amount = &bal
		ledgerIn.BalanceAfter = &bal
	case domainproduct.TypeCount:
		t := 0
		if cardIn.RemainTimes != nil {
			t = *cardIn.RemainTimes
		}
		ledgerIn.Times = &t
		ledgerIn.TimesAfter = &t
	case domainproduct.TypePack:
		ledgerIn.Times = &zero
	}

	card, entry, err := s.txns.OpenCard(ctx, cardIn, ledgerIn)
	if err != nil {
		if (product.Type == domainproduct.TypeValue || product.Type == domainproduct.TypeCount) && ent.IsConstraintError(err) {
			existing, getErr := s.cards.GetByMemberAndType(ctx, storeID, m.ID, product.Type)
			if getErr == nil && existing != nil {
				return &dto.OpenCardResponse{
					Member:   &dto.MemberBrief{ID: fmt.Sprintf("%d", m.ID), Name: m.Name, Phone: m.Phone},
					Card:     toCardDetail(existing, m),
					LedgerID: "",
				}, nil
			}
		}
		return nil, ierr.Internal(err)
	}
	s.log.Info(ctx, "card opened", "store_id", storeID, "member_id", m.ID, "card_id", card.ID)
	return &dto.OpenCardResponse{
		Member:   &dto.MemberBrief{ID: fmt.Sprintf("%d", m.ID), Name: m.Name, Phone: m.Phone},
		Card:     toCardDetail(card, m),
		LedgerID: fmt.Sprintf("%d", entry.ID),
	}, nil
}

func buildOpenCardInput(storeID, memberID, operatorID int64, product *domainproduct.Product) (domaincard.CreateInput, error) {
	now := time.Now().UTC()
	in := domaincard.CreateInput{
		StoreID:      storeID,
		MemberID:     memberID,
		ProductID:    product.ID,
		Type:         product.Type,
		NameSnapshot: product.Name,
		OpenedBy:     operatorID,
		ValidFrom:    &now,
	}
	if product.ValidMonths != nil && *product.ValidMonths > 0 {
		to := now.AddDate(0, *product.ValidMonths, 0)
		in.ValidTo = &to
	}
	switch product.Type {
	case domainproduct.TypeValue:
		in.Balance = product.Price
	case domainproduct.TypeCount:
		if product.Times == nil || *product.Times <= 0 {
			return in, ierr.Validation("次卡卡种缺少次数配置")
		}
		t := *product.Times
		in.RemainTimes = &t
	case domainproduct.TypePack:
		if len(product.Items) == 0 {
			return in, ierr.Validation("套餐卡种缺少项目配置")
		}
		in.Items = make([]domaincard.ItemBalanceInput, 0, len(product.Items))
		for _, it := range product.Items {
			in.Items = append(in.Items, domaincard.ItemBalanceInput{
				ProductItemID: it.ID,
				NameSnapshot:  it.Name,
				RemainTimes:   it.Times,
			})
		}
	default:
		return in, ierr.Validation("卡种类型不正确")
	}
	return in, nil
}

func ParseCustomerID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, ierr.Validation("会员 ID 不正确")
	}
	return id, nil
}

func ParseCardID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, ierr.Validation("卡 ID 不正确")
	}
	return id, nil
}

func ParseLedgerID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, ierr.Validation("流水 ID 不正确")
	}
	return id, nil
}
