package service

import (
	"context"
	"testing"

	"card_manager/api_service/internal/api/dto"
	"card_manager/api_service/internal/config"
	domainledger "card_manager/api_service/internal/domain/ledger"
	domainmember "card_manager/api_service/internal/domain/member"
	domaincard "card_manager/api_service/internal/domain/membercard"
	domainstore "card_manager/api_service/internal/domain/store"
	domainmemberstaff "card_manager/api_service/internal/domain/storemember"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/repository"
)

type memCardRepo struct {
	byID map[int64]*domaincard.Card
}

func (m *memCardRepo) GetByID(_ context.Context, id int64) (*domaincard.Card, error) {
	return m.byID[id], nil
}
func (m *memCardRepo) ListByMember(_ context.Context, storeID, memberID int64) ([]*domaincard.Card, error) {
	var out []*domaincard.Card
	for _, c := range m.byID {
		if c.StoreID == storeID && c.MemberID == memberID {
			out = append(out, c)
		}
	}
	return out, nil
}

type memCustomerRepo struct {
	byID map[int64]*domainmember.Member
}

func (m *memCustomerRepo) GetByID(_ context.Context, id int64) (*domainmember.Member, error) {
	return m.byID[id], nil
}
func (m *memCustomerRepo) GetByStorePhone(context.Context, int64, string) (*domainmember.Member, error) {
	return nil, nil
}
func (m *memCustomerRepo) Create(context.Context, domainmember.CreateInput) (*domainmember.Member, error) {
	return nil, nil
}
func (m *memCustomerRepo) ListByStore(context.Context, int64, string, int, int) ([]*domainmember.Member, int, error) {
	return nil, 0, nil
}

type memTxnRepo struct {
	cards *memCardRepo
	last  *domainledger.Entry
	seq   int64
}

func (t *memTxnRepo) Recharge(context.Context, int64, int, int64, *string) (*domaincard.Card, *domainledger.Entry, error) {
	return nil, nil, nil
}
func (t *memTxnRepo) ConsumeValue(context.Context, int64, int, int64, *string) (*domaincard.Card, *domainledger.Entry, error) {
	return nil, nil, nil
}
func (t *memTxnRepo) ConsumeCount(context.Context, int64, int, int64, *string) (*domaincard.Card, *domainledger.Entry, error) {
	return nil, nil, nil
}
func (t *memTxnRepo) OpenCard(context.Context, domaincard.CreateInput, domainledger.CreateInput) (*domaincard.Card, *domainledger.Entry, error) {
	return nil, nil, nil
}
func (t *memTxnRepo) ConsumePack(_ context.Context, cardID int64, items []domaincard.PackDeductItem, operatorID int64, remark *string) (*domaincard.Card, *domainledger.Entry, error) {
	card := t.cards.byID[cardID]
	if card == nil {
		return nil, nil, repository.ErrInsufficient
	}
	byID := map[int64]*domaincard.ItemBalance{}
	for _, it := range card.Items {
		byID[it.ID] = it
	}
	ledgerItems := make([]*domainledger.EntryItem, 0, len(items))
	names := make([]string, 0, len(items))
	total := 0
	for _, d := range items {
		it, ok := byID[d.ItemBalanceID]
		if !ok || it.RemainTimes < d.Times {
			return nil, nil, repository.ErrInsufficient
		}
		it.RemainTimes -= d.Times
		total += d.Times
		names = append(names, it.NameSnapshot)
		ledgerItems = append(ledgerItems, &domainledger.EntryItem{
			ItemBalanceID: it.ID, ProductItemID: it.ProductItemID,
			NameSnapshot: it.NameSnapshot, Times: d.Times, TimesAfter: it.RemainTimes,
		})
	}
	t.seq++
	neg := -total
	name := names[0]
	if len(names) > 1 {
		name = names[0] + "、" + names[1]
	}
	entry := &domainledger.Entry{
		ID: t.seq, StoreID: card.StoreID, MemberID: card.MemberID, CardID: card.ID,
		Type: domainledger.TypeConsumePack, Times: &neg, ItemName: &name,
		Remark: remark, OperatorID: operatorID, Items: ledgerItems,
	}
	t.last = entry
	return card, entry, nil
}

type okStoreSvc struct{}

func (okStoreSvc) ListMine(context.Context, int64) (*dto.StoreListResponse, error) {
	return nil, nil
}
func (okStoreSvc) Create(context.Context, int64, dto.CreateStoreRequest) (*dto.StoreDetailResponse, error) {
	return nil, nil
}
func (okStoreSvc) Get(context.Context, int64, int64) (*dto.StoreDetailResponse, error) {
	return nil, nil
}
func (okStoreSvc) Update(context.Context, int64, int64, dto.UpdateStoreRequest) (*dto.StoreDetailResponse, error) {
	return nil, nil
}
func (okStoreSvc) PreviewInvite(context.Context, string) (*dto.InvitePreviewResponse, error) {
	return nil, nil
}
func (okStoreSvc) Join(context.Context, int64, dto.JoinStoreRequest) (*dto.StoreListItem, error) {
	return nil, nil
}
func (okStoreSvc) GetInviteCode(context.Context, int64, int64) (*dto.InviteCodeResponse, error) {
	return nil, nil
}
func (okStoreSvc) RefreshInviteCode(context.Context, int64, int64) (*dto.InviteCodeResponse, error) {
	return nil, nil
}
func (okStoreSvc) ListStaff(context.Context, int64, int64) (*dto.StaffListResponse, error) {
	return nil, nil
}
func (okStoreSvc) ListApplications(context.Context, int64, int64) (*dto.StaffListResponse, error) {
	return nil, nil
}
func (okStoreSvc) Approve(context.Context, int64, int64, int64) error { return nil }
func (okStoreSvc) Reject(context.Context, int64, int64, int64) error  { return nil }
func (okStoreSvc) RequireActiveMember(context.Context, int64, int64) (*domainmemberstaff.Member, error) {
	return &domainmemberstaff.Member{Role: domainstore.RoleOwner, Status: domainstore.MemberActive}, nil
}
func (okStoreSvc) RequireOwner(context.Context, int64, int64) (*domainmemberstaff.Member, error) {
	return &domainmemberstaff.Member{Role: domainstore.RoleOwner, Status: domainstore.MemberActive}, nil
}

func TestCardService_ConsumePackMultiItems(t *testing.T) {
	cards := &memCardRepo{byID: map[int64]*domaincard.Card{
		10: {
			ID: 10, StoreID: 1, MemberID: 2, Type: domaincard.TypePack,
			NameSnapshot: "洗剪吹", Status: domaincard.StatusActive,
			Items: []*domaincard.ItemBalance{
				{ID: 101, CardID: 10, ProductItemID: 1, NameSnapshot: "洗", RemainTimes: 3},
				{ID: 102, CardID: 10, ProductItemID: 2, NameSnapshot: "剪", RemainTimes: 4},
				{ID: 103, CardID: 10, ProductItemID: 3, NameSnapshot: "护理", RemainTimes: 1},
			},
		},
	}}
	members := &memCustomerRepo{byID: map[int64]*domainmember.Member{
		2: {ID: 2, StoreID: 1, Name: "李明", Phone: "13900001111"},
	}}
	txns := &memTxnRepo{cards: cards}
	svc := NewCardService(cards, members, txns, okStoreSvc{}, logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}}))

	resp, err := svc.Consume(context.Background(), 9, 1, 10, dto.ConsumeRequest{
		Items: []dto.ConsumePackItemRequest{
			{ItemID: "101", Times: 1},
			{ItemID: "102", Times: 2},
		},
		Remark: "一次多项",
	})
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if resp.LedgerID == "" {
		t.Fatal("expected ledger id")
	}
	if cards.byID[10].Items[0].RemainTimes != 2 || cards.byID[10].Items[1].RemainTimes != 2 {
		t.Fatalf("unexpected remain: %+v", cards.byID[10].Items)
	}
	if txns.last == nil || len(txns.last.Items) != 2 {
		t.Fatalf("expected 2 ledger items, got %+v", txns.last)
	}
	if *txns.last.Times != -3 {
		t.Fatalf("expected times -3, got %v", *txns.last.Times)
	}
}

func TestCardService_ConsumePackInsufficient(t *testing.T) {
	cards := &memCardRepo{byID: map[int64]*domaincard.Card{
		10: {
			ID: 10, StoreID: 1, MemberID: 2, Type: domaincard.TypePack,
			NameSnapshot: "套餐", Status: domaincard.StatusActive,
			Items: []*domaincard.ItemBalance{
				{ID: 101, CardID: 10, NameSnapshot: "洗", RemainTimes: 1},
			},
		},
	}}
	svc := NewCardService(cards, &memCustomerRepo{byID: map[int64]*domainmember.Member{}}, &memTxnRepo{cards: cards}, okStoreSvc{}, logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}}))
	_, err := svc.Consume(context.Background(), 9, 1, 10, dto.ConsumeRequest{
		Items: []dto.ConsumePackItemRequest{{ItemID: "101", Times: 2}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	app, ok := ierr.AsAppError(err)
	if !ok || app.Code() != 40901 {
		t.Fatalf("expected 40901, got %v", err)
	}
}
