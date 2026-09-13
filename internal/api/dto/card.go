package dto

type PackItemBalanceResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	RemainTimes int    `json:"remain_times"`
}

type MemberCardBrief struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Balance     *int   `json:"balance,omitempty"`
	RemainTimes *int   `json:"remain_times,omitempty"`
	ValidTo     *string `json:"valid_to,omitempty"`
	Status      string `json:"status"`
}

type MemberCardDetail struct {
	ID          string                     `json:"id"`
	MemberID    string                     `json:"member_id"`
	Type        string                     `json:"type"`
	Name        string                     `json:"name"`
	NameSnapshot string                    `json:"name_snapshot"`
	Balance     *int                       `json:"balance,omitempty"`
	RemainTimes *int                       `json:"remain_times,omitempty"`
	ValidTo     *string                    `json:"valid_to,omitempty"`
	Status      string                     `json:"status"`
	Items       []*PackItemBalanceResponse `json:"items,omitempty"`
	Member      *MemberBrief               `json:"member,omitempty"`
}

type MemberCardListResponse struct {
	List []*MemberCardDetail `json:"list"`
}

type RechargeRequest struct {
	Amount int `json:"amount" binding:"required"`
}

type ConsumePackItemRequest struct {
	ItemID string `json:"item_id" binding:"required"`
	Times  int    `json:"times"`
}

type ConsumeRequest struct {
	Amount *int                      `json:"amount"`
	Times  *int                      `json:"times"`
	ItemID *string                   `json:"item_id"` // 兼容单项目
	Items  []ConsumePackItemRequest  `json:"items"`   // 一次扣多项
	Remark string                    `json:"remark"`
}

type TxnResultResponse struct {
	CardID       string                     `json:"card_id"`
	BalanceAfter *int                       `json:"balance_after,omitempty"`
	TimesAfter   *int                       `json:"times_after,omitempty"`
	LedgerID     string                     `json:"ledger_id"`
	Items        []*PackItemBalanceResponse `json:"items,omitempty"`
}
