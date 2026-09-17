package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"card_manager/api_service/internal/api/dto"
	domainledger "card_manager/api_service/internal/domain/ledger"
	domaincard "card_manager/api_service/internal/domain/membercard"
	domainmember "card_manager/api_service/internal/domain/member"
	domainuser "card_manager/api_service/internal/domain/user"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
)

type LedgerListQuery struct {
	Kind         string // 空=不限；txn=充值+消费（不含开卡）
	Type         string // 精确类型：recharge|consume_value|consume_count|consume_pack|open
	CardType     string // value|count|pack
	MemberID     string
	CardID       string
	From         string
	To           string
	Page         int
	PageSize     int
	IncludeStats bool // 为 true 时在列表响应中附带 stats，减少二次请求
}

type LedgerService interface {
	List(ctx context.Context, userID, storeID int64, q LedgerListQuery) (*dto.LedgerListResponse, error)
	Get(ctx context.Context, userID, storeID, ledgerID int64) (*dto.LedgerEntryResponse, error)

	ListRecharges(ctx context.Context, userID, storeID int64, q LedgerListQuery) (*dto.LedgerListResponse, error)
	ListConsumes(ctx context.Context, userID, storeID int64, q LedgerListQuery) (*dto.LedgerListResponse, error)
	GetRecharge(ctx context.Context, userID, storeID, id int64) (*dto.LedgerEntryResponse, error)
	GetConsume(ctx context.Context, userID, storeID, id int64) (*dto.LedgerEntryResponse, error)
	StatsRecharges(ctx context.Context, userID, storeID int64, q LedgerListQuery) (*dto.RecordStatsResponse, error)
	StatsConsumes(ctx context.Context, userID, storeID int64, q LedgerListQuery) (*dto.RecordStatsResponse, error)
	StatsTxns(ctx context.Context, userID, storeID int64, q LedgerListQuery) (*dto.RecordStatsResponse, error)
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

func (s *ledgerService) List(ctx context.Context, userID, storeID int64, q LedgerListQuery) (*dto.LedgerListResponse, error) {
	types, err := resolveLedgerTypes(q, nil)
	if err != nil {
		return nil, err
	}
	return s.list(ctx, userID, storeID, q, types)
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

func (s *ledgerService) ListRecharges(ctx context.Context, userID, storeID int64, q LedgerListQuery) (*dto.LedgerListResponse, error) {
	return s.list(ctx, userID, storeID, q, []string{domainledger.TypeRecharge})
}

func (s *ledgerService) ListConsumes(ctx context.Context, userID, storeID int64, q LedgerListQuery) (*dto.LedgerListResponse, error) {
	types, err := resolveLedgerTypes(q, domainledger.ConsumeTypes())
	if err != nil {
		return nil, err
	}
	return s.list(ctx, userID, storeID, q, types)
}

func (s *ledgerService) GetRecharge(ctx context.Context, userID, storeID, id int64) (*dto.LedgerEntryResponse, error) {
	resp, err := s.Get(ctx, userID, storeID, id)
	if err != nil {
		return nil, err
	}
	if resp.Type != domainledger.TypeRecharge {
		return nil, ierr.NotFound("充值记录不存在")
	}
	return resp, nil
}

func (s *ledgerService) GetConsume(ctx context.Context, userID, storeID, id int64) (*dto.LedgerEntryResponse, error) {
	resp, err := s.Get(ctx, userID, storeID, id)
	if err != nil {
		return nil, err
	}
	if !domainledger.IsConsumeType(resp.Type) {
		return nil, ierr.NotFound("消费记录不存在")
	}
	return resp, nil
}

func (s *ledgerService) StatsRecharges(ctx context.Context, userID, storeID int64, q LedgerListQuery) (*dto.RecordStatsResponse, error) {
	return s.stats(ctx, userID, storeID, q, []string{domainledger.TypeRecharge})
}

func (s *ledgerService) StatsConsumes(ctx context.Context, userID, storeID int64, q LedgerListQuery) (*dto.RecordStatsResponse, error) {
	types, err := resolveLedgerTypes(q, domainledger.ConsumeTypes())
	if err != nil {
		return nil, err
	}
	return s.stats(ctx, userID, storeID, q, types)
}

func (s *ledgerService) StatsTxns(ctx context.Context, userID, storeID int64, q LedgerListQuery) (*dto.RecordStatsResponse, error) {
	types, err := resolveLedgerTypes(LedgerListQuery{
		Kind: "txn", Type: q.Type, CardType: q.CardType,
		MemberID: q.MemberID, CardID: q.CardID, From: q.From, To: q.To,
	}, nil)
	if err != nil {
		return nil, err
	}
	return s.stats(ctx, userID, storeID, q, types)
}

func (s *ledgerService) list(ctx context.Context, userID, storeID int64, q LedgerListQuery, fixedTypes []string) (*dto.LedgerListResponse, error) {
	f, err := s.buildListFilter(storeID, q, fixedTypes)
	if err != nil {
		return nil, err
	}
	list, total, err := s.ledgers.List(ctx, f)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	resp := &dto.LedgerListResponse{List: s.toLedgerDTOs(list), Total: total}
	if q.IncludeStats {
		stats, err := s.stats(ctx, userID, storeID, q, fixedTypes)
		if err != nil {
			return nil, err
		}
		resp.Stats = stats
	}
	return resp, nil
}

func (s *ledgerService) stats(ctx context.Context, userID, storeID int64, q LedgerListQuery, types []string) (*dto.RecordStatsResponse, error) {
	sf, err := s.buildSummaryFilter(storeID, q, types)
	if err != nil {
		return nil, err
	}
	sum, err := s.ledgers.Summarize(ctx, sf)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	resp := &dto.RecordStatsResponse{
		TotalCount:  sum.TotalCount,
		TotalAmount: sum.TotalAmount,
		TotalTimes:  sum.TotalTimes,
		ByType:      make([]*dto.RecordTypeStat, 0, len(sum.ByType)),
	}
	for _, b := range sum.ByType {
		resp.ByType = append(resp.ByType, &dto.RecordTypeStat{
			Type: b.Type, CardType: b.CardType, Count: b.Count, Amount: b.Amount, Times: b.Times,
		})
	}
	if len(sum.ByPackItem) > 0 {
		resp.ByPackItem = make([]*dto.PackItemStat, 0, len(sum.ByPackItem))
		for _, b := range sum.ByPackItem {
			resp.ByPackItem = append(resp.ByPackItem, &dto.PackItemStat{
				ProductItemID: fmt.Sprintf("%d", b.ProductItemID),
				Name:          b.Name,
				Times:         b.Times,
				Count:         b.Count,
			})
		}
	}
	return resp, nil
}

func (s *ledgerService) buildListFilter(storeID int64, q LedgerListQuery, fixedTypes []string) (domainledger.ListFilter, error) {
	page, pageSize := q.Page, q.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	f := domainledger.ListFilter{
		StoreID:  storeID,
		Types:    fixedTypes,
		CardType: strings.TrimSpace(q.CardType),
		Limit:    pageSize,
		Offset:   (page - 1) * pageSize,
	}
	if f.CardType != "" && f.CardType != domaincard.TypeValue && f.CardType != domaincard.TypeCount && f.CardType != domaincard.TypePack {
		return f, ierr.Validation("卡类型不正确")
	}
	if q.MemberID != "" {
		id, err := ParseCustomerID(q.MemberID)
		if err != nil {
			return f, err
		}
		f.MemberID = &id
	}
	if q.CardID != "" {
		id, err := ParseCardID(q.CardID)
		if err != nil {
			return f, err
		}
		f.CardID = &id
	}
	if q.From != "" {
		t, err := parseDayStart(q.From)
		if err != nil {
			return f, ierr.Validation("起始日期不正确")
		}
		f.From = &t
	}
	if q.To != "" {
		t, err := parseDayEnd(q.To)
		if err != nil {
			return f, ierr.Validation("结束日期不正确")
		}
		f.To = &t
	}
	return f, nil
}

// resolveLedgerTypes 解析 kind/type 到具体 type 列表。
// base 非空时表示业务入口已限定范围（如消费接口默认 ConsumeTypes）。
func resolveLedgerTypes(q LedgerListQuery, base []string) ([]string, error) {
	kind := strings.TrimSpace(q.Kind)
	typ := strings.TrimSpace(q.Type)

	var allowed []string
	switch kind {
	case "":
		allowed = base
	case "txn":
		allowed = append([]string{domainledger.TypeRecharge}, domainledger.ConsumeTypes()...)
	case "recharge":
		allowed = []string{domainledger.TypeRecharge}
	case "consume":
		allowed = domainledger.ConsumeTypes()
	default:
		return nil, ierr.Validation("kind 不正确，可选 txn|recharge|consume")
	}

	if typ == "" {
		return allowed, nil
	}
	if !validLedgerType(typ) {
		return nil, ierr.Validation("流水类型不正确")
	}
	if len(allowed) == 0 {
		return []string{typ}, nil
	}
	for _, a := range allowed {
		if a == typ {
			return []string{typ}, nil
		}
	}
	return nil, ierr.Validation("type 与 kind 不匹配")
}

func (s *ledgerService) buildSummaryFilter(storeID int64, q LedgerListQuery, types []string) (domainledger.SummaryFilter, error) {
	lf, err := s.buildListFilter(storeID, q, types)
	if err != nil {
		return domainledger.SummaryFilter{}, err
	}
	return domainledger.SummaryFilter{
		StoreID:  lf.StoreID,
		Types:    lf.Types,
		CardType: lf.CardType,
		MemberID: lf.MemberID,
		CardID:   lf.CardID,
		From:     lf.From,
		To:       lf.To,
	}, nil
}

func (s *ledgerService) toLedgerDTO(ctx context.Context, e *domainledger.Entry) *dto.LedgerEntryResponse {
	list := s.toLedgerDTOs([]*domainledger.Entry{e})
	if len(list) == 0 {
		return nil
	}
	return list[0]
}

func (s *ledgerService) toLedgerDTOs(list []*domainledger.Entry) []*dto.LedgerEntryResponse {
	if len(list) == 0 {
		return nil
	}
	out := make([]*dto.LedgerEntryResponse, 0, len(list))
	for _, e := range list {
		out = append(out, s.entryToDTO(e))
	}
	return out
}

func (s *ledgerService) entryToDTO(e *domainledger.Entry) *dto.LedgerEntryResponse {
	resp := &dto.LedgerEntryResponse{
		ID:           fmt.Sprintf("%d", e.ID),
		Type:         e.Type,
		CardType:     e.CardType,
		Amount:       e.Amount,
		Times:        e.Times,
		ItemName:     e.ItemName,
		BalanceAfter: e.BalanceAfter,
		TimesAfter:   e.TimesAfter,
		Remark:       e.Remark,
		CreatedAt:    e.CreatedAt.UTC().Format(time.RFC3339),
	}
	if m := e.Member; m != nil {
		resp.Member = &dto.MemberBrief{ID: fmt.Sprintf("%d", m.ID), Name: m.Name, Phone: m.Phone}
	}
	if c := e.Card; c != nil {
		resp.Card = toCardBrief(c)
	}
	if u := e.Operator; u != nil {
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
