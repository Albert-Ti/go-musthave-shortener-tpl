package pool

import "sync"

// Resetter определяет интерфейс для объектов, которые могут быть сброшены.
type Resetter interface {
	Reset()
}

// Pool — типобезопасная generic-обёртка над sync.Pool.
type Pool[T Resetter] struct {
	sp sync.Pool
}

// New создает и возвращает новый пул объектов.
// newFunc — фабричная функция, аналог поля New у sync.Pool,
// вызывается, когда в пуле нет свободных объектов.
func New[T Resetter](newFunc func() T) *Pool[T] {
	p := &Pool[T]{}

	if newFunc != nil {
		p.sp.New = func() any {
			return newFunc()
		}
	}
	return p
}

// Get возвращает объект из пула. Если свободных нет — вызывает newFunc.
func (p *Pool[T]) Get() T {
	el := p.sp.Get()

	if el == nil {
		var zero T
		return zero
	}

	return el.(T)
}

// Put сбрасывает объект и возвращает его в пул.
func (p *Pool[T]) Put(el T) {
	el.Reset()
	p.sp.Put(el)
}
