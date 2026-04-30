package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/jackc/pgx/v5"
)

func Init() {
	conn, err := pgx.Connect(context.Background(), config.Envs.DatabaseURL)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		conn.Close(ctx)
		panic(err)
	}

	slog.Info("Connect to DB")
}
