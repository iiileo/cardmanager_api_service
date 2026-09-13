package v1

import (
	"strconv"

	"card_manager/api_service/internal/api/response"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/service"
	"github.com/gin-gonic/gin"
)

type LedgerHandler struct {
	service service.LedgerService
	log     *logger.Logger
}

func NewLedgerHandler(svc service.LedgerService, log *logger.Logger) *LedgerHandler {
	return &LedgerHandler{service: svc, log: log}
}

func (h *LedgerHandler) List(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	resp, err := h.service.List(
		c.Request.Context(), userID, storeID,
		c.Query("type"), c.Query("from"), c.Query("to"),
		page, pageSize,
	)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *LedgerHandler) Get(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	ledgerID, err := service.ParseLedgerID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	resp, err := h.service.Get(c.Request.Context(), userID, storeID, ledgerID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}
