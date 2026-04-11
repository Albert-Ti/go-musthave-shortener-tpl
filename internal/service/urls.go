package service

import (
	"strconv"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/repository"
)

type URLService struct {
	repository repository.URLRepository
	Count      uint
}

func NewURLService(r repository.URLRepository) *URLService {
	return &URLService{
		repository: r,
		Count:      1,
	}
}

func (u *URLService) Get(key string) string {
	return u.repository.Get(key)
}

func (u *URLService) Set(url string) string {
	key := "key_" + strconv.Itoa(int(u.Count))
	u.repository.Save(key, url)
	u.Count++

	return key
}
