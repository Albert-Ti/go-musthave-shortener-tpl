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

func (ms *MemoryStorage) Get(ctx context.Context, key string) (string, error) {
	return ms.urls[key], nil
}

func (ms *MemoryStorage) GetAllByUserID(ctx context.Context, userID string) ([]map[string]string, error) {

	var results = make([]map[string]string, 0)
	for k, v := range ms.urls {
		results = append(results, map[string]string{
			"key": k,
			"url": v,
		})
	}
	return results, nil
}

func (ms *MemoryStorage) Save(ctx context.Context, key string, url string) (string, error) {
	ms.urls[key] = url

	return key, nil
}

func (ms *MemoryStorage) Close() error {
	return nil
}

func (ms *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}
