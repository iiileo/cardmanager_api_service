package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"card_manager/api_service/internal/api/dto"
	domainstore "card_manager/api_service/internal/domain/store"
	domainmember "card_manager/api_service/internal/domain/storemember"
	domainuser "card_manager/api_service/internal/domain/user"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
)

var (
	timeRE       = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)
	inviteCodeRE = regexp.MustCompile(`^[0-9A-Z]{6}$`)
)

type StoreService interface {
	ListMine(ctx context.Context, userID int64) (*dto.StoreListResponse, error)
	Create(ctx context.Context, userID int64, req dto.CreateStoreRequest) (*dto.StoreDetailResponse, error)
	Get(ctx context.Context, userID, storeID int64) (*dto.StoreDetailResponse, error)
	Update(ctx context.Context, userID, storeID int64, req dto.UpdateStoreRequest) (*dto.StoreDetailResponse, error)
	PreviewInvite(ctx context.Context, code string) (*dto.InvitePreviewResponse, error)
	Join(ctx context.Context, userID int64, req dto.JoinStoreRequest) (*dto.StoreListItem, error)
	GetInviteCode(ctx context.Context, userID, storeID int64) (*dto.InviteCodeResponse, error)
	RefreshInviteCode(ctx context.Context, userID, storeID int64) (*dto.InviteCodeResponse, error)

	ListStaff(ctx context.Context, userID, storeID int64) (*dto.StaffListResponse, error)
	ListApplications(ctx context.Context, userID, storeID int64) (*dto.StaffListResponse, error)
	Approve(ctx context.Context, userID, storeID, memberID int64) error
	Reject(ctx context.Context, userID, storeID, memberID int64) error

	RequireActiveMember(ctx context.Context, userID, storeID int64) (*domainmember.Member, error)
	RequireOwner(ctx context.Context, userID, storeID int64) (*domainmember.Member, error)
}

type storeService struct {
	stores   domainstore.Repository
	members  domainmember.Repository
	users    domainuser.Repository
	bizTypes BizTypeService
	log      *logger.Logger
}

func NewStoreService(
	stores domainstore.Repository,
	members domainmember.Repository,
	users domainuser.Repository,
	bizTypes BizTypeService,
	log *logger.Logger,
) StoreService {
	return &storeService{stores: stores, members: members, users: users, bizTypes: bizTypes, log: log}
}

func (s *storeService) ListMine(ctx context.Context, userID int64) (*dto.StoreListResponse, error) {
	ms, err := s.members.ListByUser(ctx, userID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	list := make([]*dto.StoreListItem, 0, len(ms))
	for _, m := range ms {
		st, err := s.stores.GetByID(ctx, m.StoreID)
		if err != nil {
			return nil, ierr.Internal(err)
		}
		if st == nil {
			continue
		}
		count := 0
		if m.Status == domainstore.MemberActive {
			count, err = s.members.CountActiveByStore(ctx, st.ID)
			if err != nil {
				return nil, ierr.Internal(err)
			}
		}
		list = append(list, &dto.StoreListItem{
			ID:           fmt.Sprintf("%d", st.ID),
			Name:         st.Name,
			Role:         m.Role,
			Status:       m.Status,
			City:         st.City,
			MembersCount: count,
		})
	}
	return &dto.StoreListResponse{List: list}, nil
}

func (s *storeService) Create(ctx context.Context, userID int64, req dto.CreateStoreRequest) (*dto.StoreDetailResponse, error) {
	name := strings.TrimSpace(req.Name)
	city := strings.TrimSpace(req.City)
	openTime := strings.TrimSpace(req.OpenTime)
	closeTime := strings.TrimSpace(req.CloseTime)
	if name == "" || city == "" {
		return nil, ierr.Validation("请填写门店名称和城市")
	}
	if !timeRE.MatchString(openTime) || !timeRE.MatchString(closeTime) {
		return nil, ierr.Validation("营业时间格式应为 HH:MM")
	}
	bizType := trimPtr(req.BizType)
	if bizType != nil {
		if err := s.bizTypes.ValidateCode(ctx, *bizType); err != nil {
			return nil, err
		}
	}

	code, err := s.uniqueInviteCode(ctx)
	if err != nil {
		return nil, err
	}

	st, err := s.stores.CreateWithOwner(ctx, domainstore.CreateInput{
		Name:        name,
		City:        city,
		Address:     trimPtr(req.Address),
		OpenTime:    openTime,
		CloseTime:   closeTime,
		BizType:     bizType,
		InviteCode:  code,
		OwnerUserID: userID,
	})
	if err != nil {
		return nil, ierr.Internal(err)
	}

	s.log.Info(ctx, "store created", "store_id", st.ID, "user_id", userID)
	return s.toDetail(ctx, st, domainstore.RoleOwner, domainstore.MemberActive, true)
}

func (s *storeService) Get(ctx context.Context, userID, storeID int64) (*dto.StoreDetailResponse, error) {
	m, err := s.members.GetByStoreUser(ctx, storeID, userID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	if m == nil {
		return nil, ierr.Forbidden("你不在该门店中")
	}
	st, err := s.stores.GetByID(ctx, storeID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	if st == nil {
		return nil, ierr.NotFound("门店不存在")
	}
	showInvite := m.Role == domainstore.RoleOwner && m.Status == domainstore.MemberActive
	return s.toDetail(ctx, st, m.Role, m.Status, showInvite)
}

func (s *storeService) Update(ctx context.Context, userID, storeID int64, req dto.UpdateStoreRequest) (*dto.StoreDetailResponse, error) {
	if _, err := s.RequireOwner(ctx, userID, storeID); err != nil {
		return nil, err
	}
	in := domainstore.UpdateInput{}
	if req.Name != nil {
		v := strings.TrimSpace(*req.Name)
		if v == "" {
			return nil, ierr.Validation("门店名称不能为空")
		}
		in.Name = &v
	}
	if req.City != nil {
		v := strings.TrimSpace(*req.City)
		if v == "" {
			return nil, ierr.Validation("城市不能为空")
		}
		in.City = &v
	}
	if req.Address != nil {
		in.Address = trimPtr(req.Address)
	}
	if req.OpenTime != nil {
		v := strings.TrimSpace(*req.OpenTime)
		if !timeRE.MatchString(v) {
			return nil, ierr.Validation("营业时间格式应为 HH:MM")
		}
		in.OpenTime = &v
	}
	if req.CloseTime != nil {
		v := strings.TrimSpace(*req.CloseTime)
		if !timeRE.MatchString(v) {
			return nil, ierr.Validation("营业时间格式应为 HH:MM")
		}
		in.CloseTime = &v
	}
	if req.BizType != nil {
		bizType := trimPtr(req.BizType)
		if bizType != nil {
			if err := s.bizTypes.ValidateCode(ctx, *bizType); err != nil {
				return nil, err
			}
		}
		in.BizType = bizType
	}

	st, err := s.stores.Update(ctx, storeID, in)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	return s.toDetail(ctx, st, domainstore.RoleOwner, domainstore.MemberActive, true)
}

func (s *storeService) PreviewInvite(ctx context.Context, code string) (*dto.InvitePreviewResponse, error) {
	code = normalizeInviteCode(code)
	if !inviteCodeRE.MatchString(code) {
		return nil, ierr.Validation("邀请码应为6位大写字母或数字")
	}
	st, err := s.stores.GetByInviteCode(ctx, code)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	if st == nil || st.Status != domainstore.StoreStatusNormal {
		return nil, ierr.NotFound("邀请码无效")
	}
	return &dto.InvitePreviewResponse{
		ID:   fmt.Sprintf("%d", st.ID),
		Name: st.Name,
		City: st.City,
	}, nil
}

func (s *storeService) Join(ctx context.Context, userID int64, req dto.JoinStoreRequest) (*dto.StoreListItem, error) {
	if !req.Agreed {
		return nil, ierr.Validation("请先同意加入申请")
	}
	code := normalizeInviteCode(req.InviteCode)
	if !inviteCodeRE.MatchString(code) {
		return nil, ierr.Validation("邀请码应为6位大写字母或数字")
	}
	st, err := s.stores.GetByInviteCode(ctx, code)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	if st == nil || st.Status != domainstore.StoreStatusNormal {
		return nil, ierr.NotFound("邀请码无效")
	}
	if st.OwnerUserID == userID {
		return nil, ierr.Conflict("你已是该门店老板")
	}

	existing, err := s.members.GetByStoreUser(ctx, st.ID, userID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	nickname := strings.TrimSpace(req.Nickname)
	var display *string
	if nickname != "" {
		display = &nickname
	}

	var m *domainmember.Member
	switch {
	case existing == nil:
		m, err = s.members.Create(ctx, domainmember.CreateInput{
			StoreID:     st.ID,
			UserID:      userID,
			Role:        domainstore.RoleStaff,
			Status:      domainstore.MemberPending,
			DisplayName: display,
		})
		if err != nil {
			return nil, ierr.Internal(err)
		}
	case existing.Status == domainstore.MemberActive:
		return nil, ierr.Conflict("你已在该门店中")
	case existing.Status == domainstore.MemberPending:
		return nil, ierr.Conflict("申请已提交，请等待老板同意")
	default:
		m, err = s.members.UpdateStatus(ctx, existing.ID, domainstore.MemberPending, nil, display)
		if err != nil {
			return nil, ierr.Internal(err)
		}
	}

	s.log.Info(ctx, "store join requested", "store_id", st.ID, "user_id", userID)
	return &dto.StoreListItem{
		ID:     fmt.Sprintf("%d", st.ID),
		Name:   st.Name,
		Role:   m.Role,
		Status: m.Status,
		City:   st.City,
	}, nil
}

func (s *storeService) GetInviteCode(ctx context.Context, userID, storeID int64) (*dto.InviteCodeResponse, error) {
	if _, err := s.RequireOwner(ctx, userID, storeID); err != nil {
		return nil, err
	}
	st, err := s.stores.GetByID(ctx, storeID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	if st == nil {
		return nil, ierr.NotFound("门店不存在")
	}
	return &dto.InviteCodeResponse{InviteCode: st.InviteCode}, nil
}

func (s *storeService) RefreshInviteCode(ctx context.Context, userID, storeID int64) (*dto.InviteCodeResponse, error) {
	if _, err := s.RequireOwner(ctx, userID, storeID); err != nil {
		return nil, err
	}
	code, err := s.uniqueInviteCode(ctx)
	if err != nil {
		return nil, err
	}
	st, err := s.stores.UpdateInviteCode(ctx, storeID, code)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	return &dto.InviteCodeResponse{InviteCode: st.InviteCode}, nil
}

func (s *storeService) ListStaff(ctx context.Context, userID, storeID int64) (*dto.StaffListResponse, error) {
	if _, err := s.RequireActiveMember(ctx, userID, storeID); err != nil {
		return nil, err
	}
	list, err := s.members.ListByStore(ctx, storeID, []string{domainstore.MemberActive})
	if err != nil {
		return nil, ierr.Internal(err)
	}
	items, err := s.toStaffItems(ctx, list)
	if err != nil {
		return nil, err
	}
	return &dto.StaffListResponse{List: items}, nil
}

func (s *storeService) ListApplications(ctx context.Context, userID, storeID int64) (*dto.StaffListResponse, error) {
	if _, err := s.RequireOwner(ctx, userID, storeID); err != nil {
		return nil, err
	}
	list, err := s.members.ListByStore(ctx, storeID, []string{domainstore.MemberPending})
	if err != nil {
		return nil, ierr.Internal(err)
	}
	items, err := s.toStaffItems(ctx, list)
	if err != nil {
		return nil, err
	}
	return &dto.StaffListResponse{List: items}, nil
}

func (s *storeService) Approve(ctx context.Context, userID, storeID, memberID int64) error {
	if _, err := s.RequireOwner(ctx, userID, storeID); err != nil {
		return err
	}
	m, err := s.members.GetByID(ctx, memberID)
	if err != nil {
		return ierr.Internal(err)
	}
	if m == nil || m.StoreID != storeID {
		return ierr.NotFound("申请不存在")
	}
	if m.Status != domainstore.MemberPending {
		return ierr.Conflict("该申请已处理")
	}
	now := time.Now().UTC()
	if _, err := s.members.UpdateStatus(ctx, m.ID, domainstore.MemberActive, &now, nil); err != nil {
		return ierr.Internal(err)
	}
	s.log.Info(ctx, "store member approved", "store_id", storeID, "member_id", memberID)
	return nil
}

func (s *storeService) Reject(ctx context.Context, userID, storeID, memberID int64) error {
	if _, err := s.RequireOwner(ctx, userID, storeID); err != nil {
		return err
	}
	m, err := s.members.GetByID(ctx, memberID)
	if err != nil {
		return ierr.Internal(err)
	}
	if m == nil || m.StoreID != storeID {
		return ierr.NotFound("申请不存在")
	}
	if m.Status != domainstore.MemberPending {
		return ierr.Conflict("该申请已处理")
	}
	if _, err := s.members.UpdateStatus(ctx, m.ID, domainstore.MemberRejected, nil, nil); err != nil {
		return ierr.Internal(err)
	}
	s.log.Info(ctx, "store member rejected", "store_id", storeID, "member_id", memberID)
	return nil
}

func (s *storeService) RequireActiveMember(ctx context.Context, userID, storeID int64) (*domainmember.Member, error) {
	m, err := s.members.GetByStoreUser(ctx, storeID, userID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	if m == nil {
		return nil, ierr.Forbidden("你不在该门店中")
	}
	if m.Status == domainstore.MemberPending {
		return nil, ierr.Forbidden("加入申请待老板同意").WithCode(40301)
	}
	if m.Status != domainstore.MemberActive {
		return nil, ierr.Forbidden("无法进入该门店")
	}
	return m, nil
}

func (s *storeService) RequireOwner(ctx context.Context, userID, storeID int64) (*domainmember.Member, error) {
	m, err := s.RequireActiveMember(ctx, userID, storeID)
	if err != nil {
		return nil, err
	}
	if m.Role != domainstore.RoleOwner {
		return nil, ierr.Forbidden("仅门店老板可操作")
	}
	return m, nil
}

func (s *storeService) uniqueInviteCode(ctx context.Context) (string, error) {
	for i := 0; i < 20; i++ {
		code, err := randomInviteCode(6)
		if err != nil {
			return "", ierr.Internal(err)
		}
		exists, err := s.stores.InviteCodeExists(ctx, code)
		if err != nil {
			return "", ierr.Internal(err)
		}
		if !exists {
			return code, nil
		}
	}
	return "", ierr.Internal(fmt.Errorf("invite code exhausted"))
}

func (s *storeService) toDetail(ctx context.Context, st *domainstore.Store, role, status string, showInvite bool) (*dto.StoreDetailResponse, error) {
	count, err := s.members.CountActiveByStore(ctx, st.ID)
	if err != nil {
		return nil, ierr.Internal(err)
	}
	resp := &dto.StoreDetailResponse{
		ID:           fmt.Sprintf("%d", st.ID),
		Name:         st.Name,
		City:         st.City,
		Address:      st.Address,
		OpenTime:     st.OpenTime,
		CloseTime:    st.CloseTime,
		BizType:      st.BizType,
		OwnerUserID:  fmt.Sprintf("%d", st.OwnerUserID),
		Role:         role,
		Status:       status,
		MembersCount: count,
	}
	if showInvite {
		resp.InviteCode = st.InviteCode
	}
	return resp, nil
}

func (s *storeService) toStaffItems(ctx context.Context, list []*domainmember.Member) ([]*dto.StaffItem, error) {
	nickCache := make(map[int64]string, len(list))
	out := make([]*dto.StaffItem, 0, len(list))
	for _, m := range list {
		nickname, ok := nickCache[m.UserID]
		if !ok {
			u, err := s.users.GetByID(ctx, m.UserID)
			if err != nil {
				return nil, ierr.Internal(err)
			}
			if u != nil {
				nickname = u.Nickname
			}
			nickCache[m.UserID] = nickname
		}

		displayName := m.DisplayName
		if displayName == nil || strings.TrimSpace(*displayName) == "" {
			if nickname != "" {
				displayName = &nickname
			}
		}

		item := &dto.StaffItem{
			ID:          fmt.Sprintf("%d", m.ID),
			UserID:      fmt.Sprintf("%d", m.UserID),
			Role:        m.Role,
			Status:      m.Status,
			Nickname:    nickname,
			DisplayName: displayName,
		}
		if m.JoinedAt != nil {
			v := m.JoinedAt.UTC().Format(time.RFC3339)
			item.JoinedAt = &v
		}
		out = append(out, item)
	}
	return out, nil
}

func trimPtr(v *string) *string {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return &s
}

func normalizeInviteCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// randomInviteCode 生成含数字与大写字母的邀请码，且至少包含 1 个大写字母。
func randomInviteCode(n int) (string, error) {
	const charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if n < 2 {
		return "", fmt.Errorf("invite code length too short")
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i := range out {
		out[i] = charset[int(buf[i])%len(charset)]
	}
	hasLetter := false
	for _, c := range out {
		if c >= 'A' && c <= 'Z' {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		pos := int(buf[0]) % n
		letterBuf := make([]byte, 1)
		if _, err := rand.Read(letterBuf); err != nil {
			return "", err
		}
		out[pos] = letters[int(letterBuf[0])%len(letters)]
	}
	return string(out), nil
}

func ParseStoreID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, ierr.Validation("门店 ID 不正确")
	}
	return id, nil
}

func ParseMemberID(raw string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || id <= 0 {
		return 0, ierr.Validation("成员 ID 不正确")
	}
	return id, nil
}
