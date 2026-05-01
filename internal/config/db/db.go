package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
	"github.com/jackc/pgx/v5"
)

var Connect *pgx.Conn

func Connection() {
	conn, err := pgx.Connect(context.Background(), config.Envs.DatabaseDSN)
	if err != nil {
		panic(err)
	}
	Connect = conn
	slog.Info("Connect to DB", "conn_str: ", config.Envs.DatabaseDSN)
}

func Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := Connect.Ping(ctx); err != nil {
		Connect.Close(ctx)
		return err
	}
	return nil
}
