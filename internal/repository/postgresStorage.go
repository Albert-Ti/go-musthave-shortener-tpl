package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/jackc/pgx/v5"
)

var ErrConflict error = errors.New("URL is already exist")

type PostgresStorage struct {
	ctx  context.Context
	conn *pgx.Conn
}

type shortenURLRecord struct {
	key string
	url string
}

func NewPostgresStorage(path string, ctx context.Context) (*PostgresStorage, error) {
	conn, err := pgx.Connect(ctx, path)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{
		ctx:  ctx,
		conn: conn,
	}, nil
}

func (pg *PostgresStorage) Get(key string) (string, error) {
	var u shortenURLRecord

	queryStr := "SELECT key, url FROM shorten_url WHERE key = $1"

	err := pg.conn.QueryRow(pg.ctx, queryStr, key).Scan(&u.key, &u.url)
	if err != nil {
		return "", err
	}

	return u.url, nil
}

func (pg *PostgresStorage) Save(key string, url string) (string, error) {

	queryStr := `
		INSERT INTO shorten_url (key, url)
		VALUES ($1, $2)
		ON CONFLICT (url)
		DO UPDATE SET url = shorten_url.url
		RETURNING key
	`

	var returnedKey string
	err := pg.conn.QueryRow(pg.ctx, queryStr, key, url).Scan(&returnedKey)

	if err != nil {
		return "", err
	}

	if returnedKey != key {
		return returnedKey, ErrConflict
	}

	return key, nil
}

func (pg *PostgresStorage) BatchSave(keys []string, batch []model.JSONBatchReq) (BatchConflict, error) {
	tx, err := pg.conn.BeginTx(pg.ctx, pgx.TxOptions{})
	if err != nil {
		return BatchConflict{}, err
	}
	defer tx.Rollback(pg.ctx)

	queryStr := `
		INSERT INTO shorten_url (key, url)
		VALUES ($1, $2)
		ON CONFLICT (url)
		DO UPDATE SET url = shorten_url.url
		RETURNING key
	`

	for i, v := range batch {
		var returnedKey string
		err = tx.QueryRow(pg.ctx, queryStr, keys[i], v.OriginalURL).Scan(&returnedKey)
		if err != nil {
			return BatchConflict{}, err
		}

		if returnedKey != keys[i] {
			return BatchConflict{
				CorrelationID: v.CorrelationID,
				Key:           returnedKey,
			}, ErrConflict
		}
	}

	err = tx.Commit(pg.ctx)
	if err != nil {
		return BatchConflict{}, err
	}
	return BatchConflict{}, nil
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
