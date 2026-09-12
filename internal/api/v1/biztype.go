package v1

import (
	"card_manager/api_service/internal/api/response"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/service"
	"github.com/gin-gonic/gin"
)

type BizTypeHandler struct {
	service service.BizTypeService
	log     *logger.Logger
}

func NewBizTypeHandler(svc service.BizTypeService, log *logger.Logger) *BizTypeHandler {
	return &BizTypeHandler{service: svc, log: log}
}

func (h *BizTypeHandler) List(c *gin.Context) {
	resp, err := h.service.List(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}
