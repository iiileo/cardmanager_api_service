package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"card_manager/api_service/internal/api/dto"
	domainledger "card_manager/api_service/internal/domain/ledger"
	domainmember "card_manager/api_service/internal/domain/member"
	domaincard "card_manager/api_service/internal/domain/membercard"
	domainuser "card_manager/api_service/internal/domain/user"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
)

type LedgerService interface {
	List(ctx context.Context, userID, storeID int64, typ, from, to string, page, pageSize int) (*dto.LedgerListResponse, error)
	Get(ctx context.Context, userID, storeID, ledgerID int64) (*dto.LedgerEntryResponse, error)
}

type ledgerService struct {
	ledgers  domainledger.Repository
	members  domainmember.Repository
	cards    domaincard.Repository
	users    domainuser.Repository
	storeSvc StoreService
	log      *logger.Logger
}

func NewLedgerService(
	ledgers domainledger.Repository,
	members domainmember.Repository,
	cards domaincard.Repository,
	users domainuser.Repository,
	storeSvc StoreService,
	log *logger.Logger,
) LedgerService {
	return &ledgerService{
		ledgers: ledgers, members: members, cards: cards, users: users, storeSvc: storeSvc, log: log,
	}
}

func (s *ledgerService) List(ctx context.Context, userID, storeID int64, typ, from, to string, page, pageSize int) (*dto.LedgerListResponse, error) {
	if _, err := s.storeSvc.RequireActiveMember(ctx, userID, storeID); err != nil {
		return nil, err
	}
	typ = strings.TrimSpace(typ)
	if typ != "" && !validLedgerType(typ) {
		return nil, ierr.Validation("流水类型不正确")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	f := domainledger.ListFilter{
		StoreID: storeID,
		Type:    typ,
		Limit:   pageSize,
		Offset:  (page - 1) * pageSize,
	}
	if from != "" {
		t, err := parseDayStart(from)
		if err != nil {
			return nil, ierr.Validation("起始日期不正确")
		}
		f.From = &t
	}
	if to != "" {
		t, err := parseDayEnd(to)
		if err != nil {
			return nil, ierr.Validation("结束日期不正确")
		}
		f.To = &t
	}
	list, total, err := s.ledgers.List(ctx, f)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	out := make([]*dto.LedgerEntryResponse, 0, len(list))
	for _, e := range list {
		out = append(out, s.toLedgerDTO(ctx, e))
	}
	return &dto.LedgerListResponse{List: out, Total: total}, nil
}

func (s *ledgerService) Get(ctx context.Context, userID, storeID, ledgerID int64) (*dto.LedgerEntryResponse, error) {
	if _, err := s.storeSvc.RequireActiveMember(ctx, userID, storeID); err != nil {
		return nil, err
	}
	e, err := s.ledgers.GetByID(ctx, ledgerID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	if e == nil || e.StoreID != storeID {
		return nil, ierr.NotFound("流水不存在")
	}
	return s.toLedgerDTO(ctx, e), nil
}

func (s *ledgerService) toLedgerDTO(ctx context.Context, e *domainledger.Entry) *dto.LedgerEntryResponse {
	resp := &dto.LedgerEntryResponse{
		ID:           fmt.Sprintf("%d", e.ID),
		Type:         e.Type,
		Amount:       e.Amount,
		Times:        e.Times,
		ItemName:     e.ItemName,
		BalanceAfter: e.BalanceAfter,
		TimesAfter:   e.TimesAfter,
		Remark:       e.Remark,
		CreatedAt:    e.CreatedAt.UTC().Format(time.RFC3339),
	}
	if m, err := s.members.GetByID(ctx, e.MemberID); err == nil && m != nil {
		resp.Member = &dto.MemberBrief{ID: fmt.Sprintf("%d", m.ID), Name: m.Name, Phone: m.Phone}
	}
	if c, err := s.cards.GetByID(ctx, e.CardID); err == nil && c != nil {
		resp.Card = toCardBrief(c)
	}
	if u, err := s.users.GetByID(ctx, e.OperatorID); err == nil && u != nil {
		resp.Operator = &dto.OperatorBrief{ID: fmt.Sprintf("%d", u.ID), Nickname: u.Nickname}
	}
	if len(e.Items) > 0 {
		resp.Items = make([]*dto.LedgerItemResponse, 0, len(e.Items))
		for _, it := range e.Items {
			resp.Items = append(resp.Items, &dto.LedgerItemResponse{
				ID:         fmt.Sprintf("%d", it.ID),
				ItemID:     fmt.Sprintf("%d", it.ItemBalanceID),
				Name:       it.NameSnapshot,
				Times:      it.Times,
				TimesAfter: it.TimesAfter,
			})
		}
	}
	return resp
}

func validLedgerType(typ string) bool {
	switch typ {
	case domainledger.TypeOpen, domainledger.TypeRecharge,
		domainledger.TypeConsumeValue, domainledger.TypeConsumeCount, domainledger.TypeConsumePack:
		return true
	default:
		return false
	}
}

func parseDayStart(s string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(s), time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func parseDayEnd(s string) (time.Time, error) {
	t, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(s), time.Local)
	if err != nil {
		return time.Time{}, err
	}
	return t.Add(24*time.Hour - time.Nanosecond), nil
}