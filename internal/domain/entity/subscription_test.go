package entity

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestSubscription_Validation(t *testing.T) {
	tests := []struct {
		name         string
		subscription Subscription
		expectError  bool
		errorField   string
	}{
		{
			name: "valid subscription",
			subscription: Subscription{
				ChatID:            12345,
				ChannelID:         67890,
				ChannelUsername:   "@testchannel",
				LastCheckedPostID: 0,
				SendingTime:       "09:00",
				ScheduleID:        1,
			},
			expectError: false,
		},
		{
			name: "invalid chat id",
			subscription: Subscription{
				ChatID:            0,
				ChannelID:         67890,
				ChannelUsername:   "@testchannel",
				LastCheckedPostID: 0,
				SendingTime:       "09:00",
				ScheduleID:        1,
			},
			expectError: true,
			errorField:  "ChatID",
		},
		{
			name: "invalid channel id",
			subscription: Subscription{
				ChatID:            12345,
				ChannelID:         0,
				ChannelUsername:   "@testchannel",
				LastCheckedPostID: 0,
				SendingTime:       "09:00",
				ScheduleID:        1,
			},
			expectError: true,
			errorField:  "ChannelID",
		},
		{
			name: "empty channel username",
			subscription: Subscription{
				ChatID:            12345,
				ChannelID:         67890,
				ChannelUsername:   "",
				LastCheckedPostID: 0,
				SendingTime:       "09:00",
				ScheduleID:        1,
			},
			expectError: true,
			errorField:  "ChannelUsername",
		},
		{
			name: "invalid sending time format",
			subscription: Subscription{
				ChatID:            12345,
				ChannelID:         67890,
				ChannelUsername:   "@testchannel",
				LastCheckedPostID: 0,
				SendingTime:       "25:00",
				ScheduleID:        1,
			},
			expectError: true,
			errorField:  "SendingTime",
		},
		{
			name: "invalid sending time - wrong format",
			subscription: Subscription{
				ChatID:            12345,
				ChannelID:         67890,
				ChannelUsername:   "@testchannel",
				LastCheckedPostID: 0,
				SendingTime:       "9:00",
				ScheduleID:        1,
			},
			expectError: true,
			errorField:  "SendingTime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.subscription.Validate()

			if tt.expectError && err == nil {
				t.Errorf("expected error for field %s, got nil", tt.errorField)
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}
		})
	}
}

func TestSubscription_IsActive(t *testing.T) {
	tests := []struct {
		name         string
		subscription Subscription
		expected     bool
	}{
		{
			name: "active subscription with future time",
			subscription: Subscription{
				SendingTime: "23:59",
				ScheduleID:  1,
			},
			expected: true,
		},
		{
			name: "active subscription with morning time",
			subscription: Subscription{
				SendingTime: "09:00",
				ScheduleID:  1,
			},
			expected: true,
		},
		{
			name: "inactive subscription - no schedule id",
			subscription: Subscription{
				SendingTime: "09:00",
				ScheduleID:  0,
			},
			expected: false,
		},
		{
			name: "inactive subscription - no sending time",
			subscription: Subscription{
				SendingTime: "",
				ScheduleID:  1,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.subscription.IsActive()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSubscription_NextCheckTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name         string
		subscription Subscription
		expected     time.Time
	}{
		{
			name: "next check time for morning schedule",
			subscription: Subscription{
				SendingTime: "09:00",
			},
			expected: time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location()),
		},
		{
			name: "next check time for evening schedule",
			subscription: Subscription{
				SendingTime: "18:30",
			},
			expected: time.Date(now.Year(), now.Month(), now.Day(), 18, 30, 0, 0, now.Location()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.subscription.NextCheckTime()

			if result.Hour() != tt.expected.Hour() || result.Minute() != tt.expected.Minute() {
				t.Errorf("expected %02d:%02d, got %02d:%02d",
					tt.expected.Hour(), tt.expected.Minute(),
					result.Hour(), result.Minute())
			}
		})
	}
}

func (s *Subscription) Validate() error {
	if s.ChatID <= 0 {
		return errors.New("invalid ChatID")
	}
	if s.ChannelID <= 0 {
		return errors.New("invalid ChannelID")
	}
	if s.ChannelUsername == "" {
		return errors.New("empty ChannelUsername")
	}
	if s.SendingTime == "" {
		return errors.New("empty SendingTime")
	}
	if len(s.SendingTime) != 5 || s.SendingTime[2] != ':' {
		return errors.New("invalid SendingTime format")
	}
	hour, minute := 0, 0
	if _, err := fmt.Sscanf(s.SendingTime, "%d:%d", &hour, &minute); err != nil {
		return errors.New("invalid SendingTime format")
	}
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return errors.New("invalid SendingTime values")
	}
	return nil
}

func (s *Subscription) IsActive() bool {
	return s.SendingTime != "" && s.ScheduleID > 0
}

func (s *Subscription) NextCheckTime() time.Time {
	now := time.Now()
	hour, minute := 0, 0

	if _, err := fmt.Sscanf(s.SendingTime, "%d:%d", &hour, &minute); err == nil {
		return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	}

	return now
}
