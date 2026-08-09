package utility

import "sync"

type CType[T any] struct {
	mutex sync.Mutex
	value T
}

func (ct *CType[T]) Set(value T) {
	ct.mutex.Lock()

	ct.value = value

	ct.mutex.Unlock()
}

func (ct *CType[T]) Get() T {
	ct.mutex.Lock()

	value := ct.value

	ct.mutex.Unlock()

	return value
}
