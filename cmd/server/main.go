package main

import (
	"card_manager/api_service/internal/accesslog"
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

func main() {
	fx.New(
		fx.Provide(
			config.NewConfig,
			accesslog.NewStore,
			logger.NewLogger,
			postgres.NewEntClient,
			auth.NewTokenManager,

			repository.NewUserRepository,
			repository.NewRefreshTokenRepository,
			repository.NewSmsCodeRepository,
			repository.NewStoreRepository,
			repository.NewStoreMemberRepository,
			repository.NewBizTypeRepository,
			repository.NewCardProductRepository,
			repository.NewMemberRepository,
			repository.NewMemberCardRepository,
			repository.NewLedgerRepository,
			repository.NewCardTxnRepository,
			repository.NewDashboardRepository,
			repository.NewNotifySettingRepository,

			service.NewAuthService,
			service.NewBizTypeService,
			service.NewStoreService,
			service.NewCardProductService,
			service.NewMemberService,
			service.NewCardService,
			service.NewLedgerService,
			service.NewDashboardService,
			service.NewNotifySettingService,

			v1.NewHealthHandler,
			v1.NewAuthHandler,
			v1.NewStoreHandler,
			v1.NewBizTypeHandler,
			v1.NewCardProductHandler,
			v1.NewMemberHandler,
			v1.NewCardHandler,
			v1.NewLedgerHandler,
			v1.NewDashboardHandler,
			v1.NewNotifySettingHandler,
			api.NewRouter,
		),
		fx.Invoke(postgres.SeedDefaults),
		fx.Invoke(api.StartServer),
	).Run()
}
