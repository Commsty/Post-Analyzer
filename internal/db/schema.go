package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	createSubscriptionTable = `
	CREATE TABLE IF NOT EXISTS subscription (
		id BIGSERIAL PRIMARY KEY,
		chat_id BIGINT NOT NULL,
		channel_id BIGINT NOT NULL,
		last_checked_id BIGINT DEFAULT -1,
		send_time TIME NOT NULL,
		schedule_id INTEGER NOT NULL,
		creation_data TIMESTAMPTZ DEFAULT NOW(),

		UNIQUE(chat_id, channel_id, send_time),

		CONSTRAINT fk_channel
			FOREIGN KEY (channel_id)
			REFERENCES channel(channel_id)
			ON DELETE CASCADE	
	);`

	createChannelTable = `
	CREATE TABLE IF NOT EXISTS channel (
		id BIGSERIAL PRIMARY KEY,
		channel_id BIGINT UNIQUE NOT NULL,
		username TEXT UNIQUE NOT NULL
	);`
)

func EnshureSchema(ctx context.Context, pool *pgxpool.Pool) error {

	if _, err := pool.Exec(ctx, createChannelTable); err != nil {
		return err
	}

	if _, err := pool.Exec(ctx, createSubscriptionTable); err != nil {
		return err
	}

	return nil
}
