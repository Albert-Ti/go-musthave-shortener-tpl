// Pool представляет универсальный пул объектов с generic параметром,
// который ограничен типами, реализующими интерфейс Resetter.
package pool

import (
	"sync"
)

// Resetter определяет интерфейс для объектов, которые могут быть сброшены.
type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	mu   sync.Mutex
	free []T
}

// NewPool создает и возвращает новый пул объектов.
//
// Пример использования:
//
//	struct ExampleStruct {}
//	func (e *ExampleStruct) Reset() {}
//
//	pool = pool.NewPool[*ExampleStruct]()
func NewPool[T Resetter]() *Pool[T] {
	return &Pool[T]{}
}

// Get возвращает свободный объект из пула (LIFO).
// Возвращает объект и true, если объект доступен, иначе - zero-значение и false.
func (p *Pool[T]) Get() (T, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var zero T
	if len(p.free) == 0 {
		return zero, false
	}

	el := p.free[len(p.free)-1]     // берем
	p.free = p.free[:len(p.free)-1] // удаляем
	return el, true
}

// Put возвращает объект в пул (LIFO) и автоматически вызывает его метод Reset().
func (p *Pool[T]) Put(el T) {
	el.Reset()
	p.mu.Lock()
	defer p.mu.Unlock()
	p.free = append(p.free, el)
}
