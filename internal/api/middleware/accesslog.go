package middleware

import (
	"card_manager/api_service/internal/accesslog"

	"github.com/gin-gonic/gin"
)

// AccessLog 记录请求/响应与 SQL（见 internal/accesslog）。
func AccessLog(store *accesslog.Store) gin.HandlerFunc {
	return accesslog.Middleware(store)
}
