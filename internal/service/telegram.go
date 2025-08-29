package service

import (
	"context"
	"fmt"
	"post-analyzer/internal/client/telegram/bot"
	"post-analyzer/internal/client/telegram/user"
	"post-analyzer/internal/entity"
	"regexp"
	"strings"

	"github.com/go-telegram/bot/models"
	"github.com/gotd/td/tg"
)

type telegramProvider interface {
	sendMessage(context.Context, *entity.Subscription, string) (*models.Message, error)
	getNewChannelPosts(context.Context, *entity.Subscription) ([]*tg.Message, error)
	getChatInfo(context.Context, *entity.Subscription) (*models.ChatFullInfo, error)
	validateChannel(context.Context, string) (bool, error)
	extractUsernameFromIdentificator(string) (string, error)
}

type telegramService struct {
	tgcUser *user.TelegramUserClient
	tgcBot  *bot.TelegramBotClient
}

func NewTelegramProvider(tgcUser *user.TelegramUserClient, tgcBot *bot.TelegramBotClient) telegramProvider {
	return &telegramService{
		tgcUser: tgcUser,
		tgcBot:  tgcBot,
	}
}

func (t *telegramService) sendMessage(ctx context.Context, sub *entity.Subscription, text string) (*models.Message, error) {

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	msg, err := t.tgcBot.SendMessage(ctx, sub.ChatID, text)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (t *telegramService) getNewChannelPosts(ctx context.Context, sub *entity.Subscription) ([]*tg.Message, error) {

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	posts, err := t.tgcUser.GetNewChannelPosts(ctx, sub.ChannelUsername, int(sub.LastCheckedPostID))
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (t *telegramService) getChatInfo(ctx context.Context, sub *entity.Subscription) (*models.ChatFullInfo, error) {

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	info, err := t.tgcBot.GetChatInfo(ctx, sub.ChannelUsername)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (t *telegramService) validateChannel(ctx context.Context, identificator string) (bool, error) {

	username, err := t.extractUsernameFromIdentificator(identificator)
	if err != nil {
		return false, fmt.Errorf("incorrect channel identificator: %w", err)
	}

	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	chat, err := t.tgcBot.GetChatInfo(ctx, username)
	if err != nil {
		return false, fmt.Errorf("no public channels with username \"%s\"", username)
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
		return "", fmt.Errorf("username too short")
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(identificator) {
		return "", fmt.Errorf("invalid characters in username")
	}

	return identificator, nil
}
