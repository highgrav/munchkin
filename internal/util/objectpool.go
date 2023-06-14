package util

type ObjectPool[T any] struct {
	size int
	Pool chan *T
}

func NewObjectPool[T any](size int) *ObjectPool[T] {
	oc := &ObjectPool[T]{
		size: size,
		Pool: make(chan *T, size),
	}
	return oc
}
