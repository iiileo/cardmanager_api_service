package service

import (
	"context"
	"testing"
	"time"

	"card_manager/api_service/internal/config"
	domainledger "card_manager/api_service/internal/domain/ledger"
	domainmember "card_manager/api_service/internal/domain/member"
	domaincard "card_manager/api_service/internal/domain/membercard"
	domainuser "card_manager/api_service/internal/domain/user"
	"card_manager/api_service/internal/logger"
)

type memLedgerRepo struct {
	entries []*domainledger.Entry
}

func (m *memLedgerRepo) Create(context.Context, domainledger.CreateInput) (*domainledger.Entry, error) {
	return nil, nil
}
func (m *memLedgerRepo) GetByID(_ context.Context, id int64) (*domainledger.Entry, error) {
	for _, e := range m.entries {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, nil
}
func (m *memLedgerRepo) List(_ context.Context, f domainledger.ListFilter) ([]*domainledger.Entry, int, error) {
	var out []*domainledger.Entry
	for _, e := range m.entries {
		if e.StoreID != f.StoreID {
			continue
		}
		types := f.Types
		if len(types) == 0 && f.Type != "" {
			types = []string{f.Type}
		}
		if len(types) > 0 {
			ok := false
			for _, t := range types {
				if e.Type == t {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}
		if f.CardType != "" && e.CardType != f.CardType {
			continue
		}
		if f.MemberID != nil && e.MemberID != *f.MemberID {
			continue
		}
		if f.CardID != nil && e.CardID != *f.CardID {
			continue
		}
		out = append(out, e)
	}
	total := len(out)
	start := f.Offset
	if start > total {
		start = total
	}
	end := start + f.Limit
	if f.Limit <= 0 || end > total {
		end = total
	}
	return out[start:end], total, nil
}
func (m *memLedgerRepo) Summarize(_ context.Context, f domainledger.SummaryFilter) (*domainledger.Summary, error) {
	sum := &domainledger.Summary{}
	byKey := map[string]*domainledger.TypeBucket{}
	pack := map[int64]*domainledger.PackItemBucket{}
	for _, e := range m.entries {
		if e.StoreID != f.StoreID {
			continue
		}
		if len(f.Types) > 0 {
			ok := false
			for _, t := range f.Types {
				if e.Type == t {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}
		if f.CardType != "" && e.CardType != f.CardType {
			continue
		}
		key := e.Type + "|" + e.CardType
		b := byKey[key]
		if b == nil {
			b = &domainledger.TypeBucket{Type: e.Type, CardType: e.CardType}
			byKey[key] = b
		}
		b.Count++
		if e.Amount != nil {
			a := *e.Amount
			if a < 0 {
				a = -a
			}
			b.Amount += a
		}
		if e.Times != nil {
			t := *e.Times
			if t < 0 {
				t = -t
			}
			b.Times += t
		}
		if e.Type == domainledger.TypeConsumePack {
			for _, it := range e.Items {
				p := pack[it.ProductItemID]
				if p == nil {
					p = &domainledger.PackItemBucket{ProductItemID: it.ProductItemID, Name: it.NameSnapshot}
					pack[it.ProductItemID] = p
				}
				p.Times += it.Times
				p.Count++
			}
		}
	}
	for _, b := range byKey {
		sum.ByType = append(sum.ByType, *b)
		sum.TotalCount += b.Count
		sum.TotalAmount += b.Amount
		sum.TotalTimes += b.Times
	}
	for _, p := range pack {
		sum.ByPackItem = append(sum.ByPackItem, *p)
	}
	return sum, nil
}

func TestLedgerRechargeConsumeAndStats(t *testing.T) {
	amt := 100
	negAmt := -30
	negTimes := -2
	ledgers := &memLedgerRepo{entries: []*domainledger.Entry{
		{ID: 1, StoreID: 1, MemberID: 2, CardID: 10, Type: domainledger.TypeRecharge, CardType: domaincard.TypeValue, Amount: &amt, CreatedAt: time.Now()},
		{ID: 2, StoreID: 1, MemberID: 2, CardID: 10, Type: domainledger.TypeConsumeValue, CardType: domaincard.TypeValue, Amount: &negAmt, CreatedAt: time.Now()},
		{ID: 3, StoreID: 1, MemberID: 2, CardID: 11, Type: domainledger.TypeConsumePack, CardType: domaincard.TypePack, Times: &negTimes, CreatedAt: time.Now(),
			Items: []*domainledger.EntryItem{
				{ProductItemID: 100, NameSnapshot: "洗", Times: 1},
				{ProductItemID: 101, NameSnapshot: "剪", Times: 1},
			},
		},
		{ID: 4, StoreID: 1, MemberID: 2, CardID: 10, Type: domainledger.TypeOpen, CardType: domaincard.TypeValue, CreatedAt: time.Now()},
	}}
	log := logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}})
	svc := NewLedgerService(
		ledgers,
		&memCustomerRepo{byID: map[int64]*domainmember.Member{2: {ID: 2, StoreID: 1, Name: "李", Phone: "1"}}},
		&memCardRepo{byID: map[int64]*domaincard.Card{
			10: {ID: 10, StoreID: 1, MemberID: 2, Type: domaincard.TypeValue, NameSnapshot: "储值"},
			11: {ID: 11, StoreID: 1, MemberID: 2, Type: domaincard.TypePack, NameSnapshot: "套餐"},
		}},
		&memUserRepo{byPhone: map[string]*domainuser.User{}, byID: map[int64]*domainuser.User{}},
		okStoreSvc{},
		log,
	)

	recharges, err := svc.ListRecharges(context.Background(), 9, 1, LedgerListQuery{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListRecharges: %v", err)
	}
	if recharges.Total != 1 || len(recharges.List) != 1 || recharges.List[0].Type != domainledger.TypeRecharge {
		t.Fatalf("unexpected recharges: %+v", recharges)
	}

	txns, err := svc.List(context.Background(), 9, 1, LedgerListQuery{Kind: "txn", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("List kind=txn: %v", err)
	}
	if txns.Total != 3 {
		t.Fatalf("txn list should be recharge+consumes (3), got %d", txns.Total)
	}

	consumes, err := svc.ListConsumes(context.Background(), 9, 1, LedgerListQuery{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("ListConsumes: %v", err)
	}
	if consumes.Total != 2 {
		t.Fatalf("want 2 consumes, got %d", consumes.Total)
	}
	packOnly, err := svc.ListConsumes(context.Background(), 9, 1, LedgerListQuery{
		Type: domainledger.TypeConsumePack, CardType: domaincard.TypePack, Page: 1, PageSize: 20,
	})
	if err != nil {
		t.Fatalf("ListConsumes pack: %v", err)
	}
	if packOnly.Total != 1 || len(packOnly.List[0].Items) != 2 {
		t.Fatalf("pack consume should include items: %+v", packOnly)
	}

	rStats, err := svc.StatsRecharges(context.Background(), 9, 1, LedgerListQuery{})
	if err != nil {
		t.Fatalf("StatsRecharges: %v", err)
	}
	if rStats.TotalCount != 1 || rStats.TotalAmount != 100 {
		t.Fatalf("recharge stats: %+v", rStats)
	}

	cStats, err := svc.StatsConsumes(context.Background(), 9, 1, LedgerListQuery{})
	if err != nil {
		t.Fatalf("StatsConsumes: %v", err)
	}
	if cStats.TotalCount != 2 || cStats.TotalAmount != 30 || cStats.TotalTimes != 2 {
		t.Fatalf("consume stats: %+v", cStats)
	}
	if len(cStats.ByPackItem) != 2 {
		t.Fatalf("want pack item stats, got %+v", cStats.ByPackItem)
	}

	if _, err := svc.GetRecharge(context.Background(), 9, 1, 2); err == nil {
		t.Fatal("consume id should not be a recharge")
	}
	if _, err := svc.GetConsume(context.Background(), 9, 1, 3); err != nil {
		t.Fatalf("GetConsume: %v", err)
	}
}
