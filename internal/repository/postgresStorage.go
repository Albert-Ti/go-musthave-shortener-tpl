package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrConflict error = errors.New("URL is already exist")
var ErrStatusGone error = errors.New("URL deleted")
var ErrNoRows error = errors.New("URL no rows")

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(dsn string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{
		pool: pool,
	}, nil
}

func (ps *PostgresStorage) Get(ctx context.Context, key string) (string, error) {
	var url string
	var isDeleted bool

	queryStr := `SELECT url, is_deleted FROM shorten_url WHERE key = $1`
	err := ps.pool.QueryRow(ctx, queryStr, key).Scan(&url, &isDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNoRows
		}
		return "", err
	}
	if isDeleted {
		return url, ErrStatusGone
	}

	return url, nil
}

func (ps *PostgresStorage) GetAll(ctx context.Context, userID string) ([]map[string]string, error) {

	queryStr := `SELECT key, url FROM shorten_url WHERE user_id = $1`

	rows, err := ps.pool.Query(ctx, queryStr, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results = make([]map[string]string, 0)

	for rows.Next() {
		var key, url string

		err := rows.Scan(&key, &url)
		if err != nil {
			return nil, err
		}
		results = append(results, map[string]string{
			"key": key,
			"url": url,
		})
	}

	return results, nil
}

func (ps *PostgresStorage) Save(ctx context.Context, key string, url string, userID string) (string, error) {

	queryStr := `
		INSERT INTO shorten_url (key, url, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (url)
		DO UPDATE SET url = shorten_url.url
		RETURNING key
	`

	var returnedKey string
	err := ps.pool.QueryRow(ctx, queryStr, key, url, userID).Scan(&returnedKey)

	if err != nil {
		return "", err
	}

	if returnedKey != key {
		return returnedKey, ErrConflict
	}

	return key, nil
}

func (ps *PostgresStorage) BatchDelete(ctx context.Context, keys []string, userID string) error {
	queryStr := `
		UPDATE shorten_url
		SET is_deleted = true
		WHERE user_id = $1
		AND key = ANY($2)
		AND is_deleted = false
	`
	_, err := ps.pool.Exec(ctx, queryStr, userID, keys)
	if err != nil {
		return err
	}

	return nil
}

func (ps *PostgresStorage) Ping(ctx context.Context) error {
	return ps.pool.Ping(ctx)
}

func (ps *PostgresStorage) Close() error {
	ps.pool.Close()
	return nil
}
