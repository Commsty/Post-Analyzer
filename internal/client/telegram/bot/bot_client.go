package bot

import (
	"context"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type TelegramBotClient struct {
	b *bot.Bot
}

func NewTelegramBotClient(bot *bot.Bot) TelegramBotClient {
	return TelegramBotClient{b: bot}
}

func (t *TelegramBotClient) GetChatInfo(ctx context.Context, username string) (*models.ChatFullInfo, error) {

	chat, err := t.b.GetChat(ctx, &bot.GetChatParams{
		ChatID: "@" + username,
	})

	if err != nil {
		return &models.ChatFullInfo{}, err
	}

	return chat, nil
}
