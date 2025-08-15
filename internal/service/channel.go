package service

import (
	"context"
	"fmt"
	"log"
	"post-analyzer/internal/entity"
	"post-analyzer/internal/repository"
	"post-analyzer/internal/scheduler"
	"time"
)

type ChannelService interface {
	AddChannel(context.Context, string, string, *entity.ChannelInfo) error
	GetAllChannels(context.Context) ([]string, error)
	UpdateChannel(context.Context, *entity.ChannelInfo) error
	DeleteChannel(context.Context, *entity.ChannelInfo) error
}

type channelService struct {
	repository      repository.ChannelStorageRepository
	scheduler       *scheduler.Scheduler
	analysisService analysisProvider
	telegramService telegramProvider
}

func NewChannelService(repo repository.ChannelStorageRepository, sched *scheduler.Scheduler,
	analysisServ analysisProvider, telegramServ telegramProvider) ChannelService {

	return &channelService{
		repository:      repo,
		scheduler:       sched,
		analysisService: analysisServ,
		telegramService: telegramServ,
	}
}

// CRUD methods
func (m *channelService) AddChannel(ctx context.Context, link, timeString string, addedChanInfo *entity.ChannelInfo) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}

	flag, err := m.telegramService.validateChannel(ctx, link)
	if err != nil || !flag {
		return fmt.Errorf("Validation failed: %w", err)
	}

	hour, minute, err := m.scheduler.ParseTimeString(timeString)
	if err != nil {
		return fmt.Errorf("Parsing time failed: %w", err)
	}

	addedChanInfo.ChannelUsername, _ = m.telegramService.extractUsernameFromIdentificator(link)

	if ctx.Err() != nil {
		return ctx.Err()
	}
	extraInfo, err := m.telegramService.getChatInfo(ctx, addedChanInfo)
	addedChanInfo.ChannelID = extraInfo.ID

	schedID, err := m.scheduler.ScheduleJob(hour, minute, func() {
		m.processChannel(addedChanInfo)
	})
	if err != nil {
		return fmt.Errorf("Scheduling analysis failed: %w", err)
	}

	addedChanInfo.ScheduleID = schedID

	if ctx.Err() != nil {
		return ctx.Err()
	}
	err = m.repository.AddChannelInfo(ctx, addedChanInfo)

	if err != nil {
		return fmt.Errorf("Repository failure: %w", err)
	}

	return nil
}

func (m *channelService) GetAllChannels(ctx context.Context) ([]string, error) {

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	infos, err := m.repository.GetAllChannelInfos(ctx)
	if err != nil {
		return nil, fmt.Errorf("Repository failure: %v", err)
	}

	var channels = make([]string, len(infos))
	for i, c := range infos {
		channels[i] = "@" + c.ChannelUsername
	}

	return channels, nil
}

func (m *channelService) UpdateChannel(ctx context.Context, chanInfo *entity.ChannelInfo) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}
	err := m.repository.UpdateChannelInfo(ctx, chanInfo)
	if err != nil {
		return fmt.Errorf("Repository failure: %v", err)
	}

	return nil
}

func (m *channelService) DeleteChannel(ctx context.Context, chanInfo *entity.ChannelInfo) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}
	err := m.repository.DeleteChannelInfo(ctx, chanInfo)
	if err != nil {
		return fmt.Errorf("Repository failure: %v", err)
	}

	return nil
}

// Analysis methods
func (m *channelService) processChannel(chanInfo *entity.ChannelInfo) {

	analysisCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	analyzedResult, lastCheckedID, err := m.analyzeChannel(analysisCtx, chanInfo)
	if err != nil {
		log.Printf("processing channel @%s: Analysis failed: %v", chanInfo.ChannelUsername, err)
	}

	analyzedResult = "Самые важные новости таковы:\n\n" + analyzedResult

	if analysisCtx.Err() != nil {
		log.Printf("processing channel @%s: analysis was too long: %v", chanInfo.ChannelUsername, analysisCtx.Err())
		return
	}

	_, err = m.telegramService.sendMessage(analysisCtx, chanInfo, analyzedResult)
	if err != nil {
		log.Printf("processing channel @%s: Sending result message failed: %v", chanInfo.ChannelUsername, err)
	}

	chanInfo.LastCheckedPostID = lastCheckedID

	err = m.UpdateChannel(analysisCtx, chanInfo)
	if err != nil {
		log.Printf("processing channel @%s: Updating channel info failed: %v", chanInfo.ChannelUsername, err)
	}
}

func (m *channelService) analyzeChannel(ctx context.Context, chanInfo *entity.ChannelInfo) (string, int64, error) {

	if ctx.Err() != nil {
		return "", chanInfo.LastCheckedPostID, ctx.Err()
	}

	newPosts, err := m.telegramService.getNewChannelPosts(ctx, chanInfo)
	if err != nil {
		return "", chanInfo.LastCheckedPostID, fmt.Errorf("Failed to get new posts: %w", err)
	}

	var postsText string
	for _, msg := range newPosts {
		postsText += msg.Message + "\n"
	}

	if ctx.Err() != nil {
		return "", chanInfo.LastCheckedPostID, ctx.Err()
	}
	analysisResult, err := m.analysisService.AnalyzeTextWithBasicPrompt(ctx, postsText)
	if err != nil {
		return "", chanInfo.LastCheckedPostID, fmt.Errorf("Failed to analyse new posts: %w", err)
	}

	lastCheckedID := newPosts[0].ID

	return analysisResult, int64(lastCheckedID), nil
}
