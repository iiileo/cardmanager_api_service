package v1

import (
	"strings"

	"card_manager/api_service/internal/api/response"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/service"
	"github.com/gin-gonic/gin"
)

type LedgerHandler struct {
	service service.LedgerService
	log     *logger.Logger
}

func NewLedgerHandler(svc service.LedgerService, log *logger.Logger) *LedgerHandler {
	return &LedgerHandler{service: svc, log: log}
}

func (h *LedgerHandler) List(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	page, pageSize := pageFromQuery(c)
	resp, err := h.service.List(
		c.Request.Context(), userID, storeID,
		service.LedgerListQuery{
			Kind:         c.Query("kind"),
			Type:         c.Query("type"),
			CardType:     c.Query("card_type"),
			MemberID:     c.Query("member_id"),
			CardID:       c.Query("card_id"),
			From:         c.Query("from"),
			To:           c.Query("to"),
			Page:         page,
			PageSize:     pageSize,
			IncludeStats: queryBool(c.Query("include_stats")),
		},
	)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *LedgerHandler) Get(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	ledgerID, err := service.ParseLedgerID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	resp, err := h.service.Get(c.Request.Context(), userID, storeID, ledgerID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *LedgerHandler) ListRecharges(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	resp, err := h.service.ListRecharges(c.Request.Context(), userID, storeID, listQueryFrom(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *LedgerHandler) ListOpens(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	resp, err := h.service.ListOpens(c.Request.Context(), userID, storeID, listQueryFrom(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *LedgerHandler) GetRecharge(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	id, err := service.ParseLedgerID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	resp, err := h.service.GetRecharge(c.Request.Context(), userID, storeID, id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *LedgerHandler) GetOpen(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	id, err := service.ParseLedgerID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	resp, err := h.service.GetOpen(c.Request.Context(), userID, storeID, id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *LedgerHandler) StatsTxns(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	q := listQueryFrom(c)
	if q.Kind == "" {
		q.Kind = "txn"
	}
	resp, err := h.service.StatsTxns(c.Request.Context(), userID, storeID, q)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *LedgerHandler) StatsRecharges(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	resp, err := h.service.StatsRecharges(c.Request.Context(), userID, storeID, listQueryFrom(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *LedgerHandler) StatsOpens(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	resp, err := h.service.StatsOpens(c.Request.Context(), userID, storeID, listQueryFrom(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *LedgerHandler) ListConsumes(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	resp, err := h.service.ListConsumes(c.Request.Context(), userID, storeID, listQueryFrom(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *LedgerHandler) GetConsume(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	id, err := service.ParseLedgerID(c.Param("id"))
	if err != nil {
		response.Fail(c, err)
		return
	}
	resp, err := h.service.GetConsume(c.Request.Context(), userID, storeID, id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func (h *LedgerHandler) StatsConsumes(c *gin.Context) {
	userID, storeID, ok := userAndStore(c)
	if !ok {
		return
	}
	resp, err := h.service.StatsConsumes(c.Request.Context(), userID, storeID, listQueryFrom(c))
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.OK(c, resp)
}

func listQueryFrom(c *gin.Context) service.LedgerListQuery {
	page, pageSize := pageFromQuery(c)
	return service.LedgerListQuery{
		Kind:         c.Query("kind"),
		Type:         c.Query("type"),
		CardType:     c.Query("card_type"),
		MemberID:     c.Query("member_id"),
		CardID:       c.Query("card_id"),
		From:         c.Query("from"),
		To:           c.Query("to"),
		Page:         page,
		PageSize:     pageSize,
		IncludeStats: queryBool(c.Query("include_stats")),
	}
}

func queryBool(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
