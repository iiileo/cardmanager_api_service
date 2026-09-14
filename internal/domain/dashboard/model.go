package dashboard

import (
	"context"
	"time"
)

// PeriodStats 某时间范围内的流水与会员统计。
type PeriodStats struct {
	Recharge   int // 充值金额（分，正数）
	Consume    int // 储值消费金额（分，正数）
	NewMembers int // 新开会员数
	TxnCount   int // 笔数（充值 + 各类消费，不含开卡）
}

// HomeStats 门店首页统计。
type HomeStats struct {
	Today        PeriodStats
	Month        PeriodStats
	StoreBalance int // 在店储值卡余额合计（分，实时快照）
}

type Repository interface {
	HomeStats(ctx context.Context, storeID int64, todayFrom, todayTo, monthFrom, monthTo time.Time) (*HomeStats, error)
}
