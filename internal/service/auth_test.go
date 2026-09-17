package service

import (
	"context"
	"testing"

	"card_manager/api_service/internal/api/dto"
	"card_manager/api_service/internal/auth"
	"card_manager/api_service/internal/config"
	domainrt "card_manager/api_service/internal/domain/refreshtoken"
	domainsms "card_manager/api_service/internal/domain/smscode"
	domainuser "card_manager/api_service/internal/domain/user"
	"card_manager/api_service/internal/logger"
)

type memUserRepo struct {
	byPhone map[string]*domainuser.User
	byID    map[int64]*domainuser.User
	seq     int64
}

func (m *memUserRepo) GetByID(_ context.Context, id int64) (*domainuser.User, error) {
	return m.byID[id], nil
}
func (m *memUserRepo) ListByIDs(_ context.Context, ids []int64) (map[int64]*domainuser.User, error) {
	out := make(map[int64]*domainuser.User, len(ids))
	for _, id := range ids {
		if u := m.byID[id]; u != nil {
			out[id] = u
		}
	}
	return out, nil
}
func (m *memUserRepo) GetByPhone(_ context.Context, phone string) (*domainuser.User, error) {
	return m.byPhone[phone], nil
}
func (m *memUserRepo) Create(_ context.Context, phone, nickname string) (*domainuser.User, error) {
	m.seq++
	u := &domainuser.User{ID: m.seq, Phone: phone, Nickname: nickname, Status: 1}
	m.byPhone[phone] = u
	m.byID[u.ID] = u
	return u, nil
}
func (m *memUserRepo) UpdateNickname(_ context.Context, id int64, nickname string) (*domainuser.User, error) {
	u := m.byID[id]
	u.Nickname = nickname
	return u, nil
}
func (m *memUserRepo) UpdatePhone(_ context.Context, id int64, phone string) (*domainuser.User, error) {
	u := m.byID[id]
	delete(m.byPhone, u.Phone)
	u.Phone = phone
	m.byPhone[phone] = u
	return u, nil
}

type memSMSRepo struct {
	items []*domainsms.SmsCode
	seq   int64
}

func (m *memSMSRepo) Create(_ context.Context, in domainsms.CreateInput) (*domainsms.SmsCode, error) {
	m.seq++
	c := &domainsms.SmsCode{ID: m.seq, Phone: in.Phone, Scene: in.Scene, CodeHash: in.CodeHash, ExpiresAt: in.ExpiresAt}
	m.items = append(m.items, c)
	return c, nil
}
func (m *memSMSRepo) FindLatestValid(_ context.Context, phone, scene string) (*domainsms.SmsCode, error) {
	for i := len(m.items) - 1; i >= 0; i-- {
		c := m.items[i]
		if c.Phone == phone && c.Scene == scene && c.UsedAt == nil {
			return c, nil
		}
	}
	return nil, nil
}
func (m *memSMSRepo) MarkUsed(_ context.Context, id int64) error {
	for _, c := range m.items {
		if c.ID == id {
			now := c.ExpiresAt
			c.UsedAt = &now
		}
	}
	return nil
}

type memRTRepo struct {
	byHash map[string]*domainrt.RefreshToken
	seq    int64
}

func (m *memRTRepo) Create(_ context.Context, in domainrt.CreateInput) (*domainrt.RefreshToken, error) {
	m.seq++
	t := &domainrt.RefreshToken{ID: m.seq, UserID: in.UserID, TokenHash: in.TokenHash, ExpiresAt: in.ExpiresAt}
	m.byHash[in.TokenHash] = t
	return t, nil
}
func (m *memRTRepo) GetByHash(_ context.Context, hash string) (*domainrt.RefreshToken, error) {
	return m.byHash[hash], nil
}
func (m *memRTRepo) Revoke(_ context.Context, id int64) error {
	for _, t := range m.byHash {
		if t.ID == id {
			now := t.ExpiresAt
			t.RevokedAt = &now
		}
	}
	return nil
}

func TestAuthService_LoginSMS(t *testing.T) {
	cfg := &config.Config{Auth: config.AuthConfig{
		JWTSecret:         "test",
		AccessTTLSeconds:  7200,
		RefreshTTLSeconds: 2592000,
		SmsCodeTTLSeconds: 300,
		SmsDevCode:        "123456",
		SmsDevMode:        true,
	}}
	log := logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}})
	users := &memUserRepo{byPhone: map[string]*domainuser.User{}, byID: map[int64]*domainuser.User{}}
	sms := &memSMSRepo{}
	rt := &memRTRepo{byHash: map[string]*domainrt.RefreshToken{}}
	svc := NewAuthService(users, rt, sms, auth.NewTokenManager(cfg), cfg, log)

	if _, err := svc.SendSMS(context.Background(), dto.SendSMSRequest{Phone: "13800138000", Scene: "login"}); err != nil {
		t.Fatal(err)
	}
	tok, err := svc.LoginSMS(context.Background(), dto.LoginSMSRequest{Phone: "13800138000", Code: "123456"}, LoginMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if tok.AccessToken == "" || tok.RefreshToken == "" || tok.User == nil {
		t.Fatalf("incomplete token response: %+v", tok)
	}
}

func TestAuthService_UpdatePhone(t *testing.T) {
	cfg := &config.Config{Auth: config.AuthConfig{
		JWTSecret:         "test",
		AccessTTLSeconds:  7200,
		RefreshTTLSeconds: 2592000,
		SmsCodeTTLSeconds: 300,
		SmsDevCode:        "123456",
		SmsDevMode:        true,
	}}
	log := logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}})
	users := &memUserRepo{byPhone: map[string]*domainuser.User{}, byID: map[int64]*domainuser.User{}}
	sms := &memSMSRepo{}
	rt := &memRTRepo{byHash: map[string]*domainrt.RefreshToken{}}
	svc := NewAuthService(users, rt, sms, auth.NewTokenManager(cfg), cfg, log)

	users.byPhone["13800138000"] = &domainuser.User{ID: 1, Phone: "13800138000", Nickname: "店长", Status: 1}
	users.byID[1] = users.byPhone["13800138000"]

	if _, err := svc.SendSMS(context.Background(), dto.SendSMSRequest{
		Phone: "13900139000", Scene: "bind_phone",
	}); err != nil {
		t.Fatal(err)
	}

	user, err := svc.UpdatePhone(context.Background(), 1, dto.UpdatePhoneRequest{
		Phone: "13900139000", Code: "123456",
	})
	if err != nil {
		t.Fatal(err)
	}
	if user.Phone != "139****9000" {
		t.Fatalf("unexpected masked phone: %s", user.Phone)
	}
	if users.byID[1].Phone != "13900139000" {
		t.Fatalf("phone not updated in repo: %s", users.byID[1].Phone)
	}
}

func TestAuthService_UpdatePhone_Conflict(t *testing.T) {
	cfg := &config.Config{Auth: config.AuthConfig{
		SmsDevCode: "123456", SmsDevMode: true, SmsCodeTTLSeconds: 300,
		JWTSecret: "test", AccessTTLSeconds: 7200, RefreshTTLSeconds: 2592000,
	}}
	log := logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}})
	users := &memUserRepo{
		byPhone: map[string]*domainuser.User{
			"13800138000": {ID: 1, Phone: "13800138000", Status: 1},
			"13900139000": {ID: 2, Phone: "13900139000", Status: 1},
		},
		byID: map[int64]*domainuser.User{
			1: {ID: 1, Phone: "13800138000", Status: 1},
			2: {ID: 2, Phone: "13900139000", Status: 1},
		},
	}
	sms := &memSMSRepo{}
	svc := NewAuthService(users, &memRTRepo{byHash: map[string]*domainrt.RefreshToken{}}, sms, auth.NewTokenManager(cfg), cfg, log)

	if _, err := svc.SendSMS(context.Background(), dto.SendSMSRequest{
		Phone: "13900139000", Scene: "bind_phone",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdatePhone(context.Background(), 1, dto.UpdatePhoneRequest{
		Phone: "13900139000", Code: "123456",
	}); err == nil {
		t.Fatal("expected conflict")
	}
}
