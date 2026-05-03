package repository

type MemoryStorage struct {
	urls map[string]string
}

func NewMemoryStorage() (*MemoryStorage, error) {
	return &MemoryStorage{urls: map[string]string{}}, nil
}

func (u *MemoryStorage) Get(key string) string {
	return u.urls[key]
}

func (u *MemoryStorage) Save(key string, url string) {
	u.urls[key] = url
}

func (u *MemoryStorage) Length() int {
	return len(u.urls) + 1
}

func (u *MemoryStorage) Close() error {
	return u.Close()
}
