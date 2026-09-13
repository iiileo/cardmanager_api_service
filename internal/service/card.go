package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"card_manager/api_service/internal/api/dto"
	domainmember "card_manager/api_service/internal/domain/member"
	domaincard "card_manager/api_service/internal/domain/membercard"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/repository"
)

type CardService interface {
	Get(ctx context.Context, userID, storeID, cardID int64) (*dto.MemberCardDetail, error)
	ListByMember(ctx context.Context, userID, storeID, memberID int64) (*dto.MemberCardListResponse, error)
	Recharge(ctx context.Context, userID, storeID, cardID int64, req dto.RechargeRequest) (*dto.TxnResultResponse, error)
	Consume(ctx context.Context, userID, storeID, cardID int64, req dto.ConsumeRequest) (*dto.TxnResultResponse, error)
}

type cardService struct {
	cards    domaincard.Repository
	members  domainmember.Repository
	txns     repository.CardTxnRepository
	storeSvc StoreService
	log      *logger.Logger
}

func NewCardService(
	cards domaincard.Repository,
	members domainmember.Repository,
	txns repository.CardTxnRepository,
	storeSvc StoreService,
	log *logger.Logger,
) CardService {
	return &cardService{cards: cards, members: members, txns: txns, storeSvc: storeSvc, log: log}
}

func (s *cardService) Get(ctx context.Context, userID, storeID, cardID int64) (*dto.MemberCardDetail, error) {
	if _, err := s.storeSvc.RequireActiveMember(ctx, userID, storeID); err != nil {
		return nil, err
	}
	card, err := s.cards.GetByID(ctx, cardID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	if card == nil || card.StoreID != storeID {
		return nil, ierr.NotFound("卡不存在")
	}
	m, err := s.members.GetByID(ctx, card.MemberID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	return toCardDetail(card, m), nil
}

func (s *cardService) ListByMember(ctx context.Context, userID, storeID, memberID int64) (*dto.MemberCardListResponse, error) {
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
	list, err := s.cards.ListByMember(ctx, storeID, memberID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	out := make([]*dto.MemberCardDetail, 0, len(list))
	for _, c := range list {
		out = append(out, toCardDetail(c, m))
	}
	return &dto.MemberCardListResponse{List: out}, nil
}

func (s *cardService) Recharge(ctx context.Context, userID, storeID, cardID int64, req dto.RechargeRequest) (*dto.TxnResultResponse, error) {
	if _, err := s.storeSvc.RequireActiveMember(ctx, userID, storeID); err != nil {
		return nil, err
	}
	if req.Amount <= 0 {
		return nil, ierr.Validation("充值金额须大于 0")
	}
	card, err := s.requireStoreCard(ctx, storeID, cardID)
	if err != nil {
		return nil, err
	}
	if card.Type != domaincard.TypeValue {
		return nil, ierr.Validation("仅储值卡可充值")
	}
	if err := ensureCardUsable(card); err != nil {
		return nil, err
	}
	updated, entry, err := s.txns.Recharge(ctx, cardID, req.Amount, userID, nil)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	bal := updated.Balance
	return &dto.TxnResultResponse{
		CardID: fmt.Sprintf("%d", updated.ID), BalanceAfter: &bal, LedgerID: fmt.Sprintf("%d", entry.ID),
	}, nil
}

func (s *cardService) Consume(ctx context.Context, userID, storeID, cardID int64, req dto.ConsumeRequest) (*dto.TxnResultResponse, error) {
	if _, err := s.storeSvc.RequireActiveMember(ctx, userID, storeID); err != nil {
		return nil, err
	}
	card, err := s.requireStoreCard(ctx, storeID, cardID)
	if err != nil {
		return nil, err
	}
	if err := ensureCardUsable(card); err != nil {
		return nil, err
	}

	var remark *string
	if r := strings.TrimSpace(req.Remark); r != "" {
		remark = &r
	}

	switch card.Type {
	case domaincard.TypeValue:
		if req.Amount == nil || *req.Amount <= 0 {
			return nil, ierr.Validation("请填写扣款金额")
		}
		amount := *req.Amount
		if amount > card.Balance {
			return nil, ierr.Conflict("余额不足").WithCode(40901).WithData(map[string]int{
				"balance": card.Balance, "need": amount,
			})
		}
		updated, entry, err := s.txns.ConsumeValue(ctx, cardID, amount, userID, remark)
		if repository.IsInsufficient(err) {
			return nil, ierr.Conflict("余额不足").WithCode(40901)
		}
		if err != nil {
			return nil, ierr.Internal(err)
		}
		bal := updated.Balance
		return &dto.TxnResultResponse{
			CardID: fmt.Sprintf("%d", updated.ID), BalanceAfter: &bal, LedgerID: fmt.Sprintf("%d", entry.ID),
		}, nil

	case domaincard.TypeCount:
		times := 1
		if req.Times != nil {
			times = *req.Times
		}
		if times <= 0 {
			return nil, ierr.Validation("扣次须大于 0")
		}
		remain := 0
		if card.RemainTimes != nil {
			remain = *card.RemainTimes
		}
		if times > remain {
			return nil, ierr.Conflict("剩余次数不足").WithCode(40901).WithData(map[string]int{
				"remain_times": remain, "need": times,
			})
		}
		updated, entry, err := s.txns.ConsumeCount(ctx, cardID, times, userID, remark)
		if repository.IsInsufficient(err) {
			return nil, ierr.Conflict("剩余次数不足").WithCode(40901)
		}
		if err != nil {
			return nil, ierr.Internal(err)
		}
		return &dto.TxnResultResponse{
			CardID: fmt.Sprintf("%d", updated.ID), TimesAfter: updated.RemainTimes, LedgerID: fmt.Sprintf("%d", entry.ID),
		}, nil

	case domaincard.TypePack:
		deducts, err := normalizePackDeducts(card, req)
		if err != nil {
			return nil, err
		}
		updated, entry, err := s.txns.ConsumePack(ctx, cardID, deducts, userID, remark)
		if repository.IsInsufficient(err) {
			return nil, ierr.Conflict("套餐项目剩余次数不足").WithCode(40901)
		}
		if err != nil {
			if strings.Contains(err.Error(), "item not found") {
				return nil, ierr.NotFound("套餐项目不存在")
			}
			return nil, ierr.Internal(err)
		}
		items := make([]*dto.PackItemBalanceResponse, 0, len(updated.Items))
		for _, it := range updated.Items {
			items = append(items, &dto.PackItemBalanceResponse{
				ID: fmt.Sprintf("%d", it.ID), Name: it.NameSnapshot, RemainTimes: it.RemainTimes,
			})
		}
		return &dto.TxnResultResponse{
			CardID: fmt.Sprintf("%d", updated.ID), LedgerID: fmt.Sprintf("%d", entry.ID), Items: items,
		}, nil
	default:
		return nil, ierr.Validation("卡类型不正确")
	}
}

func normalizePackDeducts(card *domaincard.Card, req dto.ConsumeRequest) ([]domaincard.PackDeductItem, error) {
	merged := map[int64]int{}

	add := func(rawID string, times int) error {
		if times <= 0 {
			times = 1
		}
		id, err := strconv.ParseInt(strings.TrimSpace(rawID), 10, 64)
		if err != nil || id <= 0 {
			return ierr.Validation("套餐项目 ID 不正确")
		}
		found := false
		for _, it := range card.Items {
			if it.ID == id {
				found = true
				if times > it.RemainTimes {
					return ierr.Conflict("套餐项目剩余次数不足").WithCode(40901).WithData(map[string]any{
						"item_id": rawID, "remain_times": it.RemainTimes, "need": times,
					})
				}
				break
			}
		}
		if !found {
			return ierr.NotFound("套餐项目不存在")
		}
		merged[id] += times
		return nil
	}

	if len(req.Items) > 0 {
		for _, it := range req.Items {
			if err := add(it.ItemID, it.Times); err != nil {
				return nil, err
			}
		}
	} else if req.ItemID != nil && strings.TrimSpace(*req.ItemID) != "" {
		times := 1
		if req.Times != nil {
			times = *req.Times
		}
		if err := add(*req.ItemID, times); err != nil {
			return nil, err
		}
	} else {
		return nil, ierr.Validation("请选择至少一个套餐项目")
	}

	out := make([]domaincard.PackDeductItem, 0, len(merged))
	for id, times := range merged {
		for _, it := range card.Items {
			if it.ID == id && times > it.RemainTimes {
				return nil, ierr.Conflict("套餐项目剩余次数不足").WithCode(40901)
			}
		}
		out = append(out, domaincard.PackDeductItem{ItemBalanceID: id, Times: times})
	}
	return out, nil
}

func (s *cardService) requireStoreCard(ctx context.Context, storeID, cardID int64) (*domaincard.Card, error) {
	card, err := s.cards.GetByID(ctx, cardID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	if card == nil || card.StoreID != storeID {
		return nil, ierr.NotFound("卡不存在")
	}
	return card, nil
}

func ensureCardUsable(card *domaincard.Card) error {
	if card.Status != domaincard.StatusActive {
		return ierr.Conflict("该卡不可用")
	}
	if card.ValidTo != nil && time.Now().UTC().After(*card.ValidTo) {
		return ierr.Conflict("卡已过期")
	}
	return nil
}

func toCardBrief(c *domaincard.Card) *dto.MemberCardBrief {
	b := &dto.MemberCardBrief{
		ID: fmt.Sprintf("%d", c.ID), Type: c.Type, Name: c.NameSnapshot, Status: c.Status,
	}
	switch c.Type {
	case domaincard.TypeValue:
		bal := c.Balance
		b.Balance = &bal
	case domaincard.TypeCount:
		b.RemainTimes = c.RemainTimes
	}
	if c.ValidTo != nil {
		s := c.ValidTo.Format("2006-01-02")
		b.ValidTo = &s
	}
	return b
}

func toCardDetail(c *domaincard.Card, m *domainmember.Member) *dto.MemberCardDetail {
	d := &dto.MemberCardDetail{
		ID: fmt.Sprintf("%d", c.ID),
		MemberID: fmt.Sprintf("%d", c.MemberID),
		Type: c.Type,
		Name: c.NameSnapshot,
		NameSnapshot: c.NameSnapshot,
		Status: c.Status,
	}
	switch c.Type {
	case domaincard.TypeValue:
		bal := c.Balance
		d.Balance = &bal
	case domaincard.TypeCount:
		d.RemainTimes = c.RemainTimes
	case domaincard.TypePack:
		d.Items = make([]*dto.PackItemBalanceResponse, 0, len(c.Items))
		for _, it := range c.Items {
			d.Items = append(d.Items, &dto.PackItemBalanceResponse{
				ID: fmt.Sprintf("%d", it.ID), Name: it.NameSnapshot, RemainTimes: it.RemainTimes,
			})
		}
	}
	if c.ValidTo != nil {
		s := c.ValidTo.Format("2006-01-02")
		d.ValidTo = &s
	}
	if m != nil {
		d.Member = &dto.MemberBrief{ID: fmt.Sprintf("%d", m.ID), Name: m.Name, Phone: m.Phone}
	}
	return d
}
