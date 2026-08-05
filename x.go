package gox

// TernaryOperator returns a if v is true; otherwise it returns b.
func TernaryOperator[T any](v bool, a, b T) T {
	if v {
		return a
	}
	return b
}

// Match returns a if yes is true; otherwise it returns b.
func Match[T any](yes bool, a, b T) T {
	if yes {
		return a
	}
	return b
}

// Pointer returns a pointer to t.
func Pointer[T any](t T) *T {
	return &t
}

// Zero returns the zero value for T.
func Zero[T any]() T {
	var zero T
	return zero
}

// Nil returns a nil pointer of type *T.
func Nil[T any]() *T {
	return (*T)(nil)
}

// zero returns the zero value for T (unexported helper).
func zero[T any]() T {
	return *new(T)
}

// Two conversion forms: any(i).(T) and T(any(i)).
