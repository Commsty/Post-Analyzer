package notifier

import (
	"context"
	"errors"
	"fmt"
	"post-analyzer/internal/adapters/telegram/bot"
)

var (
	ErrTextNotification = errors.New("sending notification failed")
)

type Notifier interface {
	NotifyWithText(ctx context.Context, id int64, notification string) error
}

type botNotifier struct {
	client *bot.TelegramBotClient
}

func NewNotifier(client *bot.TelegramBotClient) *botNotifier {
	return &botNotifier{
		client: client,
	}
}

func (b botNotifier) NotifyWithText(ctx context.Context, id int64, notification string) error {

	if _, err := b.client.SendTextMessage(ctx, id, notification); err != nil {
		return fmt.Errorf("%w: %s", ErrTextNotification, err)
	}

	return nil
}
