package v1

import (
	"card_manager/api_service/internal/api/dto"
	"card_manager/api_service/internal/api/response"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/service"
	"github.com/gin-gonic/gin"
)

type NotifySettingHandler struct {
	service service.NotifySettingService
	log     *logger.Logger
}

func NewNotifySettingHandler(svc service.NotifySettingService, log *logger.Logger) *NotifySettingHandler {
	return &NotifySettingHandler{service: svc, log: log}
}

func (h *NotifySettingHandler) List(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	resp, err := h.service.List(c.Request.Context(), userID, storeID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *NotifySettingHandler) Get(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	resp, err := h.service.Get(c.Request.Context(), userID, storeID, c.Param("event"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *NotifySettingHandler) Update(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	var req dto.UpdateNotifySettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ierr.Validation(ierr.MsgBadRequest))
		return
	}
	resp, err := h.service.Update(c.Request.Context(), userID, storeID, c.Param("event"), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}
