package repository

type ShortenURLRepository interface {
	Get(key string) string
	Save(key string, value string)
}

type ShortenURLStorage struct {
	urls map[string]string
}

func NewShortenURLStorage() *ShortenURLStorage {
	return &ShortenURLStorage{
		urls: map[string]string{},
	}
}

func (u *ShortenURLStorage) Get(key string) string {
	return u.urls[key]
}

func (u *ShortenURLStorage) Save(key string, url string) {
	u.urls[key] = url
}
