package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"card_manager/api_service/internal/api/dto"
	"card_manager/api_service/internal/auth"
	"card_manager/api_service/internal/config"
	domainrt "card_manager/api_service/internal/domain/refreshtoken"
	domainsms "card_manager/api_service/internal/domain/smscode"
	domainuser "card_manager/api_service/internal/domain/user"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
)

var phoneRE = regexp.MustCompile(`^1\d{10}$`)

type AuthService interface {
	SendSMS(ctx context.Context, req dto.SendSMSRequest) (*dto.SendSMSResponse, error)
	LoginSMS(ctx context.Context, req dto.LoginSMSRequest, meta LoginMeta) (*dto.TokenResponse, error)
	Refresh(ctx context.Context, refreshToken string, meta LoginMeta) (*dto.TokenResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	Me(ctx context.Context, userID int64) (*dto.UserInfo, error)
	UpdateMe(ctx context.Context, userID int64, nickname string) (*dto.UserInfo, error)
}

type LoginMeta struct {
	IP        string
	UserAgent string
	DeviceID  string
}

type authService struct {
	users  domainuser.Repository
	tokens domainrt.Repository
	sms    domainsms.Repository
	tm     *auth.TokenManager
	cfg    *config.Config
	log    *logger.Logger
}

func NewAuthService(
	users domainuser.Repository,
	tokens domainrt.Repository,
	sms domainsms.Repository,
	tm *auth.TokenManager,
	cfg *config.Config,
	log *logger.Logger,
) AuthService {
	return &authService{
		users:  users,
		tokens: tokens,
		sms:    sms,
		tm:     tm,
		cfg:    cfg,
		log:    log,
	}
}

func (s *authService) SendSMS(ctx context.Context, req dto.SendSMSRequest) (*dto.SendSMSResponse, error) {
	phone := strings.TrimSpace(req.Phone)
	scene := strings.TrimSpace(req.Scene)
	if !phoneRE.MatchString(phone) {
		return nil, ierr.NewError("invalid phone").WithHint("phone must be 11-digit CN mobile").Mark(ierr.ErrValidation)
	}
	if scene != "login" && scene != "bind_phone" {
		return nil, ierr.NewError("invalid scene").WithHint("scene must be login or bind_phone").Mark(ierr.ErrValidation)
	}

	code := s.cfg.Auth.SmsDevCode
	if !s.cfg.Auth.SmsDevMode {
		var err error
		code, err = auth.RandomDigits(6)
		if err != nil {
			return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
		}
	}

	expiresAt := time.Now().UTC().Add(time.Duration(s.cfg.Auth.SmsCodeTTLSeconds) * time.Second)
	if _, err := s.sms.Create(ctx, domainsms.CreateInput{
		Phone:     phone,
		Scene:     scene,
		CodeHash:  auth.HashCode(code),
		ExpiresAt: expiresAt,
	}); err != nil {
		return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
	}

	s.log.Info(ctx, "sms code created", "phone", phone, "scene", scene)
	resp := &dto.SendSMSResponse{ExpireIn: s.cfg.Auth.SmsCodeTTLSeconds}
	if s.cfg.Auth.SmsDevMode {
		resp.DevCode = code
	}
	return resp, nil
}

func (s *authService) LoginSMS(ctx context.Context, req dto.LoginSMSRequest, meta LoginMeta) (*dto.TokenResponse, error) {
	phone := strings.TrimSpace(req.Phone)
	code := strings.TrimSpace(req.Code)
	if !phoneRE.MatchString(phone) {
		return nil, ierr.NewError("invalid phone").Mark(ierr.ErrValidation)
	}
	if len(code) != 6 {
		return nil, ierr.NewError("invalid code").WithHint("code must be 6 digits").Mark(ierr.ErrValidation)
	}

	record, err := s.sms.FindLatestValid(ctx, phone, "login")
	if err != nil {
		return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
	}
	if record == nil || record.CodeHash != auth.HashCode(code) {
		return nil, ierr.NewError("invalid or expired code").Mark(ierr.ErrValidation).WithCode(40001)
	}
	if err := s.sms.MarkUsed(ctx, record.ID); err != nil {
		return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
	}

	u, err := s.users.GetByPhone(ctx, phone)
	if err != nil {
		return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
	}
	if u == nil {
		u, err = s.users.Create(ctx, phone, "")
		if err != nil {
			return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
		}
	}
	if u.Status != 1 {
		return nil, ierr.NewError("user disabled").Mark(ierr.ErrForbidden)
	}

	return s.issueTokens(ctx, u, meta)
}

func (s *authService) Refresh(ctx context.Context, refreshToken string, meta LoginMeta) (*dto.TokenResponse, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ierr.NewError("refresh_token is required").Mark(ierr.ErrValidation)
	}
	hash := auth.HashToken(refreshToken)
	stored, err := s.tokens.GetByHash(ctx, hash)
	if err != nil {
		return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
	}
	if stored == nil || stored.RevokedAt != nil || time.Now().UTC().After(stored.ExpiresAt) {
		return nil, ierr.NewError("refresh token invalid").Mark(ierr.ErrUnauthorized)
	}
	if err := s.tokens.Revoke(ctx, stored.ID); err != nil {
		return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
	}

	u, err := s.users.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
	}
	if u == nil || u.Status != 1 {
		return nil, ierr.NewError("user not found").Mark(ierr.ErrUnauthorized)
	}
	return s.issueTokens(ctx, u, meta)
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	hash := auth.HashToken(strings.TrimSpace(refreshToken))
	stored, err := s.tokens.GetByHash(ctx, hash)
	if err != nil {
		return ierr.WithError(err).Mark(ierr.ErrInternal)
	}
	if stored == nil {
		return nil
	}
	if stored.RevokedAt != nil {
		return nil
	}
	return s.tokens.Revoke(ctx, stored.ID)
}

func (s *authService) Me(ctx context.Context, userID int64) (*dto.UserInfo, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
	}
	if u == nil {
		return nil, ierr.NewError("user not found").Mark(ierr.ErrUnauthorized)
	}
	return toUserInfo(u), nil
}

func (s *authService) UpdateMe(ctx context.Context, userID int64, nickname string) (*dto.UserInfo, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" || len([]rune(nickname)) > 32 {
		return nil, ierr.NewError("invalid nickname").Mark(ierr.ErrValidation)
	}
	u, err := s.users.UpdateNickname(ctx, userID, nickname)
	if err != nil {
		return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
	}
	return toUserInfo(u), nil
}

func (s *authService) issueTokens(ctx context.Context, u *domainuser.User, meta LoginMeta) (*dto.TokenResponse, error) {
	access, err := s.tm.IssueAccess(u.ID, u.Phone)
	if err != nil {
		return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
	}
	raw, hash, expiresAt, err := s.tm.NewRefreshToken()
	if err != nil {
		return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
	}

	in := domainrt.CreateInput{
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: expiresAt,
	}
	if meta.DeviceID != "" {
		in.DeviceID = &meta.DeviceID
	}
	if meta.UserAgent != "" {
		in.UserAgent = &meta.UserAgent
	}
	if meta.IP != "" {
		in.IP = &meta.IP
	}
	if _, err := s.tokens.Create(ctx, in); err != nil {
		return nil, ierr.WithError(err).Mark(ierr.ErrInternal)
	}

	s.log.Info(ctx, "tokens issued", "user_id", u.ID)
	return &dto.TokenResponse{
		TokenType:        "Bearer",
		AccessToken:      access,
		AccessExpiresIn:  s.tm.AccessTTLSeconds(),
		RefreshToken:     raw,
		RefreshExpiresIn: s.tm.RefreshTTLSeconds(),
		User:             toUserInfo(u),
	}, nil
}

func toUserInfo(u *domainuser.User) *dto.UserInfo {
	return &dto.UserInfo{
		ID:       fmt.Sprintf("%d", u.ID),
		Phone:    maskPhoneDisplay(u.Phone),
		Nickname: u.Nickname,
	}
}

func maskPhoneDisplay(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}
