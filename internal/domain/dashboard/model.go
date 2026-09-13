package dashboard

import (
	"context"
	"time"
)

// HomeStats 门店首页统计。
type HomeStats struct {
	TodayRecharge    int // 今日充值金额（分，正数）
	TodayConsume     int // 今日储值消费金额（分，正数）
	TodayNewMembers  int // 今日新开会员数
	StoreBalance     int // 在店储值卡余额合计（分）
	TodayTxnCount    int // 今日笔数（充值 + 各类消费，不含开卡）
}

type Repository interface {
	HomeStats(ctx context.Context, storeID int64, from, to time.Time) (*HomeStats, error)
}
