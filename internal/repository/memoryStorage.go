package repository

import (
	"context"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
)

type MemoryStorage struct {
	urls map[string]string
}

func NewMemoryStorage() (*MemoryStorage, error) {
	return &MemoryStorage{urls: map[string]string{}}, nil
}

func (u *MemoryStorage) Get(ctx context.Context, key string) (string, error) {
	return u.urls[key], nil
}

func (u *MemoryStorage) Save(ctx context.Context, key string, url string) (string, error) {
	u.urls[key] = url

	return key, nil
}

func (u *MemoryStorage) BatchSave(ctx context.Context, keys []string, batch []model.JSONBatchReq) (BatchConflict, error) {
	for i, v := range batch {
		u.urls[keys[i]] = v.OriginalURL
	}

	return BatchConflict{}, nil
}

func (u *MemoryStorage) Close() error {
	return nil
}

func (u *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}
