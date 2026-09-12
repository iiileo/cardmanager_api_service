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
