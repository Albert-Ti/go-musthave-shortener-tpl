package service

import (
	"strconv"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
)

type ShortenURLService struct {
	repository repository.ShortenURLRepository
	count      int
}

func NewShortenURLService(r repository.ShortenURLRepository) *ShortenURLService {
	return &ShortenURLService{
		repository: r,
		count:      int(r.Length()),
	}
}

func (u *ShortenURLService) Get(key string) string {
	return u.repository.Get(key)
}

func (u *ShortenURLService) Set(url string) string {
	key := "key_" + strconv.Itoa(u.repository.Length())
	u.repository.Save(key, url)

	return key
}
