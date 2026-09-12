package v1

import (
	"net/http"

	"card_manager/api_service/internal/api/dto"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health godoc
// @Summary Health check
// @Tags System
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Router /healthz [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, dto.HealthResponse{Status: "ok"})
}
