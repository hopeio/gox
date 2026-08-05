/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package maps

import "maps"

// Map returns the result.
func Map[M ~map[K]V, K comparable, V, T any](m M, subValue func(K, V) T) []T {
	r := make([]T, 0, len(m))
	for k, v := range m {
		r = append(r, subValue(k, v))
	}
	return r
}

// Keys returns the result.
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	r := make([]K, 0, len(m))
	for k := range m {
		r = append(r, k)
	}
	return r
}

// KeysMap returns the result.
func KeysMap[M ~map[K]V, K comparable, V, T any](m M, transform func(K) T) []T {
	r := make([]T, 0, len(m))
	for k := range maps.Keys(m) {
		r = append(r, transform(k))
	}
	return r
}

// Values returns the result.
func Values[M ~map[K]V, K comparable, V any](m M) []V {
	r := make([]V, 0, len(m))
	for _, v := range m {
		r = append(r, v)
	}
	return r
}

// ValuesMap returns the result.
func ValuesMap[M ~map[K]V, K comparable, V, T any](m M, transform func(V) T) []T {
	r := make([]T, 0, len(m))
	for v := range maps.Values(m) {
		r = append(r, transform(v))
	}
	return r
}

// ForEach performs the operation.
func ForEach[M ~map[K]V, K comparable, V any](m M, handle func(K, V)) {
	for k, v := range m {
		handle(k, v)
	}
}

// ForEachValue performs the operation.
func ForEachValue[M ~map[K]V, K comparable, V any](m M, handle func(v V)) {
	for _, v := range m {
		handle(v)
	}
}

// ForEachKey performs the operation.
func ForEachKey[M ~map[K]V, K comparable, V any](m M, handle func(v K)) {
	for k, _ := range m {
		handle(k)
	}
}

// MultiKeys returns the result.
func MultiKeys[M ~map[K]V, K comparable, V any](maps ...M) []K {
	r := make([]K, 0, len(maps))
	for _, m := range maps {
		for k := range m {
			r = append(r, k)
		}
	}
	return r
}

// MultiValues returns the result.
func MultiValues[M ~map[K]V, K comparable, V any](maps ...M) []V {
	r := make([]V, 0, len(maps))
	for _, m := range maps {
		for _, v := range m {
			r = append(r, v)
		}
	}
	return r
}

// Merge returns the result.
func Merge[M ~map[K]V, K comparable, V any](maps ...M) M {
	r := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			r[k] = v
		}
	}
	return r
}

// Transform returns the result.
func Transform[M1 ~map[K1]V1, K1 comparable, V1 any, M2 ~map[K2]V2, K2 comparable, V2 any](m M1, transform func(K1, V1) (K2, V2)) M2 {
	m2 := make(M2)
	for k1, v1 := range m {
		k2, v2 := transform(k1, v1)
		m2[k2] = v2
	}
	return m2
}
