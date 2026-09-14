package dto

// PeriodStatsResponse 某时间范围内的统计。
type PeriodStatsResponse struct {
	Recharge   int `json:"recharge"`
	Consume    int `json:"consume"`
	NewMembers int `json:"new_members"`
	TxnCount   int `json:"txn_count"`
}

// HomeStatsResponse 首页统计。
// 金额单位：分。时间边界按服务器本地时区。
// today_* 为今日；month_* 为本月（默认当月 1 日至今，传 month=YYYY-MM 可查指定月）。
type HomeStatsResponse struct {
	TodayRecharge   int `json:"today_recharge"`
	TodayConsume    int `json:"today_consume"`
	TodayNewMembers int `json:"today_new_members"`
	TodayTxnCount   int `json:"today_txn_count"`

	MonthRecharge   int `json:"month_recharge"`
	MonthConsume    int `json:"month_consume"`
	MonthNewMembers int `json:"month_new_members"`
	MonthTxnCount   int `json:"month_txn_count"`

	StoreBalance int `json:"store_balance"`
}
