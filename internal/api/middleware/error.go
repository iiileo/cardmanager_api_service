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
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		if appErr, ok := ierr.AsAppError(err); ok {
			log.Error(c.Request.Context(), "request failed", "error", appErr, "path", c.Request.URL.Path)
			c.JSON(appErr.HTTPStatus(), appErr.ToResponse())
			return
		}

		log.Error(c.Request.Context(), "unhandled error", "error", err, "path", c.Request.URL.Path)
		c.JSON(http.StatusInternalServerError, ierr.ErrorResponse{
			Error:   ierr.ErrInternal.Error(),
			Message: "internal server error",
		})
	}
}
