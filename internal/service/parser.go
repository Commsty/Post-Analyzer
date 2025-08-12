package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

func (m *monitoringService) extractUsernameFromLink(link string) (string, error) {
	link = strings.TrimPrefix(link, "https://")
	link = strings.TrimPrefix(link, "http://")
	link = strings.TrimPrefix(link, "t.me/")
	link = strings.TrimPrefix(link, "telegram.me/")
	link = strings.TrimPrefix(link, "@")
	link = strings.Split(link, "?")[0]
	link = strings.Split(link, "/")[0]

	if len(link) < 5 {
		return "", fmt.Errorf("Username too short")
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(link) {
		return "", fmt.Errorf("Invalid characters in username")
	}

	return link, nil
}

func (m *monitoringService) isPublicChannel(ctx context.Context, username string) (bool, error) {

	chat, err := m.tgcBot.GetChatInfo(ctx, username)
	if err != nil {
		return false, fmt.Errorf("No channels with username \"%s\"", username)
	}

	ch, us := chat.Type, chat.Username

	if ch != "channel" {
		return false, nil
	}
	if us == "" {
		return false, nil
	}

	return true, nil
}
