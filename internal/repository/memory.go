package repository

type MemoryStorage struct {
	urls map[string]string
}

func NewMemoryStorage() (*MemoryStorage, error) {
	return &MemoryStorage{urls: map[string]string{}}, nil
}

func (u *MemoryStorage) Get(key string) (string, error) {
	return u.urls[key], nil
}

func (u *MemoryStorage) Save(key string, url string) error {
	u.urls[key] = url

	return nil
}

func (u *MemoryStorage) Length() (int, error) {
	return len(u.urls), nil
}

func (u *MemoryStorage) Close() error {
	return nil
}

func (u *MemoryStorage) Ping() error {
	return nil
}
