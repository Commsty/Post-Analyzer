package service

import (
	"context"
	"fmt"
	"log"
	"post-analyzer/internal/entity"
	"post-analyzer/internal/repository"
	"post-analyzer/internal/scheduler"
	"strings"
	"time"
)

type SubscriptionService interface {
	AddSubscription(context.Context, string, string, *entity.Subscription) error
	GetSubscriptions(context.Context, *entity.Subscription) ([]string, error)
	UpdateSubscription(context.Context, *entity.Subscription) error
	DeleteSubscription(context.Context, *entity.Subscription) error
}

type subscriptionService struct {
	repository      repository.SubscriptionRepository
	scheduler       *scheduler.Scheduler
	analysisService analysisProvider
	telegramService telegramProvider
}

func NewSubscriptionService(repo repository.SubscriptionRepository, sched *scheduler.Scheduler,
	analysisServ analysisProvider, telegramServ telegramProvider) SubscriptionService {

	return &subscriptionService{
		repository:      repo,
		scheduler:       sched,
		analysisService: analysisServ,
		telegramService: telegramServ,
	}
}

// CRUD methods
func (s *subscriptionService) AddSubscription(ctx context.Context, link, timeString string, sub *entity.Subscription) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}

	flag, err := s.telegramService.validateChannel(ctx, link)
	if err != nil || !flag {
		return fmt.Errorf("validation failed: %w", err)
	}
	sub.ChannelUsername, err = s.telegramService.extractUsernameFromIdentificator(link)
	if err != nil {
		return fmt.Errorf("link parsing failed: %w", err)
	}

	hour, minute, err := s.scheduler.ParseTimeString(timeString)
	if err != nil {
		return fmt.Errorf("parsing time failed: %w", err)
	}
	sub.SendingTime = strings.TrimSpace(timeString)

	if ctx.Err() != nil {
		return ctx.Err()
	}
	extraInfo, err := s.telegramService.getChatInfo(ctx, sub)
	if err != nil {
		return fmt.Errorf("failed to get chat info: %w", err)
	}
	sub.ChannelID = extraInfo.ID

	schedID, err := s.scheduler.ScheduleJob(hour, minute, func() {
		s.processSubscription(sub)
	})
	if err != nil {
		return fmt.Errorf("scheduling analysis failed: %w", err)
	}
	sub.ScheduleID = schedID

	if ctx.Err() != nil {
		return ctx.Err()
	}
	err = s.repository.AddSubscription(ctx, sub)

	if err != nil {
		return fmt.Errorf("repository failure: %w", err)
	}

	return nil
}

func (s *subscriptionService) GetSubscriptions(ctx context.Context, sub *entity.Subscription) ([]string, error) {

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	infos, err := s.repository.GetSubscriptions(ctx, sub.ChatID)
	if err != nil {
		return nil, fmt.Errorf("repository failure: %v", err)
	}

	var Subscriptions = make([]string, len(infos))
	for i, s := range infos {
		Subscriptions[i] = fmt.Sprintf("Channel @%s at time: %s", s.ChannelUsername, s.SendingTime)
	}

	return Subscriptions, nil
}

func (s *subscriptionService) UpdateSubscription(ctx context.Context, sub *entity.Subscription) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}
	err := s.repository.UpdateSubscription(ctx, sub)
	if err != nil {
		return fmt.Errorf("repository failure: %v", err)
	}

	return nil
}

func (s *subscriptionService) DeleteSubscription(ctx context.Context, sub *entity.Subscription) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}
	err := s.repository.DeleteSubscription(ctx, sub)
	if err != nil {
		return fmt.Errorf("repository failure: %v", err)
	}

	return nil
}

// method for scheduler to call at planned time
func (s *subscriptionService) processSubscription(sub *entity.Subscription) {

	analysisCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	analyzedResult, lastCheckedID, err := s.analyzeChannel(analysisCtx, sub)
	if err != nil {
		log.Printf("processing channel @%s: Analysis failed: %v", sub.ChannelUsername, err)
	}

	analyzedResult = "Самые важные новости таковы:\n\n" + analyzedResult

	if analysisCtx.Err() != nil {
		log.Printf("processing channel @%s: analysis was too long: %v", sub.ChannelUsername, analysisCtx.Err())
		return
	}

	_, err = s.telegramService.sendMessage(analysisCtx, sub, analyzedResult)
	if err != nil {
		log.Printf("processing channel @%s: Sending result message failed: %v", sub.ChannelUsername, err)
	}

	sub.LastCheckedPostID = lastCheckedID

	err = s.UpdateSubscription(analysisCtx, sub)
	if err != nil {
		log.Printf("processing channel @%s: Updating subscription info failed: %v", sub.ChannelUsername, err)
	}
}

func (s *subscriptionService) analyzeChannel(ctx context.Context, sub *entity.Subscription) (string, int64, error) {

	if ctx.Err() != nil {
		return "", sub.LastCheckedPostID, ctx.Err()
	}

	newPosts, err := s.telegramService.getNewChannelPosts(ctx, sub)
	if err != nil {
		return "", sub.LastCheckedPostID, fmt.Errorf("failed to get new posts: %w", err)
	}

	var postsText string
	for _, msg := range newPosts {
		postsText += msg.Message + "\n"
	}

	if ctx.Err() != nil {
		return "", sub.LastCheckedPostID, ctx.Err()
	}
	analysisResult, err := s.analysisService.AnalyzeTextWithBasicPrompt(ctx, postsText)
	if err != nil {
		return "", sub.LastCheckedPostID, fmt.Errorf("failed to analyse new posts: %w", err)
	}

	lastCheckedID := newPosts[0].ID

	return analysisResult, int64(lastCheckedID), nil
}
