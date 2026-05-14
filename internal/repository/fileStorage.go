package repository

import (
	"context"
	"encoding/json"
	"io"
	"os"
)

type FileStorage struct {
	element *os.File
	encoder *json.Encoder
	urls    map[string]string
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

	urls := map[string]string{}
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
		urls[record.ShortURL] = record.OriginalURL
	}

	return &FileStorage{
		urls:    urls,
		element: file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (fs *FileStorage) Get(ctx context.Context, key string) (string, error) {
	return fs.urls[key], nil
}

func (fs *FileStorage) GetAll(ctx context.Context) ([]map[string]string, error) {

	var results = make([]map[string]string, 0)
	for k, v := range fs.urls {
		results = append(results, map[string]string{
			"key": k,
			"url": v,
		})
	}
	return results, nil
}

func (fs *FileStorage) Save(ctx context.Context, key string, url string) (string, error) {

	record := &filRecord{
		Uuid:        len(fs.urls) + 1,
		ShortURL:    key,
		OriginalURL: url,
	}

	fs.urls[key] = url

	if err := fs.encoder.Encode(&record); err != nil {
		return "", err
	}

	return key, nil
}

func (u *FileStorage) Close() error {
	return u.element.Close()
}

func (u *FileStorage) Ping(ctx context.Context) error {
	return nil
}
