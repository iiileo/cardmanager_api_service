package postgres

import (
	"context"

	domainbiz "card_manager/api_service/internal/domain/biztype"
	domainmember "card_manager/api_service/internal/domain/member"
	"card_manager/api_service/internal/logger"
	"go.uber.org/fx"
)

// SeedDefaults 在 schema migrate 之后写入字典默认数据，并补齐会员拼音检索字段。
func SeedDefaults(lc fx.Lifecycle, bizTypes domainbiz.Repository, members domainmember.Repository, log *logger.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := bizTypes.EnsureDefaults(ctx); err != nil {
				return err
			}
			log.Info(ctx, "biz_types defaults ensured")
			n, err := members.BackfillNamePinyin(ctx)
			if err != nil {
				return err
			}
			if n > 0 {
				log.Info(ctx, "member name pinyin backfilled", "count", n)
			}
			return nil
		},
	})
}
