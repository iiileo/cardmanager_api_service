package api

import (
	"context"
	"net/http"

	"card_manager/api_service/internal/api/middleware"
	v1 "card_manager/api_service/internal/api/v1"
	"card_manager/api_service/internal/config"
	"card_manager/api_service/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(
	cfg *config.Config,
	log *logger.Logger,
	hello *v1.HelloHandler,
	health *v1.HealthHandler,
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

	apiV1 := engine.Group("/v1")
	{
		apiV1.GET("/hello", hello.Hello)
		apiV1.POST("/hello", hello.HelloPost)
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
