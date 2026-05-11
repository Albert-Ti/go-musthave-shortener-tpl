package repository

import (
	"context"
	"errors"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/jackc/pgx/v5"
)

var ErrConflict error = errors.New("URL is already exist")

type PostgresStorage struct {
	conn *pgx.Conn
}

type shortenURLRecord struct {
	key string
	url string
}

func NewPostgresStorage(path string) (*PostgresStorage, error) {
	conn, err := pgx.Connect(context.Background(), path)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{
		conn: conn,
	}, nil
}

func (pg *PostgresStorage) Get(ctx context.Context, key string) (string, error) {
	var u shortenURLRecord

	queryStr := `
		SELECT key, url 
		FROM shorten_url 
		WHERE key = $1
	`

	err := pg.conn.QueryRow(ctx, queryStr, key).Scan(&u.key, &u.url)
	if err != nil {
		return "", err
	}

	return u.url, nil
}

func (pg *PostgresStorage) Save(ctx context.Context, key string, url string) (string, error) {

	queryStr := `
		INSERT INTO shorten_url (key, url)
		VALUES ($1, $2)
		ON CONFLICT (url)
		DO UPDATE SET url = shorten_url.url
		RETURNING key
	`

	var returnedKey string
	err := pg.conn.QueryRow(ctx, queryStr, key, url).Scan(&returnedKey)

	if err != nil {
		return "", err
	}

	if returnedKey != key {
		return returnedKey, ErrConflict
	}

	return key, nil
}

func (pg *PostgresStorage) BatchSave(ctx context.Context, keys []string, batch []model.JSONBatchReq) (BatchConflict, error) {
	tx, err := pg.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return BatchConflict{}, err
	}
	defer tx.Rollback(ctx)

	queryStr := `
		INSERT INTO shorten_url (key, url)
		VALUES ($1, $2)
		ON CONFLICT (url)
		DO UPDATE SET url = shorten_url.url
		RETURNING key
	`

	for i, v := range batch {
		var returnedKey string
		err = tx.QueryRow(ctx, queryStr, keys[i], v.OriginalURL).Scan(&returnedKey)
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

	err = tx.Commit(ctx)
	if err != nil {
		return BatchConflict{}, err
	}
	return BatchConflict{}, nil
}

func (pg *PostgresStorage) Ping(ctx context.Context) error {
	return pg.conn.Ping(ctx)
}

func (pg *PostgresStorage) Close() error {
	return pg.conn.Close(context.Background())
}
