package service

import (
	"context"
	"fmt"
	tgBot "post-analyzer/internal/client/telegram/bot"
	"post-analyzer/internal/client/telegram/user"
	"post-analyzer/internal/entity"
	"post-analyzer/internal/repository"
)

type MonitoringService interface {
	AddNewChannel(context.Context, string) error
	GetAllChannels() *[]string
	DeleteChannelByUsername(context.Context, string) error
}

type monitoringService struct {
	r repository.ChannelStorageRepository

	tgcBot  tgBot.TelegramBotClient
	tgcUser user.TelegramUserClient
}

func NewMonitoringService(repo repository.ChannelStorageRepository,
	tgBotClient tgBot.TelegramBotClient, tgUserClient user.TelegramUserClient) MonitoringService {

	return &monitoringService{
		r:       repo,
		tgcBot:  tgBotClient,
		tgcUser: tgUserClient,
	}
}

func (m *monitoringService) AddNewChannel(ctx context.Context, link string) error {

	username, err := m.extractUsernameFromLink(link)
	if err != nil {
		return fmt.Errorf("Incorrect username: %w", err)
	}

	flag, err := m.isPublicChannel(ctx, username)
	if err != nil {
		return err
	}

	if !flag {
		return fmt.Errorf("This link does not point to a public channel")
	}

	chanID, _ := m.tgcBot.GetChatInfo(ctx, username)

	err = m.r.AddNewChannel(entity.ChannelInfo{
		ChannelID:         chanID.ID,
		ChannelUsername:   username,
		LastCheckedPostID: -1})

	if err != nil {
		return fmt.Errorf("Failed to add new channel: %w", err)
	}

	return nil
}

func (m *monitoringService) GetAllChannels() *[]string {

	infos := m.r.GetAllChannels()

	var channels []string
	for _, c := range infos {
		channels = append(channels, "@"+c.ChannelUsername)
	}

	return &channels
}

func (m *monitoringService) DeleteChannelByUsername(ctx context.Context, username string) error {

	chat, err := m.tgcBot.GetChatInfo(ctx, username)

	if err != nil {
		return err
	}

	return m.r.DeleteChannelByID(chat.ID)
}
