package gox

func TernaryOperator[T any](v bool, a, b T) T {
	if v {
		return a
	}
	return b
}

func Match[T any](yes bool, a, b T) T {
	if yes {
		return a
	}
	return b
}

func Pointer[T any](t T) *T {
	return &t
}

func Zero[T any]() T {
	return *new(T)
}

func Nil[T any]() T {
	return *(*T)(nil)
}

func Zero2[T any]() T {
	var zero T
	return zero
}

// 两种转换,any(i).(T), T(any(i))
