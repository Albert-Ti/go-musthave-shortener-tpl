package service

import (
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

func (u *ShortenURLService) Get(key string) (string, error) {
	return u.repository.Get(key)
}

func (u *ShortenURLService) Set(url string) (string, error) {
	id, err := u.repository.Length()
	if err != nil {
		return "", err
	}

	key := "key_" + strconv.Itoa(id+1)
	if e := u.repository.Save(key, url); e != nil {
		return "", e
	}

	return key, nil
}

func (u *ShortenURLService) Ping() error {
	return u.repository.Ping()
}
