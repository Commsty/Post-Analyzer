package service

import (
	"context"
	"fmt"
	"post-analyzer/internal/entity"
	"post-analyzer/internal/repository"
)

type MonitoringService interface {
	AddNewChannel(context.Context, string) error
	GetAllChannels() []string
	DeleteChannelByUsername(context.Context, string) error
	AnalyzeChannel(username string) (string, error)
}

type monitoringService struct {
	repository repository.ChannelStorageRepository

	analysisService analysisProvider
	telegramService telegramProvider
}

func NewMonitoringService(repo repository.ChannelStorageRepository,
	as analysisProvider, ts telegramProvider) MonitoringService {

	return &monitoringService{
		repository:      repo,
		analysisService: as,
		telegramService: ts,
	}
}

func (m *monitoringService) AnalyzeChannel(username string) (string, error) {

	newPosts, err := m.telegramService.getNewChannelPosts(context.Background(), username, -1)
	if err != nil {
		return "", fmt.Errorf("Failed to get new posts: %w", err)
	}

	var postsText string
	for _, msg := range newPosts {
		postsText += msg.Message + "\n"
	}

	analysisResult, err := m.analysisService.AnalyzeTextWithBasicPrompt(context.Background(), postsText)
	if err != nil {
		return "", fmt.Errorf("Failed to analyse new posts: %w", err)
	}

	return analysisResult, nil
}

func (m *monitoringService) AddNewChannel(ctx context.Context, link string) error {

	flag, err := m.telegramService.validateChannel(ctx, link)
	if err != nil || !flag {
		return fmt.Errorf("Validation failed: %w", err)
	}

	username, _ := m.telegramService.extractUsernameFromIdentificator(link)

	chanID, err := m.telegramService.getChatInfo(ctx, username)

	err = m.repository.AddNewChannel(entity.ChannelInfo{
		ChannelID:         chanID.ID,
		ChannelUsername:   username,
		LastCheckedPostID: -1})

	if err != nil {
		return fmt.Errorf("Repository failure: %w", err)
	}

	return nil
}

func (m *monitoringService) GetAllChannels() []string {

	infos := m.repository.GetAllChannels()

	var channels = make([]string, 0, len(infos))
	for i, c := range infos {
		channels[i] = "@" + c.ChannelUsername
	}

	return channels
}

func (m *monitoringService) DeleteChannelByUsername(ctx context.Context, username string) error {

	chat, err := m.telegramService.getChatInfo(ctx, username)

	if err != nil {
		return err
	}

	return m.repository.DeleteChannelByID(chat.ID)
}
