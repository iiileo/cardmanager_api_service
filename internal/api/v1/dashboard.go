package v1

import (
	"card_manager/api_service/internal/api/response"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/service"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service service.DashboardService
	log     *logger.Logger
}

func NewDashboardHandler(svc service.DashboardService, log *logger.Logger) *DashboardHandler {
	return &DashboardHandler{service: svc, log: log}
}

func (h *DashboardHandler) HomeStats(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	resp, err := h.service.HomeStats(c.Request.Context(), userID, storeID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}
