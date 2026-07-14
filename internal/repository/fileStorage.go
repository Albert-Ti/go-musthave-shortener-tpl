package repository

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/utils"
)

type FileStorage struct {
	element *os.File
	encoder *json.Encoder
	urls    []map[string]string
}

type filRecord struct {
	Uuid        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFileStorage(path string) (*FileStorage, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	urls := []map[string]string{}
	decoder := json.NewDecoder(file)

	for {
		var record filRecord
		err := decoder.Decode(&record)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		urls = append(urls, map[string]string{
			"key":     record.ShortURL,
			"url":     record.OriginalURL,
			"user_id": "unknown",
		})
	}

	return &FileStorage{
		urls:    urls,
		element: file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (fs *FileStorage) Get(ctx context.Context, key string) (string, error) {
	for i := range fs.urls {
		if fs.urls[i]["key"] == key {
			return fs.urls[i]["url"], nil
		}
	}
	return "", errors.New("No Content")
}

func (fs *FileStorage) GetAll(ctx context.Context, userID string) ([]map[string]string, error) {
	var results = make([]map[string]string, 0)

	for i := range fs.urls {
		if fs.urls[i]["user_id"] == userID {
			results = append(results, fs.urls[i])
		}
	}

	return results, nil
}

func (fs *FileStorage) Save(ctx context.Context, key string, url string, userID string) (string, error) {
	record := &filRecord{
		Uuid:        len(fs.urls) + 1,
		ShortURL:    key,
		OriginalURL: url,
	}

	fs.urls = append(fs.urls, map[string]string{
		"key":     key,
		"url":     url,
		"user_id": userID,
	})

	if err := fs.encoder.Encode(&record); err != nil {
		return "", err
	}

	return key, nil
}

func (fs *FileStorage) BatchSave(ctx context.Context, batch []model.BatchReq, userID string) ([]model.BatchResp, error) {
	result := make([]model.BatchResp, len(batch))

	for i, v := range batch {
		key := utils.GenerateUUID()

		record := &filRecord{
			Uuid:        len(fs.urls) + 1,
			ShortURL:    key,
			OriginalURL: v.OriginalURL,
		}

		fs.urls = append(fs.urls, map[string]string{
			"key":     key,
			"url":     v.OriginalURL,
			"user_id": userID,
		})

		if err := fs.encoder.Encode(&record); err != nil {
			return nil, err
		}

		result[i] = model.BatchResp{
			CorrelationID: v.CorrelationID,
			ShortURL:      key,
		}
	}

	return result, nil
}

func (fs *FileStorage) BatchDelete(ctx context.Context, keys []string, userID string) error {
	return nil
}

func (u *FileStorage) Close() error {
	return u.element.Close()
}

func (u *FileStorage) Ping(ctx context.Context) error {
	return nil
}
