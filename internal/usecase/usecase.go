package usecase

import (
	"context"
	"log"
	"strings"
	"time"

	"post-analyzer/internal/adapters/openrouter"
	"post-analyzer/internal/adapters/telegram/user"
	"post-analyzer/internal/domain/dto"
	"post-analyzer/internal/domain/entity"
	"post-analyzer/internal/domain/presenter"
	"post-analyzer/internal/domain/validation"
	"post-analyzer/internal/infrastructure/notifier"
	"post-analyzer/internal/infrastructure/repository"
	"post-analyzer/internal/infrastructure/scheduler"
)

type UseCase interface {
	MonitorChannel(ctx context.Context, mr *dto.MonitorRequest) error
}

type useCaseManager struct {
	tgc      user.TelegramService
	repo     repository.SubscriptionRepository
	sched    scheduler.Scheduler
	ai       openrouter.AnalysisService
	notifier notifier.Notifier
}

func NewUseCaseManager(tgc user.TelegramService, repo repository.SubscriptionRepository, sched scheduler.Scheduler,
	ai openrouter.AnalysisService, notifier notifier.Notifier) *useCaseManager {

	return &useCaseManager{
		tgc:      tgc,
		repo:     repo,
		sched:    sched,
		ai:       ai,
		notifier: notifier,
	}
}

func (uc useCaseManager) MonitorChannel(ctx context.Context, mr *dto.MonitorRequest) error {

	subscription := &entity.Subscription{
		ChatID:            mr.ChatID,
		LastCheckedPostID: -1,
	}

	validationChain := validation.ArgsValidator(
		validation.TimeValidator(
			validation.ChannelNameValidator(
				validation.ChannelValidator(nil, uc.tgc),
			),
		),
	)

	if err := validationChain(ctx, mr.Message, subscription); err != nil {
		return presenter.PresentError(err)
	}

	if err := uc.repo.AddSubscription(ctx, subscription); err != nil {
		return presenter.PresentError(err)
	}

	var err error
	if subscription.ScheduleID, err = uc.sched.ScheduleEvent(subscription,
		func() {

			analysisCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			posts, err := uc.tgc.ChannelPosts(analysisCtx, subscription.ChannelUsername, subscription.LastCheckedPostID)
			if err != nil {
				log.Println(err)
				return
			}

			subscription.LastCheckedPostID = int64(posts[0].ID)
			if err := uc.repo.UpdateSubscription(analysisCtx, subscription); err != nil {
				log.Println(err)
				return
			}

			var postsBuilder strings.Builder
			for _, post := range posts {
				postsBuilder.WriteString(post.Message + "\n")
			}
			postTexts := postsBuilder.String()

			result, err := uc.ai.AnalyzePosts(analysisCtx, postTexts)
			if err != nil {
				log.Println(err)
				return
			}

			if err := uc.notifier.NotifyWithText(analysisCtx, subscription.ChatID, result); err != nil {
				log.Println(err)
			}

		}); err != nil {
		return presenter.PresentError(err)
	}

	if err := uc.repo.UpdateSubscription(ctx, subscription); err != nil {
		return presenter.PresentError(err)
	}

	return nil
}
