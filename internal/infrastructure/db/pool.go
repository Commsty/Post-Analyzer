package db

import (
	"context"
	"fmt"
	"log"
	"post-analyzer/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Init(connectionCtx context.Context, cfg *config.AppConfig) (*pgxpool.Pool, error) {

	// configurating connection
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 1
	config.HealthCheckPeriod = 5 * time.Second
	config.MaxConnLifetime = 30 * time.Minute

	// attempts to connect db
	var pool *pgxpool.Pool
	for i := range 5 {

		pool, err = pgxpool.NewWithConfig(connectionCtx, config)
		if err == nil {
			break
		}

		log.Printf("Attempt to connect DB №%d: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	// checking db connection
	if err := pool.Ping(connectionCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
