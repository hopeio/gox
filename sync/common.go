/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package sync

import (
	"sync/atomic"
	"unsafe"
)

// Node has better performance than using DirectItem.
type Node[T any] struct {
	Next  atomic.Pointer[Node[T]]
	Value T
}

// LoadNode atomically loads a Node pointer.
func LoadNode[T any](p *unsafe.Pointer) *Node[T] {
	return (*Node[T])(atomic.LoadPointer(p))
}

// CasNode performs an atomic compare-and-swap on a Node pointer.
func CasNode[T any](p *unsafe.Pointer, old, new *Node[T]) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}

type DirectItem struct {
	Next  unsafe.Pointer
	Value any
}

// LoadItem atomically loads a DirectItem pointer.
func LoadItem(p *unsafe.Pointer) *DirectItem {
	return (*DirectItem)(atomic.LoadPointer(p))
}

// CasItem performs an atomic compare-and-swap on a DirectItem pointer.
func CasItem(p *unsafe.Pointer, old, new *DirectItem) bool {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
