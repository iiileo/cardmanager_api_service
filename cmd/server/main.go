package main

import (
	"card_manager/api_service/internal/api"
	v1 "card_manager/api_service/internal/api/v1"
	"card_manager/api_service/internal/config"
	"card_manager/api_service/internal/logger"
	"card_manager/api_service/internal/service"
	"go.uber.org/fx"
)

// @title Card Manager API
// @version 1.0
// @description Card Manager API service
// @BasePath /v1
// @schemes http https
func main() {
	fx.New(
		fx.Provide(
			config.NewConfig,
			logger.NewLogger,
			service.NewHelloService,
			v1.NewHelloHandler,
			v1.NewHealthHandler,
			api.NewRouter,
		),
		fx.Invoke(api.StartServer),
	).Run()
}
