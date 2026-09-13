package service

import (
	"context"
	"testing"

	"card_manager/api_service/internal/api/dto"
	"card_manager/api_service/internal/config"
	domainproduct "card_manager/api_service/internal/domain/cardproduct"
	domainmember "card_manager/api_service/internal/domain/storemember"
	domainstore "card_manager/api_service/internal/domain/store"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
)

type memCardProductRepo struct {
	byID  map[int64]*domainproduct.Product
	seq   int64
	itemSeq int64
}

func (m *memCardProductRepo) ListByStore(_ context.Context, storeID int64, typ string) ([]*domainproduct.Product, error) {
	var out []*domainproduct.Product
	for _, p := range m.byID {
		if p.StoreID != storeID || p.Status != domainproduct.StatusActive {
			continue
		}
		if typ != "" && p.Type != typ {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

func (m *memCardProductRepo) GetByID(_ context.Context, id int64) (*domainproduct.Product, error) {
	return m.byID[id], nil
}

func (m *memCardProductRepo) CountActiveByStoreType(_ context.Context, storeID int64, typ string) (int, error) {
	n := 0
	for _, p := range m.byID {
		if p.StoreID == storeID && p.Type == typ && p.Status == domainproduct.StatusActive {
			n++
		}
	}
	return n, nil
}

func (m *memCardProductRepo) Create(_ context.Context, in domainproduct.CreateInput) (*domainproduct.Product, error) {
	m.seq++
	p := &domainproduct.Product{
		ID: m.seq, StoreID: in.StoreID, Type: in.Type, Name: in.Name,
		Price: in.Price, Times: in.Times, ValidMonths: in.ValidMonths,
		Status: domainproduct.StatusActive,
	}
	for _, it := range in.Items {
		m.itemSeq++
		p.Items = append(p.Items, &domainproduct.Item{
			ID: m.itemSeq, ProductID: p.ID, Name: it.Name, Times: it.Times, Sort: it.Sort,
		})
	}
	m.byID[p.ID] = p
	return p, nil
}

func (m *memCardProductRepo) Delete(_ context.Context, id int64) error {
	delete(m.byID, id)
	return nil
}

type stubStoreSvc struct {
	ownerUserID int64
}

func (s *stubStoreSvc) ListMine(context.Context, int64) (*dto.StoreListResponse, error) {
	return nil, nil
}
func (s *stubStoreSvc) Create(context.Context, int64, dto.CreateStoreRequest) (*dto.StoreDetailResponse, error) {
	return nil, nil
}
func (s *stubStoreSvc) Get(context.Context, int64, int64) (*dto.StoreDetailResponse, error) {
	return nil, nil
}
func (s *stubStoreSvc) Update(context.Context, int64, int64, dto.UpdateStoreRequest) (*dto.StoreDetailResponse, error) {
	return nil, nil
}
func (s *stubStoreSvc) PreviewInvite(context.Context, string) (*dto.InvitePreviewResponse, error) {
	return nil, nil
}
func (s *stubStoreSvc) Join(context.Context, int64, dto.JoinStoreRequest) (*dto.StoreListItem, error) {
	return nil, nil
}
func (s *stubStoreSvc) GetInviteCode(context.Context, int64, int64) (*dto.InviteCodeResponse, error) {
	return nil, nil
}
func (s *stubStoreSvc) RefreshInviteCode(context.Context, int64, int64) (*dto.InviteCodeResponse, error) {
	return nil, nil
}
func (s *stubStoreSvc) ListStaff(context.Context, int64, int64) (*dto.StaffListResponse, error) {
	return nil, nil
}
func (s *stubStoreSvc) ListApplications(context.Context, int64, int64) (*dto.StaffListResponse, error) {
	return nil, nil
}
func (s *stubStoreSvc) Approve(context.Context, int64, int64, int64) error { return nil }
func (s *stubStoreSvc) Reject(context.Context, int64, int64, int64) error  { return nil }
func (s *stubStoreSvc) RequireActiveMember(_ context.Context, userID, storeID int64) (*domainmember.Member, error) {
	role := domainstore.RoleStaff
	if userID == s.ownerUserID {
		role = domainstore.RoleOwner
	}
	return &domainmember.Member{StoreID: storeID, UserID: userID, Role: role, Status: domainstore.MemberActive}, nil
}
func (s *stubStoreSvc) RequireOwner(ctx context.Context, userID, storeID int64) (*domainmember.Member, error) {
	m, err := s.RequireActiveMember(ctx, userID, storeID)
	if err != nil {
		return nil, err
	}
	if m.Role != domainstore.RoleOwner {
		return nil, ierr.Forbidden("仅门店老板可操作")
	}
	return m, nil
}

func TestCardProductService_UniqueValueAndCount(t *testing.T) {
	log := logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}})
	repo := &memCardProductRepo{byID: map[int64]*domainproduct.Product{}}
	storeSvc := &stubStoreSvc{ownerUserID: 1}
	svc := NewCardProductService(repo, storeSvc, log)

	price := 60000
	times := 20
	_, err := svc.Create(context.Background(), 1, 10, dto.CreateCardProductRequest{
		Type: "value", Name: "通用储值",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Create(context.Background(), 1, 10, dto.CreateCardProductRequest{
		Type: "value", Name: "另一张储值",
	})
	if err == nil {
		t.Fatal("expected conflict for second value card")
	}

	_, err = svc.Create(context.Background(), 1, 10, dto.CreateCardProductRequest{
		Type: "count", Name: "洗发 20 次", Price: &price, Times: &times,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Create(context.Background(), 1, 10, dto.CreateCardProductRequest{
		Type: "count", Name: "洗发 10 次", Price: &price, Times: &times,
	})
	if err == nil {
		t.Fatal("expected conflict for second count card")
	}

	packPrice := 128000
	for i := 0; i < 2; i++ {
		_, err = svc.Create(context.Background(), 1, 10, dto.CreateCardProductRequest{
			Type: "pack", Name: "套餐", Price: &packPrice,
			Items: []dto.CardProductItemRequest{{Name: "洗", Times: 5}},
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	list, err := svc.List(context.Background(), 1, 10, "pack")
	if err != nil {
		t.Fatal(err)
	}
	if len(list.List) != 2 {
		t.Fatalf("want 2 packs, got %d", len(list.List))
	}

	id := list.List[0].ID
	pid, _ := ParseProductID(id)
	if err := svc.Delete(context.Background(), 1, 10, pid); err != nil {
		t.Fatal(err)
	}
}
