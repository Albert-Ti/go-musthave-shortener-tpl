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
//
// Пример использования:
//
//	type ExampleStruct struct{}
//	func (e *ExampleStruct) Reset() {}
//
//	p := pool.New[*ExampleStruct](func() *ExampleStruct {
//	    return &ExampleStruct{}
//	})
func New[T Resetter](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		sp: sync.Pool{
			New: func() any {
				return newFunc()
			},
		},
	}
}

// Get возвращает объект из пула. Если свободных нет — вызывает newFunc.
func (p *Pool[T]) Get() T {
	return p.sp.Get().(T)
}

// Put сбрасывает объект и возвращает его в пул.
func (p *Pool[T]) Put(el T) {
	el.Reset()
	p.sp.Put(el)
}
