package sync

import "sync"

// Pool 泛型对象池
type Pool[T any] sync.Pool

// NewTypedPool 创建泛型对象池
func NewTypedPool[T any](newFunc func() T) *Pool[T] {
	return (*Pool[T])(&sync.Pool{
		New: func() any { return newFunc() },
	})
}

// Get 获取对象
func (p *Pool[T]) Get() T {
	return (*sync.Pool)(p).Get().(T)
}

// Put 放回对象
func (p *Pool[T]) Put(x T) {
	(*sync.Pool)(p).Put(x)
}
