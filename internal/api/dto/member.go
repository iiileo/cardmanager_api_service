package dto

type MemberBrief struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type MemberListItem struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type MemberListResponse struct {
	List  []*MemberListItem `json:"list"`
	Total int               `json:"total"`
}

type MemberDetailResponse struct {
	ID    string              `json:"id"`
	Name  string              `json:"name"`
	Phone string              `json:"phone"`
	Cards []*MemberCardBrief  `json:"cards"`
}

type OpenCardRequest struct {
	MemberID  *string `json:"member_id"`
	Name      string  `json:"name"`
	Phone     string  `json:"phone"`
	ProductID string  `json:"product_id" binding:"required"`
	Source    string  `json:"source"`
}

type OpenCardResponse struct {
	Member *MemberBrief       `json:"member"`
	Card   *MemberCardDetail  `json:"card"`
	LedgerID string           `json:"ledger_id"`
}
