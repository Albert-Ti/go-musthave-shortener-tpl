package service

import (
	"strconv"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
)

type ShortenURLService struct {
	repository repository.ShortenURLRepository
	Count      uint
}

func NewShortenURLService(r repository.ShortenURLRepository) *ShortenURLService {
	return &ShortenURLService{
		repository: r,
		Count:      1,
	}
}

func (u *ShortenURLService) Get(key string) string {
	return u.repository.Get(key)
}

func (u *ShortenURLService) Set(url string) string {
	key := "key_" + strconv.Itoa(int(u.Count))
	u.repository.Save(key, url, u.Count)

	u.Count++

	return key
}
