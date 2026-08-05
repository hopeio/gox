/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package queue

type IQueue[T any] interface {
	// Return the current list length.
	Len() int
	// Return the current list capacity.
	Capacity() int
	// Return the current list head.
	Front() (T, bool)
	// Return the current list tail.
	Tail() (T, bool)
	// Enqueue.
	Enqueue(value T) bool
	// Dequeue.
	Dequeue() T
}
