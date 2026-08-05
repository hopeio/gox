/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package slices

type Collector[S ~[]T, T any] struct {
}

// Builder creates and returns a new instance.
func (c Collector[S, T]) Builder() *S {
	s := make(S, 0)
	return &s
}

// Append updates or inserts a value.
func (c Collector[S, T]) Append(builder *S, element T) {
	*builder = append(*builder, element)
}

// Finish returns the result.
func (c Collector[S, T]) Finish(builder *S) S {
	return *builder
}
