package repository

import (
	"context"
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

func (u *MemoryStorage) Close() error {
	return nil
}

func (u *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}
