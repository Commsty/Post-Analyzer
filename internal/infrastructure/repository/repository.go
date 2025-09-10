package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"post-analyzer/internal/domain/entity"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTimeLimit = errors.New("db query time limit reached")

	ErrTransactionFailed = errors.New("transaction failed")
	ErrRollbackFailed    = errors.New("rollback failed")
	ErrCommitFailed      = errors.New("commiting transaction failed")

	ErrSelectionFailed = errors.New("db selection failed")
	ErrInsertionFailed = errors.New("db insertion failed")
	ErrUpdateFailed    = errors.New("db update failed")
	ErrDeletingFailed  = errors.New("db deleting failed")

	ErrMappingFailed       = errors.New("mapping to subscription struct failed")
	ErrReadingStreamFailed = errors.New("error during stream reading")
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
		return ErrTimeLimit
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrTransactionFailed, err)
	}

	defer func() {
		if err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				log.Printf("%v: %s", ErrRollbackFailed, err)
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
		return fmt.Errorf("%w: %s", ErrInsertionFailed, err)
	}

	_, err = tx.Exec(ctx,
		`
		INSERT INTO subscription(chat_id, channel_id, last_checked_id, send_time, schedule_id)
		VALUES ($1, $2, $3, $4, $5)
		`,
		sub.ChatID, sub.ChannelID, sub.LastCheckedPostID, sub.SendingTime, sub.ScheduleID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInsertionFailed, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrCommitFailed, err)
	}

	return nil
}

func (r *subscriptionRepository) GetSubscriptions(ctx context.Context, chatID int64) ([]*entity.Subscription, error) {

	if err := ctx.Err(); err != nil {
		return nil, ErrTimeLimit
	}

	rows, err := r.db.Query(ctx,
		`
		SELECT s.chat_id, s.channel_id, c.username, s.last_checked_id, s.send_time, s.schedule_id
		FROM subscription s INNER JOIN channel c USING(channel_id)
		WHERE s.chat_id = $1
		`,
		chatID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrSelectionFailed, err)
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
			return nil, fmt.Errorf("%w: %s", ErrMappingFailed, err)
		}

		subs = append(subs, &sub)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrReadingStreamFailed, err)
	}

	return subs, nil
}

func (r *subscriptionRepository) UpdateSubscription(ctx context.Context, sub *entity.Subscription) error {

	if err := ctx.Err(); err != nil {
		return ErrTimeLimit
	}

	_, err := r.db.Exec(ctx,
		`
		UPDATE subscription
		SET last_checked_id = $1, schedule_id = $2
		WHERE chat_id = $3 AND channel_id = $4 AND send_time = $5
		`,
		sub.LastCheckedPostID, sub.ScheduleID, sub.ChatID, sub.ChannelID, sub.SendingTime)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrUpdateFailed, err)
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
		return fmt.Errorf("%w: %s", ErrDeletingFailed, err)
	}

	return nil
}
