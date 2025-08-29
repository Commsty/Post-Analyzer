package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Init(connectionCtx context.Context) (*pgxpool.Pool, error) {

	// configurating connection
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("SSL_MODE"),
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
