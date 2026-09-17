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
	StatsOverview(ctx context.Context, in StatsOverviewQuery) (*StatsOverview, error)
}

// StatsOverviewQuery 经营数据概览时间窗（含上一段等长对比窗）。
type StatsOverviewQuery struct {
	StoreID                          int64
	CurFrom, CurTo, PrevFrom, PrevTo time.Time
}

// StatsOverviewPeriod 单段时间内的核心指标。
type StatsOverviewPeriod struct {
	RechargeAmount int
	ConsumeAmount  int
	NewOpens       int
}

// DailyRechargePoint 某日充值合计。
type DailyRechargePoint struct {
	Date   time.Time
	Amount int
}

// CardTypeMixRow 持卡卡种分布（当前快照）。
type CardTypeMixRow struct {
	CardType string
	Count    int
}

// RechargeRankRow 时段内充值贡献。
type RechargeRankRow struct {
	MemberID int64
	Name     string
	Amount   int
}

// StatsOverview 经营数据 Tab 聚合结果。
type StatsOverview struct {
	Cur               StatsOverviewPeriod
	Prev              StatsOverviewPeriod
	RepeatCustomers   int
	ActiveMembers     int
	TotalCardHolders  int
	DailyRecharge     []DailyRechargePoint
	CardTypeMix       []CardTypeMixRow
	RechargeRank      []RechargeRankRow
}
