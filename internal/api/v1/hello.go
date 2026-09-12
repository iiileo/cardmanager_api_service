package v1

import (
	"net/http"

	"card_manager/api_service/internal/api/dto"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/service"
	"github.com/gin-gonic/gin"
)

type HelloHandler struct {
	service service.HelloService
	log     *logger.Logger
}

func NewHelloHandler(svc service.HelloService, log *logger.Logger) *HelloHandler {
	return &HelloHandler{service: svc, log: log}
}

// Hello godoc
// @Summary Hello world
// @Description Returns a greeting message
// @Tags Hello
// @Accept json
// @Produce json
// @Param name query string false "Name to greet (defaults via body if POST)"
// @Param body body dto.HelloRequest false "Hello payload"
// @Success 200 {object} dto.HelloResponse
// @Failure 400 {object} ierr.ErrorResponse
// @Router /hello [get]
func (h *HelloHandler) Hello(c *gin.Context) {
	req := dto.HelloRequest{Name: c.Query("name")}
	resp, err := h.service.SayHello(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// HelloPost godoc
// @Summary Hello world (POST)
// @Tags Hello
// @Accept json
// @Produce json
// @Param body body dto.HelloRequest true "Hello payload"
// @Success 200 {object} dto.HelloResponse
// @Failure 400 {object} ierr.ErrorResponse
// @Router /hello [post]
func (h *HelloHandler) HelloPost(c *gin.Context) {
	var req dto.HelloRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(ierr.WithError(err).
			WithHint("JSON body must include name").
			Mark(ierr.ErrValidation))
		return
	}

	resp, err := h.service.SayHello(c.Request.Context(), req)
	if err != nil {
		c.Error(err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
