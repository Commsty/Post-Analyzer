package bot

import (
	"context"
	"errors"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

var (
	ErrAPICallFailed = errors.New("API call failed")

	ErrTimeLimit = errors.New("telegram bot req time limit reached")
)

type TelegramBotClient struct {
	b *bot.Bot
}

func NewTelegramBotClient(bot *bot.Bot) *TelegramBotClient {
	return &TelegramBotClient{
		b: bot,
	}
}

func (t TelegramBotClient) SendTextMessage(ctx context.Context, chatID int64, text string) (*models.Message, error) {

	if ctx.Err() != nil {
		return nil, ErrTimeLimit
	}

	msg, err := t.b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		return nil, ErrAPICallFailed
	}

	return msg, nil
}
