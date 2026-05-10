package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	queryStr := "INSERT INTO shorten_url (key, url) VALUES ($1, $2)"

	var pgErr *pgconn.PgError
	_, err := pg.conn.Exec(pg.ctx, queryStr, key, url)
	if err != nil {
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {

				var existKey string

				err := pg.conn.QueryRow(pg.ctx,
					`SELECT key FROM shorten_url WHERE url = $1`,
					url,
				).Scan(&existKey)

				if err != nil {
					return "", err
				}
				return existKey, ErrConflict
			}
		}
		return "", err
	}
	return key, nil
}

func (pg *PostgresStorage) BatchSave(keys []string, batch []model.JSONBatchReq) (string, string, error) {
	tx, err := pg.conn.BeginTx(pg.ctx, pgx.TxOptions{})
	if err != nil {
		return "", "", err
	}
	defer tx.Rollback(pg.ctx)

	var pgErr *pgconn.PgError

	for i, v := range batch {
		queryStr := "INSERT INTO shorten_url (key, url, correlation_id) VALUES ($1, $2, $3)"
		_, err = tx.Exec(pg.ctx, queryStr, keys[i], v.OriginalURL, v.CorrelationID)
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {

				tx.Rollback(pg.ctx)

				var existKey string
				var existID *string
				err := pg.conn.QueryRow(pg.ctx,
					`SELECT key, correlation_id FROM shorten_url WHERE url = $1`,
					v.OriginalURL,
				).Scan(&existKey, &existID)

				if err != nil {
					return "", "", err
				}

				if existID != nil {
					return existKey, *existID, ErrConflict
				}
				return existKey, "", ErrConflict
			}
		}

	}

	err = tx.Commit(pg.ctx)
	if err != nil {
		return "", "", err
	}

	return "", "", nil
}

func (pg *PostgresStorage) Length() (int, error) {
	var length int
	queryStr := "SELECT COUNT(*) FROM shorten_url"

	err := pg.conn.QueryRow(pg.ctx, queryStr).Scan(&length)

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
