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

	engine.GET("/healthz", health.Health)

	apiV1 := engine.Group("/api/v1")
	{
		authGroup := apiV1.Group("/auth")
		{
			authGroup.POST("/sms/send", authHandler.SendSMS)
			authGroup.POST("/login/sms", authHandler.LoginSMS)
			authGroup.POST("/token/refresh", authHandler.Refresh)
			authGroup.POST("/logout", authHandler.Logout)

			secured := authGroup.Group("")
			secured.Use(middleware.RequireAuth(tm))
			{
				secured.GET("/me", authHandler.Me)
				secured.PATCH("/me", authHandler.UpdateMe)
			}
		}

		authed := apiV1.Group("")
		authed.Use(middleware.RequireAuth(tm))
		{
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

			staff := authed.Group("/staff")
			staff.Use(middleware.RequireStore(storeSvc))
			{
				staff.GET("", storeHandler.ListStaff)
				staff.GET("/applications", storeHandler.ListApplications)
				staff.POST("/applications/:id/approve", middleware.RequireStoreOwner(), storeHandler.Approve)
				staff.POST("/applications/:id/reject", middleware.RequireStoreOwner(), storeHandler.Reject)
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
