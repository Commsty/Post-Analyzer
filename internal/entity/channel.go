package entity

type ChannelInfo struct {
	ChatID            int64  `json:"chat_id"`
	ChannelID         int64  `json:"channel_id"`
	ChannelUsername   string `json:"username"`
	LastCheckedPostID int64  `json:"last_id"`
	ScheduleID        int    `json:"schedule_id"`
}
