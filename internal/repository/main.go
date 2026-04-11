package repository

type URLRepository interface {
	Get(key string) string
	Save(key string, value string)
}

type URLStorage struct {
	urls map[string]string
}

func NewURLStorage() *URLStorage {
	return &URLStorage{
		urls: map[string]string{},
	}
}

func (u *URLStorage) Get(key string) string {
	return u.urls[key]
}

func (u *URLStorage) Save(key string, url string) {
	u.urls[key] = url
}
