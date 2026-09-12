package response

import (
	"net/http"

	ierr "card_manager/api_service/internal/errors"
	"github.com/gin-gonic/gin"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, ierr.Envelope{
		Code:    0,
		Message: ierr.MsgOK,
		Data:    data,
	})
}

func Fail(c *gin.Context, err error) {
	if appErr, ok := ierr.AsAppError(err); ok {
		c.JSON(appErr.HTTPStatus(), appErr.ToEnvelope())
		return
	}
	c.JSON(http.StatusInternalServerError, ierr.Envelope{
		Code:    50000,
		Message: ierr.MsgInternal,
		Data:    nil,
	})
}
