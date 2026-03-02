package repository

import (
	"context"
	"fmt"
	"testing"

	"post-analyzer/internal/domain/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockSubscriptionRepository struct {
	subscriptions map[int64][]*entity.Subscription
}

func NewMockSubscriptionRepository() *MockSubscriptionRepository {
	return &MockSubscriptionRepository{
		subscriptions: make(map[int64][]*entity.Subscription),
	}
}

func (m *MockSubscriptionRepository) AddSubscription(ctx context.Context, sub *entity.Subscription) error {
	if err := ctx.Err(); err != nil {
		return ErrTimeLimit
	}

	if m.subscriptions[sub.ChatID] == nil {
		m.subscriptions[sub.ChatID] = []*entity.Subscription{}
	}

	m.subscriptions[sub.ChatID] = append(m.subscriptions[sub.ChatID], sub)
	return nil
}

func (m *MockSubscriptionRepository) GetSubscriptions(ctx context.Context, chatID int64) ([]*entity.Subscription, error) {
	if err := ctx.Err(); err != nil {
		return nil, ErrTimeLimit
	}

	subs, exists := m.subscriptions[chatID]
	if !exists {
		return []*entity.Subscription{}, nil
	}

	result := make([]*entity.Subscription, len(subs))
	for i, sub := range subs {
		subCopy := *sub
		result[i] = &subCopy
	}

	return result, nil
}

func (m *MockSubscriptionRepository) UpdateSubscription(ctx context.Context, sub *entity.Subscription) error {
	if err := ctx.Err(); err != nil {
		return ErrTimeLimit
	}

	subs := m.subscriptions[sub.ChatID]
	for i, existingSub := range subs {
		if existingSub.ChannelID == sub.ChannelID {
			subs[i] = sub
			return nil
		}
	}

	return ErrUpdateFailed
}

func (m *MockSubscriptionRepository) DeleteSubscription(ctx context.Context, sub *entity.Subscription) error {
	if err := ctx.Err(); err != nil {
		return ErrTimeLimit
	}

	subs := m.subscriptions[sub.ChatID]
	for i, existingSub := range subs {
		if existingSub.ChannelID == sub.ChannelID {
			m.subscriptions[sub.ChatID] = append(subs[:i], subs[i+1:]...)
			return nil
		}
	}

	return ErrDeletingFailed
}

func TestMockSubscriptionRepository_AddSubscription(t *testing.T) {
	repo := NewMockSubscriptionRepository()
	ctx := context.Background()

	subscription := &entity.Subscription{
		ChatID:            12345,
		ChannelID:         67890,
		ChannelUsername:   "@testchannel",
		LastCheckedPostID: 0,
		SendingTime:       "09:00",
		ScheduleID:        1,
	}

	err := repo.AddSubscription(ctx, subscription)
	assert.NoError(t, err)

	subs, err := repo.GetSubscriptions(ctx, 12345)
	assert.NoError(t, err)
	assert.Len(t, subs, 1)
	assert.Equal(t, subscription.ChatID, subs[0].ChatID)
	assert.Equal(t, subscription.ChannelID, subs[0].ChannelID)
	assert.Equal(t, subscription.ChannelUsername, subs[0].ChannelUsername)
}

func TestMockSubscriptionRepository_GetSubscriptions(t *testing.T) {
	repo := NewMockSubscriptionRepository()
	ctx := context.Background()

	subscriptions := []*entity.Subscription{
		{
			ChatID:            12345,
			ChannelID:         67890,
			ChannelUsername:   "@testchannel1",
			LastCheckedPostID: 0,
			SendingTime:       "09:00",
			ScheduleID:        1,
		},
		{
			ChatID:            12345,
			ChannelID:         67891,
			ChannelUsername:   "@testchannel2",
			LastCheckedPostID: 0,
			SendingTime:       "18:00",
			ScheduleID:        2,
		},
		{
			ChatID:            67890,
			ChannelID:         12345,
			ChannelUsername:   "@otherchannel",
			LastCheckedPostID: 0,
			SendingTime:       "12:00",
			ScheduleID:        3,
		},
	}

	for _, sub := range subscriptions {
		err := repo.AddSubscription(ctx, sub)
		require.NoError(t, err)
	}

	subs, err := repo.GetSubscriptions(ctx, 12345)
	assert.NoError(t, err)
	assert.Len(t, subs, 2)

	subs, err = repo.GetSubscriptions(ctx, 67890)
	assert.NoError(t, err)
	assert.Len(t, subs, 1)

	subs, err = repo.GetSubscriptions(ctx, 99999)
	assert.NoError(t, err)
	assert.Len(t, subs, 0)
}

func TestMockSubscriptionRepository_UpdateSubscription(t *testing.T) {
	repo := NewMockSubscriptionRepository()
	ctx := context.Background()

	subscription := &entity.Subscription{
		ChatID:            12345,
		ChannelID:         67890,
		ChannelUsername:   "@testchannel",
		LastCheckedPostID: 0,
		SendingTime:       "09:00",
		ScheduleID:        1,
	}

	err := repo.AddSubscription(ctx, subscription)
	require.NoError(t, err)

	subscription.LastCheckedPostID = 12345
	subscription.SendingTime = "18:00"

	err = repo.UpdateSubscription(ctx, subscription)
	assert.NoError(t, err)

	subs, err := repo.GetSubscriptions(ctx, 12345)
	assert.NoError(t, err)
	assert.Len(t, subs, 1)
	assert.Equal(t, int64(12345), subs[0].LastCheckedPostID)
	assert.Equal(t, "18:00", subs[0].SendingTime)
}

func TestMockSubscriptionRepository_DeleteSubscription(t *testing.T) {
	repo := NewMockSubscriptionRepository()
	ctx := context.Background()

	subscription := &entity.Subscription{
		ChatID:            12345,
		ChannelID:         67890,
		ChannelUsername:   "@testchannel",
		LastCheckedPostID: 0,
		SendingTime:       "09:00",
		ScheduleID:        1,
	}

	err := repo.AddSubscription(ctx, subscription)
	require.NoError(t, err)

	subs, err := repo.GetSubscriptions(ctx, 12345)
	assert.NoError(t, err)
	assert.Len(t, subs, 1)

	err = repo.DeleteSubscription(ctx, subscription)
	assert.NoError(t, err)

	subs, err = repo.GetSubscriptions(ctx, 12345)
	assert.NoError(t, err)
	assert.Len(t, subs, 0)
}

func TestMockSubscriptionRepository_ContextCancellation(t *testing.T) {
	repo := NewMockSubscriptionRepository()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	subscription := &entity.Subscription{
		ChatID:            12345,
		ChannelID:         67890,
		ChannelUsername:   "@testchannel",
		LastCheckedPostID: 0,
		SendingTime:       "09:00",
		ScheduleID:        1,
	}

	err := repo.AddSubscription(ctx, subscription)
	assert.Error(t, err)
	assert.Equal(t, ErrTimeLimit, err)
}

func TestMockSubscriptionRepository_ConcurrentAccess(t *testing.T) {
	repo := NewMockSubscriptionRepository()
	ctx := context.Background()

	const numGoroutines = 10
	const numSubscriptions = 5

	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(chatID int64) {
			for j := 0; j < numSubscriptions; j++ {
				subscription := &entity.Subscription{
					ChatID:            chatID,
					ChannelID:         int64(j),
					ChannelUsername:   fmt.Sprintf("@testchannel_%d", j),
					LastCheckedPostID: 0,
					SendingTime:       "09:00",
					ScheduleID:        1,
				}

				if err := repo.AddSubscription(ctx, subscription); err != nil {
					errChan <- err
					return
				}
			}
			errChan <- nil
		}(int64(i))
	}

	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		assert.NoError(t, err)
	}

	for i := 0; i < numGoroutines; i++ {
		subs, err := repo.GetSubscriptions(ctx, int64(i))
		assert.NoError(t, err)
		assert.Len(t, subs, numSubscriptions)
	}
}

func TestMockSubscriptionRepository_UpdateNonExistent(t *testing.T) {
	repo := NewMockSubscriptionRepository()
	ctx := context.Background()

	subscription := &entity.Subscription{
		ChatID:            12345,
		ChannelID:         67890,
		ChannelUsername:   "@testchannel",
		LastCheckedPostID: 0,
		SendingTime:       "09:00",
		ScheduleID:        1,
	}

	err := repo.UpdateSubscription(ctx, subscription)
	assert.Error(t, err)
	assert.Equal(t, ErrUpdateFailed, err)
}

func TestMockSubscriptionRepository_DeleteNonExistent(t *testing.T) {
	repo := NewMockSubscriptionRepository()
	ctx := context.Background()

	subscription := &entity.Subscription{
		ChatID:            12345,
		ChannelID:         67890,
		ChannelUsername:   "@testchannel",
		LastCheckedPostID: 0,
		SendingTime:       "09:00",
		ScheduleID:        1,
	}

	err := repo.DeleteSubscription(ctx, subscription)
	assert.Error(t, err)
	assert.Equal(t, ErrDeletingFailed, err)
}
