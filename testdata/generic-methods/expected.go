package genericmethods

type Mapper[T any] struct{}

func (Mapper[T]) Map[U any](value T, fn func(T) U) U { return fn(value) }
