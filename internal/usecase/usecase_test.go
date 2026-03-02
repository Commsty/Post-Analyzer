package usecase

import (
	"context"
	"errors"
	"testing"

	"post-analyzer/internal/domain/dto"
	"post-analyzer/internal/domain/entity"
	"post-analyzer/internal/infrastructure/repository"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockTelegramService struct {
	posts []*tg.Message
}

func NewMockTelegramService() *MockTelegramService {
	return &MockTelegramService{
		posts: []*tg.Message{
			{ID: 1, Message: "First post content"},
			{ID: 2, Message: "Second post content"},
			{ID: 3, Message: "Third post content"},
		},
	}
}

func (m *MockTelegramService) ChannelPosts(ctx context.Context, username string, lastPostID int64) ([]*tg.Message, error) {
	if username == "" {
		return nil, errors.New("channel username is required")
	}

	var result []*tg.Message
	for _, post := range m.posts {
		if int64(post.ID) > lastPostID {
			result = append(result, post)
		}
	}

	return result, nil
}

func (m *MockTelegramService) ChannelInfo(ctx context.Context, username string) (*tg.Channel, error) {
	if username == "" {
		return nil, errors.New("channel username is required")
	}

	return &tg.Channel{
		ID:         12345,
		AccessHash: 67890,
		Title:      "Test Channel",
		Username:   username,
	}, nil
}

type MockAnalysisService struct {
	analysis map[string]string
}

func NewMockAnalysisService() *MockAnalysisService {
	return &MockAnalysisService{
		analysis: map[string]string{
			"First post content":  "Analysis: First post is about technology",
			"Second post content": "Analysis: Second post discusses business",
			"Third post content":  "Analysis: Third post covers science",
		},
	}
}

func (m *MockAnalysisService) AnalyzePosts(ctx context.Context, text string) (string, error) {
	if text == "" {
		return "", errors.New("content is required")
	}

	analysis, exists := m.analysis[text]
	if !exists {
		return "Analysis: Generic content analysis", nil
	}

	return analysis, nil
}

type MockNotifier struct {
	notifications []string
}

func NewMockNotifier() *MockNotifier {
	return &MockNotifier{
		notifications: []string{},
	}
}

func (m *MockNotifier) NotifyWithText(ctx context.Context, chatID int64, message string) error {
	if message == "" {
		return errors.New("message is required")
	}

	m.notifications = append(m.notifications, message)
	return nil
}

func (m *MockNotifier) GetNotifications() []string {
	return m.notifications
}

type MockScheduler struct {
	jobs []int
}

func NewMockScheduler() *MockScheduler {
	return &MockScheduler{
		jobs: []int{},
	}
}

func (m *MockScheduler) ScheduleEvent(sub *entity.Subscription, job func()) (int, error) {
	if sub.SendingTime == "" {
		return 0, errors.New("schedule time is required")
	}

	jobID := len(m.jobs) + 1
	m.jobs = append(m.jobs, jobID)
	return jobID, nil
}

func (m *MockScheduler) Start() {}

func (m *MockScheduler) Stop() {}

func (m *MockScheduler) GetJobs() []int {
	return m.jobs
}

type MockRepository struct {
	subscriptions map[int64][]*entity.Subscription
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		subscriptions: make(map[int64][]*entity.Subscription),
	}
}

func (m *MockRepository) AddSubscription(ctx context.Context, sub *entity.Subscription) error {
	if err := ctx.Err(); err != nil {
		return repository.ErrTimeLimit
	}

	if m.subscriptions[sub.ChatID] == nil {
		m.subscriptions[sub.ChatID] = []*entity.Subscription{}
	}

	m.subscriptions[sub.ChatID] = append(m.subscriptions[sub.ChatID], sub)
	return nil
}

func (m *MockRepository) GetSubscriptions(ctx context.Context, chatID int64) ([]*entity.Subscription, error) {
	if err := ctx.Err(); err != nil {
		return nil, repository.ErrTimeLimit
	}

	subs, exists := m.subscriptions[chatID]
	if !exists {
		return []*entity.Subscription{}, nil
	}

	return subs, nil
}

func (m *MockRepository) UpdateSubscription(ctx context.Context, sub *entity.Subscription) error {
	if err := ctx.Err(); err != nil {
		return repository.ErrTimeLimit
	}

	subs := m.subscriptions[sub.ChatID]
	for i, existingSub := range subs {
		if existingSub.ChannelID == sub.ChannelID {
			subs[i] = sub
			return nil
		}
	}

	return repository.ErrUpdateFailed
}

func (m *MockRepository) DeleteSubscription(ctx context.Context, sub *entity.Subscription) error {
	if err := ctx.Err(); err != nil {
		return repository.ErrTimeLimit
	}

	subs := m.subscriptions[sub.ChatID]
	for i, existingSub := range subs {
		if existingSub.ChannelID == sub.ChannelID {
			m.subscriptions[sub.ChatID] = append(subs[:i], subs[i+1:]...)
			return nil
		}
	}

	return repository.ErrDeletingFailed
}

func TestUseCaseManager_MonitorChannel(t *testing.T) {
	telegramService := NewMockTelegramService()
	repo := NewMockRepository()
	scheduler := NewMockScheduler()
	aiService := NewMockAnalysisService()
	notifier := NewMockNotifier()

	ucManager := NewUseCaseManager(telegramService, repo, scheduler, aiService, notifier)

	ctx := context.Background()

	mr := &dto.MonitorRequest{
		ChatID:  12345,
		Message: "/monitor @testchannel 09:00",
	}

	err := ucManager.MonitorChannel(ctx, mr)
	assert.NoError(t, err)

	subs, err := repo.GetSubscriptions(ctx, 12345)
	assert.NoError(t, err)
	assert.Len(t, subs, 1)
	assert.Equal(t, int64(12345), subs[0].ChatID)
	assert.Equal(t, "@testchannel", subs[0].ChannelUsername)

	jobs := scheduler.GetJobs()
	assert.Len(t, jobs, 1)
}

func TestUseCaseManager_MonitorChannelWithTime(t *testing.T) {
	telegramService := NewMockTelegramService()
	repo := NewMockRepository()
	scheduler := NewMockScheduler()
	aiService := NewMockAnalysisService()
	notifier := NewMockNotifier()

	ucManager := NewUseCaseManager(telegramService, repo, scheduler, aiService, notifier)

	ctx := context.Background()

	mr := &dto.MonitorRequest{
		ChatID:  12345,
		Message: "/monitor @testchannel 18:30",
	}

	err := ucManager.MonitorChannel(ctx, mr)
	assert.NoError(t, err)

	subs, err := repo.GetSubscriptions(ctx, 12345)
	assert.NoError(t, err)
	assert.Len(t, subs, 1)
	assert.Equal(t, "18:30", subs[0].SendingTime)
}

func TestUseCaseManager_ContextCancellation(t *testing.T) {
	telegramService := NewMockTelegramService()
	repo := NewMockRepository()
	scheduler := NewMockScheduler()
	aiService := NewMockAnalysisService()
	notifier := NewMockNotifier()

	ucManager := NewUseCaseManager(telegramService, repo, scheduler, aiService, notifier)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mr := &dto.MonitorRequest{
		ChatID:  12345,
		Message: "/monitor @testchannel 09:00",
	}

	err := ucManager.MonitorChannel(ctx, mr)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrTimeLimit, err)
}

func TestUseCaseManager_InvalidChannel(t *testing.T) {
	telegramService := NewMockTelegramService()
	repo := NewMockRepository()
	scheduler := NewMockScheduler()
	aiService := NewMockAnalysisService()
	notifier := NewMockNotifier()

	ucManager := NewUseCaseManager(telegramService, repo, scheduler, aiService, notifier)

	ctx := context.Background()

	mr := &dto.MonitorRequest{
		ChatID:  12345,
		Message: "/monitor",
	}

	err := ucManager.MonitorChannel(ctx, mr)
	assert.Error(t, err)
}

func TestUseCaseManager_Integration(t *testing.T) {
	telegramService := NewMockTelegramService()
	repo := NewMockRepository()
	scheduler := NewMockScheduler()
	aiService := NewMockAnalysisService()
	notifier := NewMockNotifier()

	ucManager := NewUseCaseManager(telegramService, repo, scheduler, aiService, notifier)

	ctx := context.Background()

	mr := &dto.MonitorRequest{
		ChatID:  12345,
		Message: "/monitor @testchannel 09:00",
	}

	err := ucManager.MonitorChannel(ctx, mr)
	require.NoError(t, err)

	subs, err := repo.GetSubscriptions(ctx, 12345)
	require.NoError(t, err)
	require.Len(t, subs, 1)

	posts, err := telegramService.ChannelPosts(ctx, subs[0].ChannelUsername, subs[0].LastCheckedPostID)
	require.NoError(t, err)

	for _, post := range posts {
		analysis, err := aiService.AnalyzePosts(ctx, post.Message)
		require.NoError(t, err)

		err = notifier.NotifyWithText(ctx, 12345, analysis)
		require.NoError(t, err)
	}

	notifications := notifier.GetNotifications()
	assert.Len(t, notifications, len(posts))
}

func TestUseCaseManager_ConcurrentMonitoring(t *testing.T) {
	telegramService := NewMockTelegramService()
	repo := NewMockRepository()
	scheduler := NewMockScheduler()
	aiService := NewMockAnalysisService()
	notifier := NewMockNotifier()

	ucManager := NewUseCaseManager(telegramService, repo, scheduler, aiService, notifier)

	ctx := context.Background()

	const numGoroutines = 5
	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(chatID int64) {
			mr := &dto.MonitorRequest{
				ChatID:  chatID,
				Message: "/monitor @testchannel 09:00",
			}

			err := ucManager.MonitorChannel(ctx, mr)
			errChan <- err
		}(int64(10000 + i))
	}

	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		assert.NoError(t, err)
	}

	for i := 0; i < numGoroutines; i++ {
		subs, err := repo.GetSubscriptions(ctx, int64(10000+i))
		assert.NoError(t, err)
		assert.Len(t, subs, 1)
	}
}
