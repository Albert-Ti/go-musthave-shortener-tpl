package pool

import (
	"sync"
)

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	mu   sync.Mutex
	Free []T
}

func NewPool[T Resetter]() *Pool[T] {
	return &Pool[T]{}
}

func (p *Pool[T]) Get() (T, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var zero T
	if len(p.Free) == 0 {
		return zero, false
	}

	el := p.Free[len(p.Free)-1]     // берем
	p.Free = p.Free[:len(p.Free)-1] // удаляет
	return el, true
}

func (p *Pool[T]) Put(el T) {
	el.Reset()
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Free = append(p.Free, el)
}
