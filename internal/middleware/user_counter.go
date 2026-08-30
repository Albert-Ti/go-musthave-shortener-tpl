package middleware

import "sync"

// UserCounter - счетчик уникальных пользователей(потокобезопасный).
type UserCounter struct {
	mu    sync.Mutex
	users map[string]struct{}
}

func NewUserCounter() *UserCounter {
	return &UserCounter{users: make(map[string]struct{})}
}

func (u *UserCounter) Add(userID string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	u.users[userID] = struct{}{}
}

func (c *UserCounter) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return len(c.users)
}
