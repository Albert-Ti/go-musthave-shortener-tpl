package repository

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type PostgresStorage struct {
	ctx  context.Context
	conn *pgx.Conn
}

type shortenURLRecord struct {
	key string
	url string
}

func NewPostgresStorage(path string) (*PostgresStorage, error) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, path)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{
		ctx:  ctx,
		conn: conn,
	}, nil
}

func (pg *PostgresStorage) Get(key string) string {
	var u shortenURLRecord

	sql := "SELECT key, url FROM shorten_url WHERE key = $1"

	err := pg.conn.QueryRow(pg.ctx, sql, key).Scan(&u.key, &u.url)
	if err != nil {
		slog.Error(err.Error())
	}

	return u.url
}

func (pg *PostgresStorage) Save(key string, url string) {
	sql := "INSERT INTO shorten_url (key, url) VALUES ($1, $2) ON CONFLICT (key) DO NOTHING"

	_, err := pg.conn.Exec(pg.ctx, sql, key, url)
	if err != nil {
		slog.Error(err.Error())
	}
}

func (pg *PostgresStorage) Length() (int, error) {
	var length int
	sql := "SELECT COUNT(*) FROM shorten_url"

	err := pg.conn.QueryRow(pg.ctx, sql).Scan(&length)

	if err != nil {
		return 0, err
	}

	return length, nil
}

func (pg *PostgresStorage) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := pg.conn.Ping(ctx); err != nil {
		pg.conn.Close(ctx)
		return err
	}
	return nil
}

func (pg *PostgresStorage) Close() error {
	return pg.conn.Close(pg.ctx)
}
