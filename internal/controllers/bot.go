package controllers

import (
	"context"
	"errors"
	"log"
	"strings"

	"post-analyzer/internal/domain/dto"
	"post-analyzer/internal/domain/presenter"
	"post-analyzer/internal/usecase"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type BotController struct {
	uc usecase.UseCase
}

func NewBotController(uc usecase.UseCase) *BotController {
	return &BotController{uc: uc}
}

func (bc BotController) Reply(ctx context.Context, b *bot.Bot, chatID int64, text string) error {

	_, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatID,
		Text:   text,
	})
	return err
}

func (bc BotController) StartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

	greetings := "Привет! Бот готов к работе!"

	err := bc.Reply(ctx, b, update.Message.From.ID, greetings)
	if err != nil {
		log.Printf("StartHandler: Failed to send message to chat: %v\n", err)
	}
}

func (bc BotController) MonitorHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

	mr := &dto.MonitorRequest{
		ChatID:  update.Message.Chat.ID,
		Message: strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/monitor")),
	}

	err := bc.uc.MonitorChannel(ctx, mr)

	if err != nil {

		failMessage := "Канал не был добавлен в систему мониторинга!\n"

		var presentedError *presenter.PresentedError
		if errors.As(err, &presentedError) {
			failMessage += err.Error()
		}

		err := bc.Reply(ctx, b, mr.ChatID, failMessage)
		if err != nil {
			log.Printf("MonitorHandler: Failed to send message to chat: %v", err)
		}

		return
	}

	successMessage := "Успех! Канал успешно добавлен в систему мониторинга!"

	err = bc.Reply(ctx, b, mr.ChatID, successMessage)
	if err != nil {
		log.Printf("MonitorHandler: Failed to send message to chat: %v", err)
	}
}
