package repository

import (
	"context"
	"log"
	"post-analyzer/internal/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository interface {
	AddSubscription(context.Context, *entity.Subscription) error
	GetSubscriptions(context.Context, int64) ([]*entity.Subscription, error)
	UpdateSubscription(context.Context, *entity.Subscription) error
	DeleteSubscription(context.Context, *entity.Subscription) error
}

type subscriptionRepository struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepository(database *pgxpool.Pool) SubscriptionRepository {
	return &subscriptionRepository{
		db: database,
	}
}

func (r *subscriptionRepository) AddSubscription(ctx context.Context, sub *entity.Subscription) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			rollbackError := tx.Rollback(ctx)
			if rollbackError != nil {
				log.Printf("Rollback failed: %v", rollbackError)
			}
		}
	}()

	_, err = tx.Exec(ctx,
		`
		INSERT INTO channel(channel_id, username)
		VALUES ($1, $2)
		ON CONFLICT (channel_id)
		DO UPDATE SET username = EXCLUDED.username
		`,
		sub.ChannelID, sub.ChannelUsername)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`
		INSERT INTO subscription(chat_id, channel_id, last_checked_id, send_time, schedule_id)
		VALUES ($1, $2, $3, $4, $5)
		`,
		sub.ChatID, sub.ChannelID, sub.LastCheckedPostID, sub.SendingTime, sub.ScheduleID)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	return err
}

func (r *subscriptionRepository) GetSubscriptions(ctx context.Context, chatID int64) ([]*entity.Subscription, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx,
		`
		SELECT s.chat_id, s.channel_id, c.username, s.last_checked_id, s.send_time, s.schedule_id
		FROM subscription s INNER JOIN channel c USING(channel_id)
		WHERE s.chat_id = $1
		`,
		chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*entity.Subscription
	for rows.Next() {

		var sub entity.Subscription
		err := rows.Scan(
			&sub.ChatID,
			&sub.ChannelID,
			&sub.ChannelUsername,
			&sub.LastCheckedPostID,
			&sub.SendingTime,
			&sub.ScheduleID,
		)
		if err != nil {
			return nil, err
		}

		subs = append(subs, &sub)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return subs, nil
}

func (r *subscriptionRepository) UpdateSubscription(ctx context.Context, sub *entity.Subscription) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	_, err := r.db.Exec(ctx,
		`
		UPDATE subscription
		SET last_checked_id = $1
		WHERE chat_id = $2 AND channel_id = $3 AND send_time = $4
		`,
		sub.LastCheckedPostID, sub.ChatID, sub.ChannelID, sub.SendingTime)
	if err != nil {
		return err
	}

	return nil
}

func (r *subscriptionRepository) DeleteSubscription(ctx context.Context, sub *entity.Subscription) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	_, err := r.db.Exec(ctx,
		`
		DELETE FROM subscription
		WHERE chat_id = $1 AND channel_id = $2 AND send_time = $3
		`,
		sub.ChatID, sub.ChannelID, sub.SendingTime)
	if err != nil {
		return err
	}

	return nil
}
