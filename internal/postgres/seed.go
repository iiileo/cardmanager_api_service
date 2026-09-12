package postgres

import (
	"context"

	domainbiz "card_manager/api_service/internal/domain/biztype"
	"card_manager/api_service/internal/logger"
	"go.uber.org/fx"
)

// SeedDefaults 在 schema migrate 之后写入字典默认数据。
func SeedDefaults(lc fx.Lifecycle, bizTypes domainbiz.Repository, log *logger.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := bizTypes.EnsureDefaults(ctx); err != nil {
				return err
			}
			log.Info(ctx, "biz_types defaults ensured")
			return nil
		},
	})
}
