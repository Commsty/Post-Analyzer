package service

import (
	"context"
	"fmt"
	"post-analyzer/internal/client/telegram/bot"
	"post-analyzer/internal/client/telegram/user"
	"regexp"
	"strings"

	"github.com/go-telegram/bot/models"
	"github.com/gotd/td/tg"
)

type telegramProvider interface {
	getNewChannelPosts(context.Context, string, int) ([]*tg.Message, error)
	getChatInfo(context.Context, string) (*models.ChatFullInfo, error)
	validateChannel(context.Context, string) (bool, error)
	extractUsernameFromIdentificator(string) (string, error)
}

func NewTelegramProvider(tgcUser *user.TelegramUserClient, tgcBot *bot.TelegramBotClient) telegramProvider {
	return &telegramService{
		tgcUser: tgcUser,
		tgcBot:  tgcBot,
	}
}

type telegramService struct {
	tgcUser *user.TelegramUserClient
	tgcBot  *bot.TelegramBotClient
}

func (t *telegramService) getNewChannelPosts(ctx context.Context, username string, lastReadID int) ([]*tg.Message, error) {

	posts, err := t.tgcUser.GetNewChannelPosts(ctx, username, lastReadID)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (t *telegramService) getChatInfo(ctx context.Context, username string) (*models.ChatFullInfo, error) {

	info, err := t.tgcBot.GetChatInfo(ctx, username)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (t *telegramService) validateChannel(ctx context.Context, identificator string) (bool, error) {

	username, err := t.extractUsernameFromIdentificator(identificator)
	if err != nil {
		return false, fmt.Errorf("Incorrect channel identificator: %w", err)
	}

	chat, err := t.tgcBot.GetChatInfo(ctx, username)
	if err != nil {
		return false, fmt.Errorf("No public channels with username \"%s\"", username)
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

func (t *telegramService) extractUsernameFromIdentificator(identificator string) (string, error) {
	identificator = strings.TrimPrefix(identificator, "https://")
	identificator = strings.TrimPrefix(identificator, "http://")
	identificator = strings.TrimPrefix(identificator, "t.me/")
	identificator = strings.TrimPrefix(identificator, "telegram.me/")
	identificator = strings.TrimPrefix(identificator, "@")
	identificator = strings.Split(identificator, "?")[0]
	identificator = strings.Split(identificator, "/")[0]

	if len(identificator) < 5 {
		return "", fmt.Errorf("Username too short")
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(identificator) {
		return "", fmt.Errorf("Invalid characters in username")
	}

	return identificator, nil
}
