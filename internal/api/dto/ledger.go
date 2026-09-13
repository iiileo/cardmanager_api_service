package dto

type OperatorBrief struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
}

type LedgerItemResponse struct {
	ID         string `json:"id"`
	ItemID     string `json:"item_id"`
	Name       string `json:"name"`
	Times      int    `json:"times"`
	TimesAfter int    `json:"times_after"`
}

type LedgerEntryResponse struct {
	ID           string                `json:"id"`
	Type         string                `json:"type"`
	CardType     string                `json:"card_type,omitempty"`
	Amount       *int                  `json:"amount,omitempty"`
	Times        *int                  `json:"times,omitempty"`
	ItemName     *string               `json:"item_name,omitempty"`
	BalanceAfter *int                  `json:"balance_after,omitempty"`
	TimesAfter   *int                  `json:"times_after,omitempty"`
	Remark       *string               `json:"remark,omitempty"`
	CreatedAt    string                `json:"created_at"`
	Member       *MemberBrief          `json:"member,omitempty"`
	Card         *MemberCardBrief      `json:"card,omitempty"`
	Operator     *OperatorBrief        `json:"operator,omitempty"`
	Items        []*LedgerItemResponse `json:"items,omitempty"`
}

type LedgerListResponse struct {
	List  []*LedgerEntryResponse `json:"list"`
	Total int                    `json:"total"`
}

// RecordTypeStat 按流水类型 + 卡类型汇总，便于后续图表/报表。
type RecordTypeStat struct {
	Type     string `json:"type"`
	CardType string `json:"card_type,omitempty"`
	Count    int    `json:"count"`
	Amount   int    `json:"amount"` // 金额合计（正数）
	Times    int    `json:"times"`  // 次数合计（正数）
}

// PackItemStat 套餐项目消耗汇总（后续可做项目排行榜）。
type PackItemStat struct {
	ProductItemID string `json:"product_item_id"`
	Name          string `json:"name"`
	Times         int    `json:"times"`
	Count         int    `json:"count"`
}

type RecordStatsResponse struct {
	TotalCount  int              `json:"total_count"`
	TotalAmount int              `json:"total_amount"`
	TotalTimes  int              `json:"total_times"`
	ByType      []*RecordTypeStat `json:"by_type"`
	ByPackItem  []*PackItemStat   `json:"by_pack_item,omitempty"`
}
