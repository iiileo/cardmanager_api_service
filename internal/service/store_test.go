package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"card_manager/api_service/internal/api/dto"
	"card_manager/api_service/internal/config"
	domainstore "card_manager/api_service/internal/domain/store"
	domainmember "card_manager/api_service/internal/domain/storemember"
	domainuser "card_manager/api_service/internal/domain/user"
	"card_manager/api_service/internal/logger"
)

type memStoreRepo struct {
	byID    map[int64]*domainstore.Store
	byCode  map[string]*domainstore.Store
	seq     int64
	members *memMemberRepo
}

func (m *memStoreRepo) Create(_ context.Context, in domainstore.CreateInput) (*domainstore.Store, error) {
	m.seq++
	s := &domainstore.Store{
		ID: m.seq, Name: in.Name, City: in.City, Address: in.Address,
		OpenTime: in.OpenTime, CloseTime: in.CloseTime, BizType: in.BizType,
		InviteCode: in.InviteCode, OwnerUserID: in.OwnerUserID, Status: domainstore.StoreStatusNormal,
	}
	m.byID[s.ID] = s
	m.byCode[s.InviteCode] = s
	return s, nil
}

func (m *memStoreRepo) CreateWithOwner(ctx context.Context, in domainstore.CreateInput) (*domainstore.Store, error) {
	s, err := m.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if _, err := m.members.Create(ctx, domainmember.CreateInput{
		StoreID:  s.ID,
		UserID:   in.OwnerUserID,
		Role:     domainstore.RoleOwner,
		Status:   domainstore.MemberActive,
		JoinedAt: &now,
	}); err != nil {
		return nil, err
	}
	return s, nil
}
func (m *memStoreRepo) GetByID(_ context.Context, id int64) (*domainstore.Store, error) {
	return m.byID[id], nil
}
func (m *memStoreRepo) GetByInviteCode(_ context.Context, code string) (*domainstore.Store, error) {
	return m.byCode[code], nil
}
func (m *memStoreRepo) Update(_ context.Context, id int64, in domainstore.UpdateInput) (*domainstore.Store, error) {
	s := m.byID[id]
	if in.Name != nil {
		s.Name = *in.Name
	}
	if in.City != nil {
		s.City = *in.City
	}
	return s, nil
}
func (m *memStoreRepo) UpdateInviteCode(_ context.Context, id int64, code string) (*domainstore.Store, error) {
	s := m.byID[id]
	delete(m.byCode, s.InviteCode)
	s.InviteCode = code
	m.byCode[code] = s
	return s, nil
}
func (m *memStoreRepo) InviteCodeExists(_ context.Context, code string) (bool, error) {
	_, ok := m.byCode[code]
	return ok, nil
}

type memMemberRepo struct {
	byKey map[string]*domainmember.Member
	byID  map[int64]*domainmember.Member
	seq   int64
}

func pairKey(storeID, userID int64) string {
	return fmt.Sprintf("%d:%d", storeID, userID)
}

func (m *memMemberRepo) Create(_ context.Context, in domainmember.CreateInput) (*domainmember.Member, error) {
	m.seq++
	item := &domainmember.Member{
		ID: m.seq, StoreID: in.StoreID, UserID: in.UserID, Role: in.Role,
		Status: in.Status, DisplayName: in.DisplayName, JoinedAt: in.JoinedAt,
	}
	m.byID[item.ID] = item
	m.byKey[pairKey(in.StoreID, in.UserID)] = item
	return item, nil
}
func (m *memMemberRepo) GetByStoreUser(_ context.Context, storeID, userID int64) (*domainmember.Member, error) {
	return m.byKey[pairKey(storeID, userID)], nil
}
func (m *memMemberRepo) GetByID(_ context.Context, id int64) (*domainmember.Member, error) {
	return m.byID[id], nil
}
func (m *memMemberRepo) ListByUser(_ context.Context, userID int64) ([]*domainmember.Member, error) {
	var out []*domainmember.Member
	for _, item := range m.byID {
		if item.UserID == userID {
			out = append(out, item)
		}
	}
	return out, nil
}
func (m *memMemberRepo) ListByStore(_ context.Context, storeID int64, statuses []string) ([]*domainmember.Member, error) {
	allow := map[string]bool{}
	for _, s := range statuses {
		allow[s] = true
	}
	var out []*domainmember.Member
	for _, item := range m.byID {
		if item.StoreID == storeID && (len(allow) == 0 || allow[item.Status]) {
			out = append(out, item)
		}
	}
	return out, nil
}
func (m *memMemberRepo) CountActiveByStore(_ context.Context, storeID int64) (int, error) {
	n := 0
	for _, item := range m.byID {
		if item.StoreID == storeID && item.Status == domainstore.MemberActive {
			n++
		}
	}
	return n, nil
}
func (m *memMemberRepo) UpdateStatus(_ context.Context, id int64, status string, joinedAt *time.Time, displayName *string) (*domainmember.Member, error) {
	item := m.byID[id]
	item.Status = status
	if joinedAt != nil {
		item.JoinedAt = joinedAt
	}
	if displayName != nil {
		item.DisplayName = displayName
	}
	return item, nil
}

type memBizTypeSvc struct {
	codes map[string]string
}

func (m *memBizTypeSvc) List(_ context.Context) (*dto.BizTypeListResponse, error) {
	list := make([]*dto.BizTypeItem, 0, len(m.codes))
	for code, name := range m.codes {
		list = append(list, &dto.BizTypeItem{Code: code, Name: name})
	}
	return &dto.BizTypeListResponse{List: list}, nil
}

func (m *memBizTypeSvc) ValidateCode(_ context.Context, code string) error {
	if code == "" {
		return nil
	}
	if _, ok := m.codes[code]; !ok {
		return fmt.Errorf("invalid biz type")
	}
	return nil
}

func TestStoreService_CreateAndJoin(t *testing.T) {
	log := logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}})
	members := &memMemberRepo{byKey: map[string]*domainmember.Member{}, byID: map[int64]*domainmember.Member{}}
	stores := &memStoreRepo{
		byID: map[int64]*domainstore.Store{}, byCode: map[string]*domainstore.Store{}, members: members,
	}
	users := &memUserRepo{
		byID: map[int64]*domainuser.User{
			1: {ID: 1, Nickname: "老板A", Status: 1},
			2: {ID: 2, Nickname: "账号B", Status: 1},
		},
	}
	biz := &memBizTypeSvc{codes: map[string]string{"tea": "茶饮", "beauty": "美业", "retail": "零售"}}
	svc := NewStoreService(stores, members, users, biz, log)

	created, err := svc.Create(context.Background(), 1, dto.CreateStoreRequest{
		Name: "阳光茶饮", City: "杭州", OpenTime: "10:00", CloseTime: "22:00",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.InviteCode == "" || created.Role != "owner" {
		t.Fatalf("bad create: %+v", created)
	}
	if !inviteCodeRE.MatchString(created.InviteCode) {
		t.Fatalf("invite code should be 6 alnum uppercase, got %q", created.InviteCode)
	}
	hasLetter := false
	for _, c := range created.InviteCode {
		if c >= 'A' && c <= 'Z' {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		t.Fatalf("invite code should contain uppercase letter, got %q", created.InviteCode)
	}

	preview, err := svc.PreviewInvite(context.Background(), created.InviteCode)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Name != "阳光茶饮" {
		t.Fatalf("bad preview: %+v", preview)
	}

	joined, err := svc.Join(context.Background(), 2, dto.JoinStoreRequest{
		InviteCode: created.InviteCode, Nickname: "小周", Agreed: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if joined.Status != domainstore.MemberPending {
		t.Fatalf("want pending, got %s", joined.Status)
	}

	storeID := stores.seq
	apps, err := svc.ListApplications(context.Background(), 1, storeID)
	if err != nil {
		t.Fatal(err)
	}
	if len(apps.List) != 1 {
		t.Fatalf("want 1 application, got %d", len(apps.List))
	}
	memberID := members.seq
	if err := svc.Approve(context.Background(), 1, storeID, memberID); err != nil {
		t.Fatal(err)
	}
	staff, err := svc.ListStaff(context.Background(), 1, storeID)
	if err != nil {
		t.Fatal(err)
	}
	if len(staff.List) != 2 {
		t.Fatalf("want 2 staff, got %d", len(staff.List))
	}
	for _, item := range staff.List {
		if item.Nickname == "" {
			t.Fatalf("staff nickname required: %+v", item)
		}
		if item.DisplayName == nil || *item.DisplayName == "" {
			t.Fatalf("staff display_name required: %+v", item)
		}
	}
	byRole := map[string]*dto.StaffItem{}
	for _, item := range staff.List {
		byRole[item.Role] = item
	}
	owner := byRole[domainstore.RoleOwner]
	staffMember := byRole[domainstore.RoleStaff]
	if owner == nil || staffMember == nil {
		t.Fatalf("unexpected staff roles: %+v", staff.List)
	}
	if owner.Nickname != "老板A" || *owner.DisplayName != "老板A" {
		t.Fatalf("owner should fallback to nickname: %+v", owner)
	}
	if staffMember.Nickname != "账号B" || *staffMember.DisplayName != "小周" {
		t.Fatalf("staff should keep store display_name: %+v", staffMember)
	}
}
