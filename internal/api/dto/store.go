package dto

type CreateStoreRequest struct {
	Name      string  `json:"name" binding:"required"`
	City      string  `json:"city" binding:"required"`
	Address   *string `json:"address"`
	OpenTime  string  `json:"open_time" binding:"required"`
	CloseTime string  `json:"close_time" binding:"required"`
	BizType   *string `json:"biz_type"`
}

type UpdateStoreRequest struct {
	Name      *string `json:"name"`
	City      *string `json:"city"`
	Address   *string `json:"address"`
	OpenTime  *string `json:"open_time"`
	CloseTime *string `json:"close_time"`
	BizType   *string `json:"biz_type"`
}

type JoinStoreRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
	Nickname   string `json:"nickname"`
	Agreed     bool   `json:"agreed"`
}

type StoreListItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	Status       string `json:"status"`
	City         string `json:"city"`
	MembersCount int    `json:"members_count,omitempty"`
}

type StoreListResponse struct {
	List []*StoreListItem `json:"list"`
}

type StoreDetailResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	City         string  `json:"city"`
	Address      *string `json:"address,omitempty"`
	OpenTime     string  `json:"open_time"`
	CloseTime    string  `json:"close_time"`
	BizType      *string `json:"biz_type,omitempty"`
	InviteCode   string  `json:"invite_code,omitempty"`
	OwnerUserID  string  `json:"owner_user_id"`
	Role         string  `json:"role,omitempty"`
	Status       string  `json:"status,omitempty"`
	MembersCount int     `json:"members_count"`
}

type InvitePreviewResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	City string `json:"city"`
}

type InviteCodeResponse struct {
	InviteCode string `json:"invite_code"`
}

type StaffItem struct {
	ID          string  `json:"id"`
	UserID      string  `json:"user_id"`
	Role        string  `json:"role"`
	Status      string  `json:"status"`
	DisplayName *string `json:"display_name,omitempty"`
	JoinedAt    *string `json:"joined_at,omitempty"`
}

type StaffListResponse struct {
	List []*StaffItem `json:"list"`
}
