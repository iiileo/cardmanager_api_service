package dto

type SendSMSRequest struct {
	Phone string `json:"phone" binding:"required"`
	Scene string `json:"scene" binding:"required"`
}

type LoginSMSRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateMeRequest struct {
	Nickname string `json:"nickname" binding:"required"`
}

type UpdatePhoneRequest struct {
	Phone string `json:"phone" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

// DeleteAccountRequest 注销账号。confirm 必须为 true。
type DeleteAccountRequest struct {
	Confirm bool `json:"confirm" binding:"required"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Phone    string `json:"phone"`
	Nickname string `json:"nickname"`
}

type TokenResponse struct {
	TokenType        string    `json:"token_type"`
	AccessToken      string    `json:"access_token"`
	AccessExpiresIn  int64     `json:"access_expires_in"`
	RefreshToken     string    `json:"refresh_token"`
	RefreshExpiresIn int64     `json:"refresh_expires_in"`
	User             *UserInfo `json:"user"`
}

type SendSMSResponse struct {
	ExpireIn int64  `json:"expire_in"`
	DevCode  string `json:"dev_code,omitempty"`
}
