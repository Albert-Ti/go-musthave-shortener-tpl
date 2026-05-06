package repository

import (
	"encoding/json"
	"io"
	"os"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/model"
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

func (u *FileStorage) Get(key string) (string, error) {
	return u.urls[key], nil
}

func (u *FileStorage) Save(key string, url string) error {
	len, _ := u.Length()

	record := &filRecord{
		Uuid:        len + 1,
		ShortURL:    key,
		OriginalURL: url,
	}

	u.urls[key] = url

	if err := u.encoder.Encode(&record); err != nil {
		return err
	}

	return nil
}

func (u *FileStorage) BatchSave(keys []string, batch []model.JSONBatchReq) error {
	len, _ := u.Length()

	for i, v := range batch {

		record := &filRecord{
			Uuid:        len + 1,
			ShortURL:    keys[i],
			OriginalURL: v.OriginalURL,
		}
		len++

		u.urls[keys[i]] = v.OriginalURL

		if err := u.encoder.Encode(&record); err != nil {
			return err
		}
	}

	return nil
}

func (u *FileStorage) Length() (int, error) {
	return len(u.urls), nil
}

func (u *FileStorage) Close() error {
	return u.element.Close()
}

func (u *FileStorage) Ping() error {
	return nil
}
