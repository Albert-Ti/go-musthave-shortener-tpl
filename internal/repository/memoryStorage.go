package repository

import (
	"context"
	"errors"
)

type MemoryStorage struct {
	urls []map[string]string
}

func NewMemoryStorage() (*MemoryStorage, error) {
	return &MemoryStorage{urls: []map[string]string{}}, nil
}

func (ms *MemoryStorage) Get(ctx context.Context, key string) (string, error) {
	for i := range ms.urls {
		if ms.urls[i]["key"] == key {
			return ms.urls[i]["url"], nil
		}
	}
	return "", errors.New("No Content")
}

func (ms *MemoryStorage) GetAll(ctx context.Context, userID string) ([]map[string]string, error) {
	var results = make([]map[string]string, 0)

	for i := range ms.urls {
		if ms.urls[i]["user_id"] == userID {
			results = append(results, ms.urls[i])
		}
	}

	return results, nil
}

func (ms *MemoryStorage) Save(ctx context.Context, key string, url string, userID string) (string, error) {
	ms.urls = append(ms.urls, map[string]string{
		"key":     key,
		"url":     url,
		"user_id": userID,
	})

	return key, nil
}

func (ms *MemoryStorage) Close() error {
	return nil
}

func (ms *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}
