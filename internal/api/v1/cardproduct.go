package v1

import (
	"card_manager/api_service/internal/api/dto"
	"card_manager/api_service/internal/api/response"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/service"
	"github.com/gin-gonic/gin"
)

type CardProductHandler struct {
	service service.CardProductService
	log     *logger.Logger
}

func NewCardProductHandler(svc service.CardProductService, log *logger.Logger) *CardProductHandler {
	return &CardProductHandler{service: svc, log: log}
}

func (h *CardProductHandler) List(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	resp, err := h.service.List(c.Request.Context(), userID, storeID, c.Query("type"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *CardProductHandler) Create(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	var req dto.CreateCardProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, ierr.Validation(ierr.MsgBadRequest))
		return
	}
	resp, err := h.service.Create(c.Request.Context(), userID, storeID, req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *CardProductHandler) Delete(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	productID, err := service.ParseProductID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.Delete(c.Request.Context(), userID, storeID, productID); err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, nil)
}
