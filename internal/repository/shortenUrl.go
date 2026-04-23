package repository

import (
	"encoding/json"
	"os"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/config"
)

type ShortenURLRepository interface {
	Get(key string) string
	Save(key string, value string, count uint)
}

type fileUrlRecord struct {
	Uuid        uint   `json:"uuid"`
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
}

type ShortenURLStorage struct {
	urls map[string]string
	file *os.File
}

func NewShortenURLStorage() (*ShortenURLStorage, error) {
	file, err := os.OpenFile(config.Envs.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &ShortenURLStorage{
		urls: map[string]string{},
		file: file,
	}, nil
}

func (u *ShortenURLStorage) Get(key string) string {
	return u.urls[key]
}

func (u *ShortenURLStorage) Save(key string, url string, count uint) {
	u.urls[key] = url

	record := &fileUrlRecord{
		Uuid:        count,
		ShortUrl:    key,
		OriginalUrl: url,
	}

	json.NewEncoder(u.file).Encode(record)
}
