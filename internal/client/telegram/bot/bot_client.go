package bot

import (
	"context"
	"fmt"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type TelegramBotClient struct {
	b *bot.Bot
}

func NewTelegramBotClient(bot *bot.Bot) *TelegramBotClient {
	return &TelegramBotClient{b: bot}
}

func (t *TelegramBotClient) SendMessage(ctx context.Context, chatID int64, text string) (*models.Message, error) {

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	msg, err := t.b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	return msg, nil
}

func (t *TelegramBotClient) GetChatInfo(ctx context.Context, username string) (*models.ChatFullInfo, error) {

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	chat, err := t.b.GetChat(ctx, &bot.GetChatParams{
		ChatID: "@" + username,
	})

	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	return chat, nil
}
