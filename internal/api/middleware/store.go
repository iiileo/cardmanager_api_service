package middleware

import (
	"strconv"
	"strings"

	"card_manager/api_service/internal/api/response"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/service"
	"github.com/gin-gonic/gin"
)

const (
	HeaderStoreID = "X-Store-Id"
	CtxStoreID    = "store_id"
	CtxStoreRole  = "store_role"
)

func RequireStore(storeSvc service.StoreService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.Get("user_id")
		if !ok {
			response.Fail(c, ierr.Unauthorized(""))
			c.Abort()
			return
		}
		uid, _ := userID.(int64)

		raw := strings.TrimSpace(c.GetHeader(HeaderStoreID))
		if raw == "" {
			response.Fail(c, ierr.Validation("请选择门店"))
			c.Abort()
			return
		}
		storeID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || storeID <= 0 {
			response.Fail(c, ierr.Validation("门店 ID 不正确"))
			c.Abort()
			return
		}

		m, err := storeSvc.RequireActiveMember(c.Request.Context(), uid, storeID)
		if err != nil {
			response.Fail(c, err)
			c.Abort()
			return
		}
		c.Set(CtxStoreID, storeID)
		c.Set(CtxStoreRole, m.Role)
		c.Next()
	}
}

func RequireStoreOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(CtxStoreRole)
		if role != "owner" {
			response.Fail(c, ierr.Forbidden("仅门店老板可操作"))
			c.Abort()
			return
		}
		c.Next()
	}
}
