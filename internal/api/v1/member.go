package v1

import (
	"strconv"

	"card_manager/api_service/internal/api/dto"
	"card_manager/api_service/internal/api/response"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/service"
	"github.com/gin-gonic/gin"
)

type MemberHandler struct {
	service service.MemberService
	log     *logger.Logger
}

func NewMemberHandler(svc service.MemberService, log *logger.Logger) *MemberHandler {
	return &MemberHandler{service: svc, log: log}
}

func (h *MemberHandler) List(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	resp, err := h.service.List(c.Request.Context(), userID, storeID, c.Query("q"), page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *MemberHandler) Get(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	memberID, err := service.ParseCustomerID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	resp, err := h.service.Get(c.Request.Context(), userID, storeID, memberID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *MemberHandler) OpenCard(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	var req dto.OpenCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ierr.Validation(ierr.MsgBadRequest))
		return
	}
	resp, err := h.service.OpenCard(c.Request.Context(), userID, storeID, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}
