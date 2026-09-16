package middleware

import (
	"net/http"

	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
	"github.com/gin-gonic/gin"
)

func ErrorHandler(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err
		if appErr, ok := ierr.AsAppError(err); ok {
			status := appErr.HTTPStatus()
			log.Error(c.Request.Context(), "request failed", "error", err, "path", c.Request.URL.Path)
			c.JSON(status, appErr.ToEnvelope())
			return
		}

		log.Error(c.Request.Context(), "unhandled error", "error", err, "path", c.Request.URL.Path)
		c.JSON(http.StatusInternalServerError, ierr.Envelope{
			Code:    50000,
			Message: ierr.MsgInternal,
			Data:    nil,
		})
	}
}
