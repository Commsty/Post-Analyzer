package user

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

func (t *TelegramUserClient) GetNewChannelPosts(ctx context.Context, username string, lastReadID int) ([]*tg.Message, error) {

	var posts []*tg.Message

	client := t.createRawClient(t.appID, t.appHash)

	err := client.Run(ctx, func(ctx context.Context) error {
		var err error
		posts, err = t.getNewChannelPosts(ctx, client, username, lastReadID)
		return err
	})

	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (t *TelegramUserClient) getNewChannelPosts(ctx context.Context, client *telegram.Client, username string, lastReadID int) ([]*tg.Message, error) {

	api := tg.NewClient(client)

	channel, err := t.resolveChannel(ctx, api, username)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch channel info: %w", err)
	}

	history, err := t.getChannelHistory(ctx, api, channel, lastReadID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get channel history: %w", err)
	}

	posts := t.filterValidMessages(history)

	return posts, nil
}

func (t *TelegramUserClient) resolveChannel(ctx context.Context, api *tg.Client, username string) (*tg.Channel, error) {

	resolved, err := api.ContactsResolveUsername(ctx,
		&tg.ContactsResolveUsernameRequest{
			Username: username,
		})

	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	channel, ok := resolved.Chats[0].(*tg.Channel)
	if !ok {
		return nil, fmt.Errorf("@%s is not public channel", username)
	}

	return channel, nil
}

func (t *TelegramUserClient) getChannelHistory(ctx context.Context, api *tg.Client, channel *tg.Channel, lastReadID int) (tg.MessagesMessagesClass, error) {

	history, err := api.MessagesGetHistory(ctx,
		&tg.MessagesGetHistoryRequest{
			Peer: &tg.InputPeerChannel{
				ChannelID:  channel.ID,
				AccessHash: channel.AccessHash,
			},
			Limit: 15,
			MinID: lastReadID,
		})

	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	return history, nil
}

func (t *TelegramUserClient) filterValidMessages(history tg.MessagesMessagesClass) []*tg.Message {

	messages, ok := history.(*tg.MessagesChannelMessages)
	if !ok {
		return nil
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
