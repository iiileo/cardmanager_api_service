package dto

// HomeStatsResponse 首页统计。
// 金额单位：分。今日边界按服务器本地时区。
type HomeStatsResponse struct {
	TodayRecharge   int `json:"today_recharge"`    // 今日充值
	TodayConsume    int `json:"today_consume"`     // 今日消费（储值金额）
	TodayNewMembers int `json:"today_new_members"` // 今日新开会员
	StoreBalance    int `json:"store_balance"`     // 在店余额（储值卡合计）
	TodayTxnCount   int `json:"today_txn_count"`   // 今日笔数（充值+消费）
}
