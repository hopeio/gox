/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package container

import "golang.org/x/exp/constraints"

type Node[T any] struct {
	Next  *Node[T]
	Value T
}

type LinkedNode[T any] struct {
	Prev, Next *LinkedNode[T]
	Value      T
}

type KeyNode[K comparable, T any] struct {
	Next  *KeyNode[K, T]
	Key   K
	Value T
}

type LinkedKeyNode[K comparable, T any] struct {
	Prev, Next *LinkedKeyNode[K, T]
	Key        K
	Value      T
}

type OrderedKeyNode[K constraints.Ordered, T any] struct {
	Next  *OrderedKeyNode[K, T]
	Key   K
	Value T
}

type LinkedOrderedKeyNode[K constraints.Ordered, T any] struct {
	Prev, Next *LinkedOrderedKeyNode[K, T]
	Key        K
	Value      T
}
