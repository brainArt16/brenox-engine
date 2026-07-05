package database

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool() (*pgxpool.Pool, error) {

	/*

		Context is VERY important in Go.

		It carries:
		- cancellation signals
		- deadlines
		- request lifecycle info

	*/

	ctx := context.Background()

	/*
		Build PostgreSQL connection string.
	*/

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	/*
		Create reusable connection pool.
	*/

	pool, err := pgxpool.New(ctx, databaseURL)

	if err != nil {
		return nil, err
	}

	/*
		Ping database to verify connection.
	*/

	err = pool.Ping(ctx)

	if err != nil {
		return nil, err
	}

	return pool, nil
}

func sslMode() string {
	if mode := strings.TrimSpace(os.Getenv("DB_SSLMODE")); mode != "" {
		return mode
	}
	return "prefer"
}