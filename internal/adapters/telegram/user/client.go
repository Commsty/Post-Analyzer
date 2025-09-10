package user

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gotd/td/tg"
)

var (
	ErrSessionVault = errors.New("no session vault found")

	ErrAuthFailed = errors.New("authentication failed")

	ErrAPICallFailed = errors.New("API call failed")

	ErrChannelNotFound = errors.New("no channel found")

	ErrTimeLimit = errors.New("telegram user req time limit reached")
)

type TelegramService interface {
	ChannelPosts(ctx context.Context, username string, lastReadID int64) ([]*tg.Message, error)
	ChannelInfo(ctx context.Context, username string) (*tg.Channel, error)
}

type telegramUserClient struct {
	appID       int
	appHash     string
	sessionPath string
}

func NewTelegramUserClient(appID int, appHash string, sessionPath string) (*telegramUserClient, error) {

	if _, err := os.Stat(sessionPath); err != nil {
		directory := filepath.Dir(sessionPath)
		if err = os.MkdirAll(directory, 0700); err != nil {
			return nil, ErrSessionVault
		}
	}

	client := &telegramUserClient{
		appID:       appID,
		appHash:     appHash,
		sessionPath: sessionPath,
	}

	authCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.authenticate(authCtx); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrAuthFailed, err)
	}

	return client, nil
}

func (t telegramUserClient) ChannelPosts(ctx context.Context, username string, lastReadID int64) ([]*tg.Message, error) {

	var posts []*tg.Message

	client := t.createRawClient()

	err := client.Run(ctx, func(ctx context.Context) error {

		apiClient := tg.NewClient(client)

		channel, err := t.channel(ctx, apiClient, username)
		if err != nil {
			return err
		}

		history, err := t.channelHistory(ctx, apiClient, channel, lastReadID)
		if err != nil {
			return err
		}

		posts = t.filterMessages(history)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (t telegramUserClient) ChannelInfo(ctx context.Context, username string) (*tg.Channel, error) {

	var channel *tg.Channel

	client := t.createRawClient()

	err := client.Run(ctx, func(ctx context.Context) error {

		apiClient := tg.NewClient(client)

		var err error
		channel, err = t.channel(ctx, apiClient, username)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return channel, nil
}

func (t telegramUserClient) channel(ctx context.Context, api *tg.Client, username string) (*tg.Channel, error) {

	if ctx.Err() != nil {
		return nil, ErrTimeLimit
	}

	resolved, err := api.ContactsResolveUsername(ctx,
		&tg.ContactsResolveUsernameRequest{
			Username: username,
		})

	if err != nil {
		return nil, ErrChannelNotFound
	}

	channel, ok := resolved.Chats[0].(*tg.Channel)
	if !ok {
		return nil, ErrChannelNotFound
	}

	return channel, nil
}

func (t telegramUserClient) channelHistory(ctx context.Context, api *tg.Client, channel *tg.Channel, lastReadID int64) (tg.MessagesMessagesClass, error) {

	if ctx.Err() != nil {
		return nil, ErrTimeLimit
	}

	history, err := api.MessagesGetHistory(ctx,
		&tg.MessagesGetHistoryRequest{
			Peer: &tg.InputPeerChannel{
				ChannelID:  channel.ID,
				AccessHash: channel.AccessHash,
			},
			Limit: 30,
			MinID: int(lastReadID),
		})

	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrAPICallFailed, err)
	}

	return history, nil
}

func (t telegramUserClient) filterMessages(history tg.MessagesMessagesClass) []*tg.Message {

	messages, ok := history.(*tg.MessagesChannelMessages)
	if !ok {
		return make([]*tg.Message, 0)
	}

	var posts []*tg.Message
	for _, msg := range messages.Messages {
		if m, ok := msg.(*tg.Message); ok {
			if m.Message != "" {
				posts = append(posts, m)
			}
		}
	}

	return posts
}
