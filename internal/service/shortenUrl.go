package service

import (
	"log/slog"
	"strconv"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
)

type ShortenURLService struct {
	repository repository.ShortenURLRepository
}

func NewShortenURLService(r repository.ShortenURLRepository) *ShortenURLService {
	return &ShortenURLService{
		repository: r,
	}
}

func (u *ShortenURLService) Get(key string) string {
	return u.repository.Get(key)
}

func (u *ShortenURLService) Set(url string) string {
	id, err := u.repository.Length()
	if err != nil {
		slog.Error(err.Error())
	}

	key := "key_" + strconv.Itoa(id)
	u.repository.Save(key, url)

	return key
}

func (u *ShortenURLService) Ping() error {
	return u.repository.Ping()
}
