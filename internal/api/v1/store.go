package v1

import (
	"card_manager/api_service/internal/api/dto"
	"card_manager/api_service/internal/api/middleware"
	"card_manager/api_service/internal/api/response"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/service"
	"github.com/gin-gonic/gin"
)

type StoreHandler struct {
	service service.StoreService
	log     *logger.Logger
}

func NewStoreHandler(svc service.StoreService, log *logger.Logger) *StoreHandler {
	return &StoreHandler{service: svc, log: log}
}

func (h *StoreHandler) List(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, ierr.Unauthorized(""))
		return
	}
	resp, err := h.service.ListMine(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *StoreHandler) Create(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, ierr.Unauthorized(""))
		return
	}
	var req dto.CreateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ierr.Validation(ierr.MsgBadRequest))
		return
	}
	resp, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *StoreHandler) Get(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, ierr.Unauthorized(""))
		return
	}
	storeID, err := service.ParseStoreID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	resp, err := h.service.Get(c.Request.Context(), userID, storeID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *StoreHandler) Update(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, ierr.Unauthorized(""))
		return
	}
	storeID, err := service.ParseStoreID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req dto.UpdateStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ierr.Validation(ierr.MsgBadRequest))
		return
	}
	resp, err := h.service.Update(c.Request.Context(), userID, storeID, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *StoreHandler) PreviewInvite(c *gin.Context) {
	resp, err := h.service.PreviewInvite(c.Request.Context(), c.Param("code"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *StoreHandler) Join(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, ierr.Unauthorized(""))
		return
	}
	var req dto.JoinStoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ierr.Validation(ierr.MsgBadRequest))
		return
	}
	resp, err := h.service.Join(c.Request.Context(), userID, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *StoreHandler) GetInviteCode(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, ierr.Unauthorized(""))
		return
	}
	storeID, err := service.ParseStoreID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	resp, err := h.service.GetInviteCode(c.Request.Context(), userID, storeID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *StoreHandler) RefreshInviteCode(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, ierr.Unauthorized(""))
		return
	}
	storeID, err := service.ParseStoreID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	resp, err := h.service.RefreshInviteCode(c.Request.Context(), userID, storeID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *StoreHandler) ListStaff(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	resp, err := h.service.ListStaff(c.Request.Context(), userID, storeID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *StoreHandler) ListApplications(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	resp, err := h.service.ListApplications(c.Request.Context(), userID, storeID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *StoreHandler) Approve(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	memberID, err := service.ParseMemberID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.Approve(c.Request.Context(), userID, storeID, memberID); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func (h *StoreHandler) Reject(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	memberID, err := service.ParseMemberID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.Reject(c.Request.Context(), userID, storeID, memberID); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, gin.H{})
}

func userAndStore(c *gin.Context) (userID, storeID int64, ok bool) {
	uid, ok := currentUserID(c)
	if !ok {
		response.Fail(c, ierr.Unauthorized(""))
		return 0, 0, false
	}
	v, exists := c.Get(middleware.CtxStoreID)
	if !exists {
		response.Fail(c, ierr.Validation("请选择门店"))
		return 0, 0, false
	}
	sid, ok := v.(int64)
	if !ok {
		response.Fail(c, ierr.Validation("门店 ID 不正确"))
		return 0, 0, false
	}
	return uid, sid, true
}
