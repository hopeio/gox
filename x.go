package gox

// TernaryOperator ...
func TernaryOperator[T any](v bool, a, b T) T {
	if v {
		return a
	}
	return b
}

// Match ...
func Match[T any](yes bool, a, b T) T {
	if yes {
		return a
	}
	return b
}

// Pointer ...
func Pointer[T any](t T) *T {
	return &t
}

// Zero ...
func Zero[T any]() T {
	var zero T
	return zero
}

// Nil ...
func Nil[T any]() *T {
	return (*T)(nil)
}

// zero ...
func zero[T any]() T {
	return *new(T)
}

// 两种转换,any(i).(T), T(any(i))
