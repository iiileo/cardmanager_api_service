package dto

type NotifyChannelsResponse struct {
	Wechat bool `json:"wechat"`
	App    bool `json:"app"`
}

type NotifySettingResponse struct {
	Event      string                  `json:"event"`
	Enabled    bool                    `json:"enabled"`
	NotifyBoss bool                    `json:"notify_boss"`
	Channels   *NotifyChannelsResponse `json:"channels"`
}

type NotifySettingListResponse struct {
	List []*NotifySettingResponse `json:"list"`
}

type UpdateNotifySettingRequest struct {
	Enabled    bool `json:"enabled"`
	NotifyBoss bool `json:"notify_boss"`
	Wechat     bool `json:"wechat"`
	App        bool `json:"app"`
}
