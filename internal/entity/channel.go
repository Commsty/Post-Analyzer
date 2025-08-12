package entity

type ChannelInfo struct {
	ChannelID         int64  `json:"channel_id"`
	ChannelUsername   string `json:"username"`
	LastCheckedPostID int64  `json:"last_id"`
}
