package v1

import (
	"card_manager/api_service/internal/api/dto"
	"card_manager/api_service/internal/api/response"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/service"
	"github.com/gin-gonic/gin"
)

type CardHandler struct {
	service service.CardService
	log     *logger.Logger
}

func NewCardHandler(svc service.CardService, log *logger.Logger) *CardHandler {
	return &CardHandler{service: svc, log: log}
}

func (h *CardHandler) ListByMember(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	memberID, err := service.ParseCustomerID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	resp, err := h.service.ListByMember(c.Request.Context(), userID, storeID, memberID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *CardHandler) Get(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	cardID, err := service.ParseCardID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	resp, err := h.service.Get(c.Request.Context(), userID, storeID, cardID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *CardHandler) Recharge(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	cardID, err := service.ParseCardID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req dto.RechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ierr.Validation(ierr.MsgBadRequest))
		return
	}
	resp, err := h.service.Recharge(c.Request.Context(), userID, storeID, cardID, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *CardHandler) Consume(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	cardID, err := service.ParseCardID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req dto.ConsumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ierr.Validation(ierr.MsgBadRequest))
		return
	}
	resp, err := h.service.Consume(c.Request.Context(), userID, storeID, cardID, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}
