package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"card_manager/api_service/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type TokenPair struct {
	TokenType        string
	AccessToken      string
	AccessExpiresIn  int64
	RefreshToken     string
	RefreshExpiresIn int64
}

type AccessClaims struct {
	UserID int64  `json:"uid"`
	Phone  string `json:"phone"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenManager(cfg *config.Config) *TokenManager {
	return &TokenManager{
		secret:     []byte(cfg.Auth.JWTSecret),
		accessTTL:  time.Duration(cfg.Auth.AccessTTLSeconds) * time.Second,
		refreshTTL: time.Duration(cfg.Auth.RefreshTTLSeconds) * time.Second,
	}
}

func (m *TokenManager) AccessTTLSeconds() int64 {
	return int64(m.accessTTL.Seconds())
}

func (m *TokenManager) RefreshTTLSeconds() int64 {
	return int64(m.refreshTTL.Seconds())
}

func (m *TokenManager) IssueAccess(userID int64, phone string) (string, error) {
	now := time.Now().UTC()
	claims := AccessClaims{
		UserID: userID,
		Phone:  phone,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(m.secret)
}

func (m *TokenManager) ParseAccess(token string) (*AccessClaims, error) {
	parsed, err := jwt.ParseWithClaims(token, &AccessClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*AccessClaims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func (m *TokenManager) NewRefreshToken() (raw string, hash string, expiresAt time.Time, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", time.Time{}, err
	}
	raw = hex.EncodeToString(buf)
	hash = HashToken(raw)
	expiresAt = time.Now().UTC().Add(m.refreshTTL)
	return raw, hash, expiresAt, nil
}

func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func HashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func RandomDigits(n int) (string, error) {
	const digits = "0123456789"
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	out := make([]byte, n)
	for i := range out {
		out[i] = digits[int(buf[i])%10]
	}
	return string(out), nil
}
