package main

import (
	"card_manager/api_service/internal/api"
	v1 "card_manager/api_service/internal/api/v1"
	"card_manager/api_service/internal/auth"
	"card_manager/api_service/internal/config"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/postgres"
	"card_manager/api_service/internal/repository"
	"card_manager/api_service/internal/service"
	"go.uber.org/fx"
)

// @title Card Manager API
// @version 1.0
// @description Card Manager API service
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	fx.New(
		fx.Provide(
			config.NewConfig,
			logger.NewLogger,
			postgres.NewEntClient,
			auth.NewTokenManager,

			repository.NewUserRepository,
			repository.NewRefreshTokenRepository,
			repository.NewSmsCodeRepository,

			service.NewAuthService,

			v1.NewHealthHandler,
			v1.NewAuthHandler,
			api.NewRouter,
		),
		fx.Invoke(api.StartServer),
	).Run()
}
