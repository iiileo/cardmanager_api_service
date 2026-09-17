package dto

// StatsOverviewSummary 经营数据四格摘要（对齐 App「经营数据」Tab）。
type StatsOverviewSummary struct {
	RechargeAmount     int  `json:"recharge_amount"`
	RechargeChangePct  *int `json:"recharge_change_pct,omitempty"`
	ConsumeAmount      int  `json:"consume_amount"`
	ConsumeChangePct   *int `json:"consume_change_pct,omitempty"`
	NewOpens           int  `json:"new_opens"`
	NewOpensDelta      int  `json:"new_opens_delta"`
	RepeatCustomers    int  `json:"repeat_customers"`
	ActiveRatePct      int  `json:"active_rate_pct"`
}

// StatsDailyRechargePoint 每日充值（柱状图）；date 为 YYYY-MM-DD（本地自然日）。
type StatsDailyRechargePoint struct {
	Date   string `json:"date"`
	Amount int    `json:"amount"`
}

// StatsCardTypeMixItem 卡种占比。
type StatsCardTypeMixItem struct {
	CardType string `json:"card_type"`
	Count    int    `json:"count"`
	Percent  int    `json:"percent"`
}

// StatsCardTypeMix 卡种占比（当前在店有效持卡快照）。
type StatsCardTypeMix struct {
	TotalHolders int                    `json:"total_holders"`
	Items        []*StatsCardTypeMixItem `json:"items"`
}

// StatsRechargeRankItem 储值贡献榜条目。
type StatsRechargeRankItem struct {
	Rank     int    `json:"rank"`
	MemberID string `json:"member_id"`
	Name     string `json:"name"`
	Amount   int    `json:"amount"`
}

// StatsOverviewResponse GET /api/v1/stats/overview
type StatsOverviewResponse struct {
	RangeDays     int                        `json:"range_days"`
	From          string                     `json:"from"`
	To            string                     `json:"to"`
	Summary       StatsOverviewSummary       `json:"summary"`
	DailyRecharge []*StatsDailyRechargePoint `json:"daily_recharge"`
	CardTypeMix   StatsCardTypeMix           `json:"card_type_mix"`
	RechargeRank  []*StatsRechargeRankItem   `json:"recharge_rank"`
}
