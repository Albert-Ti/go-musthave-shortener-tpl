package repository

import (
	"context"
	"errors"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
)

type memoryStorage struct {
	urls []map[string]string
}

func NewMemoryStorage() (*memoryStorage, error) {
	return &memoryStorage{urls: []map[string]string{}}, nil
}

func (ms *memoryStorage) Get(ctx context.Context, key string) (string, error) {
	for i := range ms.urls {
		if ms.urls[i]["key"] == key {
			return ms.urls[i]["url"], nil
		}
	}
	return "", errors.New("no Content")
}

func (ms *memoryStorage) GetAll(ctx context.Context, userID string) ([]map[string]string, error) {
	var results = make([]map[string]string, 0)

	for i := range ms.urls {
		if ms.urls[i]["user_id"] == userID {
			results = append(results, ms.urls[i])
		}
	}

	return results, nil
}

func (ms *memoryStorage) Save(ctx context.Context, key string, url string, userID string) (string, error) {
	ms.urls = append(ms.urls, map[string]string{
		"key":     key,
		"url":     url,
		"user_id": userID,
	})

	return key, nil
}

func (ms *memoryStorage) BatchSave(ctx context.Context, batch []model.BatchReq, baseURL string, userID string) ([]model.BatchResp, error) {
	result := make([]model.BatchResp, len(batch))

	for i, v := range batch {
		key := utils.GenerateUUID()

		ms.urls = append(ms.urls, map[string]string{
			"key":     key,
			"url":     v.OriginalURL,
			"user_id": userID,
		})

		result[i] = model.BatchResp{
			CorrelationID: v.CorrelationID,
			ShortURL:      baseURL + "/" + key,
		}
	}

	return result, nil
}

func (ms *memoryStorage) BatchDelete(ctx context.Context, keys []string, userID string) error {
	return nil
}

func (ms *memoryStorage) Close() error {
	return nil
}

func (ms *memoryStorage) Ping(ctx context.Context) error {
	return nil
}
