package v1

import (
	"card_manager/api_service/internal/api/dto"
	"card_manager/api_service/internal/api/response"
	"card_manager/api_service/internal/auth"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/service"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service service.AuthService
	tokens  *auth.TokenManager
	log     *logger.Logger
}

func NewAuthHandler(svc service.AuthService, tokens *auth.TokenManager, log *logger.Logger) *AuthHandler {
	return &AuthHandler{service: svc, tokens: tokens, log: log}
}

func (h *AuthHandler) SendSMS(c *gin.Context) {
	var req dto.SendSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ierr.Validation(ierr.MsgBadRequest))
		return
	}
	resp, err := h.service.SendSMS(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *AuthHandler) LoginSMS(c *gin.Context) {
	var req dto.LoginSMSRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ierr.Validation(ierr.MsgBadRequest))
		return
	}
	resp, err := h.service.LoginSMS(c.Request.Context(), req, loginMeta(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ierr.Validation(ierr.MsgBadRequest))
		return
	}
	resp, err := h.service.Refresh(c.Request.Context(), req.RefreshToken, loginMeta(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ierr.Validation(ierr.MsgBadRequest))
		return
	}
	if err := h.service.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, ierr.Unauthorized(""))
		return
	}
	resp, err := h.service.Me(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *AuthHandler) UpdateMe(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, ierr.Unauthorized(""))
		return
	}
	var req dto.UpdateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ierr.Validation(ierr.MsgBadRequest))
		return
	}
	resp, err := h.service.UpdateMe(c.Request.Context(), userID, req.Nickname)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func loginMeta(c *gin.Context) service.LoginMeta {
	return service.LoginMeta{
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		DeviceID:  c.GetHeader("X-Device-Id"),
	}
}

func currentUserID(c *gin.Context) (int64, bool) {
	v, ok := c.Get("user_id")
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}
