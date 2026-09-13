package dto

type OperatorBrief struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
}

type LedgerItemResponse struct {
	ID          string `json:"id"`
	ItemID      string `json:"item_id"`
	Name        string `json:"name"`
	Times       int    `json:"times"`
	TimesAfter  int    `json:"times_after"`
}

type LedgerEntryResponse struct {
	ID           string                `json:"id"`
	Type         string                `json:"type"`
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
