package repository

import (
	"encoding/json"
	"io"
	"os"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
)

type ShortenURLRepository interface {
	Get(key string) string
	Save(key string, value string)
	Length() int
}

type fileUrlRecord struct {
	Uuid        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	element *os.File
	encoder *json.Encoder
}

type ShortenURLStorage struct {
	urls map[string]string
	file FileStorage
}

func NewShortenURLStorage() (*ShortenURLStorage, error) {
	file, err := os.OpenFile(config.Envs.FileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	urls := map[string]string{}
	decoder := json.NewDecoder(file)

	for {
		var record fileUrlRecord
		err := decoder.Decode(&record)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		urls[record.ShortURL] = record.OriginalURL
	}

	return &ShortenURLStorage{
		urls: urls,
		file: FileStorage{
			element: file,
			encoder: json.NewEncoder(file),
		},
	}, nil
}

func (u *ShortenURLStorage) Get(key string) string {
	return u.urls[key]
}

func (u *ShortenURLStorage) Save(key string, url string) {
	record := &fileUrlRecord{
		Uuid:        u.Length(),
		ShortURL:    key,
		OriginalURL: url,
	}

	u.urls[key] = url

	if err := u.file.encoder.Encode(&record); err != nil {
		panic(err)
	}
}

func (u *ShortenURLStorage) Length() int {
	return len(u.urls) + 1
}

func (u *ShortenURLStorage) Close() error {
	return u.file.element.Close()
}
func (u *ShortenURLStorage) Remove() {
	os.Remove(u.file.element.Name())
}
