package entity

type Subscription struct {
	ChatID            int64
	ChannelID         int64
	ChannelUsername   string
	LastCheckedPostID int64
	SendingTime       string
	ScheduleID        int
}
