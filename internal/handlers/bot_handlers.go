package handlers

import (
	"context"
	"post-analyzer/internal/service"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type BotHandler struct {
	s service.MonitoringService
}

func NewBotHandler(service service.MonitoringService) *BotHandler {
	return &BotHandler{s: service}
}

func (h *BotHandler) StartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Отправь сюда ссылку на канал, за которым хочешь следить!",
	})
}

func (h *BotHandler) LinkHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	link := update.Message.Text

	err := h.s.AddNewChannel(ctx, link)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Неправильная ссылка на канал\n\n" + err.Error(),
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Успех! Канал добавлен в систему мониторинга.",
	})

}
