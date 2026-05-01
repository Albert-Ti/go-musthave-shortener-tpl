package db

import (
	"context"
	"time"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/jackc/pgx/v5"
)

func connection() (*pgx.Conn, error) {
	return pgx.Connect(context.Background(), config.Envs.DatabaseDSN)
}

func Ping() error {
	conn, err := connection()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		conn.Close(ctx)
		return err
	}
	return nil
}
