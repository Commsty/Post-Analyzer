package handlers

import (
	"context"
	"log"
	"post-analyzer/internal/entity"
	"post-analyzer/internal/service"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type BotHandler struct {
	service service.ChannelService
}

func NewBotHandler(service service.ChannelService) *BotHandler {
	return &BotHandler{service: service}
}

func (h *BotHandler) StartHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

	sendCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	_, err := b.SendMessage(sendCtx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Привет! Бот готов к работе!",
	})
	if err != nil {
		log.Printf("StartHandler: Failed to send message to chat: %v\n", err)
	}
}

func (h *BotHandler) MonitorHandler(ctx context.Context, b *bot.Bot, update *models.Update) {

	monitorCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	chatID := update.Message.Chat.ID
	args := strings.Fields(strings.TrimSpace(strings.TrimPrefix(update.Message.Text, "/monitor")))

	if len(args) != 2 {
		_, err := b.SendMessage(monitorCtx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Некорректный формат команды.\nОжидается: /monitor {link/username} {time}",
		})
		if err != nil {
			log.Printf("MonitorHandler: Failed to send message to chat: %v", err)
		}
		return
	}

	link, timeString := args[0], args[1]
	chanInfo := &entity.ChannelInfo{
		ChatID:            chatID,
		LastCheckedPostID: -1,
	}

	err := h.service.AddChannel(monitorCtx, link, timeString, chanInfo)

	if err != nil {
		_, err := b.SendMessage(monitorCtx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Channel adding failure: " + err.Error(),
		})
		if err != nil {
			log.Printf("MonitorHandler: Failed to send message to chat: %v", err)
		}
		return
	}

	_, err = b.SendMessage(monitorCtx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Успех! Канал добавлен в систему мониторинга.",
	})
	if err != nil {
		log.Printf("MonitorHandler: Failed to send message to chat: %v", err)
	}

}
