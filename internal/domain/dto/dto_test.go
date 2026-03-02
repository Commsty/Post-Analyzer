package dto

import (
	"errors"
	"strings"
	"testing"
)

func TestMonitorRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     MonitorRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid monitor request",
			request: MonitorRequest{
				ChatID:  12345,
				Message: "/monitor @testchannel",
			},
			expectError: false,
		},
		{
			name: "invalid chat id",
			request: MonitorRequest{
				ChatID:  0,
				Message: "/monitor @testchannel",
			},
			expectError: true,
			errorMsg:    "invalid ChatID",
		},
		{
			name: "empty message",
			request: MonitorRequest{
				ChatID:  12345,
				Message: "",
			},
			expectError: true,
			errorMsg:    "empty message",
		},
		{
			name: "message without monitor command",
			request: MonitorRequest{
				ChatID:  12345,
				Message: "/start",
			},
			expectError: true,
			errorMsg:    "not a monitor command",
		},
		{
			name: "monitor command without channel",
			request: MonitorRequest{
				ChatID:  12345,
				Message: "/monitor",
			},
			expectError: true,
			errorMsg:    "no channel specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.expectError && err == nil {
				t.Errorf("expected error '%s', got nil", tt.errorMsg)
			}

			if !tt.expectError && err != nil {
				t.Errorf("expected no error, got: %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if err.Error() != tt.errorMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestMonitorRequest_ExtractChannel(t *testing.T) {
	tests := []struct {
		name     string
		request  MonitorRequest
		expected string
	}{
		{
			name: "extract channel with @",
			request: MonitorRequest{
				Message: "/monitor @testchannel",
			},
			expected: "testchannel",
		},
		{
			name: "extract channel without @",
			request: MonitorRequest{
				Message: "/monitor testchannel",
			},
			expected: "testchannel",
		},
		{
			name: "extract channel with time",
			request: MonitorRequest{
				Message: "/monitor @testchannel 09:00",
			},
			expected: "testchannel",
		},
		{
			name: "empty message",
			request: MonitorRequest{
				Message: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.ExtractChannel()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestMonitorRequest_ExtractTime(t *testing.T) {
	tests := []struct {
		name     string
		request  MonitorRequest
		expected string
	}{
		{
			name: "extract time with channel",
			request: MonitorRequest{
				Message: "/monitor @testchannel 09:00",
			},
			expected: "09:00",
		},
		{
			name: "no time specified",
			request: MonitorRequest{
				Message: "/monitor @testchannel",
			},
			expected: "",
		},
		{
			name: "invalid time format",
			request: MonitorRequest{
				Message: "/monitor @testchannel 25:00",
			},
			expected: "25:00",
		},
		{
			name: "multiple time formats",
			request: MonitorRequest{
				Message: "/monitor @testchannel 18:30",
			},
			expected: "18:30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.ExtractTime()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestMonitorRequest_IsMonitorCommand(t *testing.T) {
	tests := []struct {
		name     string
		request  MonitorRequest
		expected bool
	}{
		{
			name: "valid monitor command",
			request: MonitorRequest{
				Message: "/monitor @testchannel",
			},
			expected: true,
		},
		{
			name: "monitor command with prefix",
			request: MonitorRequest{
				Message: "/monitor",
			},
			expected: true,
		},
		{
			name: "different command",
			request: MonitorRequest{
				Message: "/start",
			},
			expected: false,
		},
		{
			name: "empty message",
			request: MonitorRequest{
				Message: "",
			},
			expected: false,
		},
		{
			name: "monitor with extra spaces",
			request: MonitorRequest{
				Message: "  /monitor @testchannel  ",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.IsMonitorCommand()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func (mr *MonitorRequest) Validate() error {
	if mr.ChatID <= 0 {
		return errors.New("invalid ChatID")
	}
	if mr.Message == "" {
		return errors.New("empty message")
	}
	if !mr.IsMonitorCommand() {
		return errors.New("not a monitor command")
	}
	if mr.ExtractChannel() == "" {
		return errors.New("no channel specified")
	}
	return nil
}

func (mr *MonitorRequest) ExtractChannel() string {
	parts := strings.Fields(mr.Message)
	if len(parts) >= 2 {
		channel := parts[1]
		return strings.TrimPrefix(channel, "@")
	}
	return ""
}

func (mr *MonitorRequest) ExtractTime() string {
	parts := strings.Fields(mr.Message)
	if len(parts) >= 3 {
		return parts[2]
	}
	return ""
}

func (mr *MonitorRequest) IsMonitorCommand() bool {
	return strings.HasPrefix(strings.TrimSpace(mr.Message), "/monitor")
}
