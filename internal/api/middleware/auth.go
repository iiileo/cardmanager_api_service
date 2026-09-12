package middleware

import (
	"card_manager/api_service/internal/api/response"
	"card_manager/api_service/internal/auth"
	ierr "card_manager/api_service/internal/errors"
	"github.com/gin-gonic/gin"
)

func RequireAuth(tm *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := bearer(c)
		if raw == "" {
			response.Fail(c, ierr.Unauthorized(""))
			c.Abort()
			return
		}
		claims, err := tm.ParseAccess(raw)
		if err != nil {
			response.Fail(c, ierr.Unauthorized(ierr.MsgTokenInvalid))
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("phone", claims.Phone)
		c.Next()
	}
}

func bearer(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if len(h) < 8 {
		return ""
	}
	if !(h[0] == 'B' || h[0] == 'b') {
		return ""
	}
	const prefix = "Bearer "
	if len(h) <= len(prefix) {
		return ""
	}
	if h[:7] != "Bearer " && h[:7] != "bearer " {
		return ""
	}
	return h[7:]
}
