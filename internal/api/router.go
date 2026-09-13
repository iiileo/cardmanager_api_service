package api

import (
	"context"
	"net/http"

	"card_manager/api_service/internal/api/middleware"
	v1 "card_manager/api_service/internal/api/v1"
	"card_manager/api_service/internal/auth"
	"card_manager/api_service/internal/config"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(
	cfg *config.Config,
	log *logger.Logger,
	tm *auth.TokenManager,
	health *v1.HealthHandler,
	authHandler *v1.AuthHandler,
	storeHandler *v1.StoreHandler,
	bizTypeHandler *v1.BizTypeHandler,
	cardProductHandler *v1.CardProductHandler,
	memberHandler *v1.MemberHandler,
	cardHandler *v1.CardHandler,
	ledgerHandler *v1.LedgerHandler,
	dashboardHandler *v1.DashboardHandler,
	storeSvc service.StoreService,
) *Router {
	if cfg.Server.Mode == "local" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.ErrorHandler(log))

	// System
	engine.GET("/healthz", health.Health)

	apiV1 := engine.Group("/api/v1")
	{
		// Auth（部分公开）
		authGroup := apiV1.Group("/auth")
		{
			authGroup.POST("/sms/send", authHandler.SendSMS)
			authGroup.POST("/login/sms", authHandler.LoginSMS)
			authGroup.POST("/token/refresh", authHandler.Refresh)
			authGroup.POST("/logout", authHandler.Logout)

			me := authGroup.Group("")
			me.Use(middleware.RequireAuth(tm))
			{
				me.GET("/me", authHandler.Me)
				me.PATCH("/me", authHandler.UpdateMe)
			}
		}

		authed := apiV1.Group("")
		authed.Use(middleware.RequireAuth(tm))
		{
			// BizTypes
			authed.GET("/biz-types", bizTypeHandler.List)

			// Stores
			stores := authed.Group("/stores")
			{
				stores.GET("", storeHandler.List)
				stores.POST("", storeHandler.Create)
				stores.GET("/invite/:code", storeHandler.PreviewInvite)
				stores.POST("/join", storeHandler.Join)
				stores.GET("/:id", storeHandler.Get)
				stores.PATCH("/:id", storeHandler.Update)
				stores.GET("/:id/invite-code", storeHandler.GetInviteCode)
				stores.POST("/:id/invite-code/refresh", storeHandler.RefreshInviteCode)
			}

			// 以下均需 X-Store-Id
			storeScoped := authed.Group("")
			storeScoped.Use(middleware.RequireStore(storeSvc))
			{
				// 首页统计
				storeScoped.GET("/home/stats", dashboardHandler.HomeStats)

				// Staff
				staff := storeScoped.Group("/staff")
				{
					staff.GET("", storeHandler.ListStaff)
					staff.GET("/applications", storeHandler.ListApplications)
					staff.POST("/applications/:id/approve", middleware.RequireStoreOwner(), storeHandler.Approve)
					staff.POST("/applications/:id/reject", middleware.RequireStoreOwner(), storeHandler.Reject)
				}

				// CardProducts 卡种
				products := storeScoped.Group("/card-products")
				{
					products.GET("", cardProductHandler.List)
					products.POST("", middleware.RequireStoreOwner(), cardProductHandler.Create)
					products.DELETE("/:id", middleware.RequireStoreOwner(), cardProductHandler.Delete)
				}

				// Members 会员
				members := storeScoped.Group("/members")
				{
					members.GET("", memberHandler.List)
					members.POST("/cards", memberHandler.OpenCard)
					members.GET("/:id", memberHandler.Get)
					members.GET("/:id/cards", cardHandler.ListByMember)
				}

				// Cards 持卡 / 充值 / 扣除
				cards := storeScoped.Group("/cards")
				{
					cards.GET("/:id", cardHandler.Get)
					cards.POST("/:id/recharge", cardHandler.Recharge)
					cards.POST("/:id/consume", cardHandler.Consume)
				}

				// Ledger 流水（全量 / 混合列表）
				ledger := storeScoped.Group("/ledger")
				{
					ledger.GET("", ledgerHandler.List)
					ledger.GET("/stats", ledgerHandler.StatsTxns)
					ledger.GET("/:id", ledgerHandler.Get)
				}

				// 充值记录
				recharges := storeScoped.Group("/recharges")
				{
					recharges.GET("", ledgerHandler.ListRecharges)
					recharges.GET("/stats", ledgerHandler.StatsRecharges)
					recharges.GET("/:id", ledgerHandler.GetRecharge)
				}

				// 消费记录（含套餐 items 明细）
				consumes := storeScoped.Group("/consumes")
				{
					consumes.GET("", ledgerHandler.ListConsumes)
					consumes.GET("/stats", ledgerHandler.StatsConsumes)
					consumes.GET("/:id", ledgerHandler.GetConsume)
				}
			}
		}
	}

	return &Router{engine: engine}
}

func (r *Router) Engine() *gin.Engine {
	return r.engine
}

func StartServer(lc fx.Lifecycle, cfg *config.Config, router *Router, log *logger.Logger) {
	srv := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: router.Engine(),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info(ctx, "starting http server", "addr", cfg.Server.Address)
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Error(context.Background(), "http server failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info(ctx, "stopping http server")
			return srv.Shutdown(ctx)
		},
	})
}
