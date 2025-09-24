/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package node

import "golang.org/x/exp/constraints"

type Node[T any] struct {
	Next  *Node[T]
	Value T
}

type LinkedNode[T any] struct {
	Prev, Next *LinkedNode[T]
	Value      T
}

type KNode[K comparable, T any] struct {
	Next  *KNode[K, T]
	Key   K
	Value T
}

type LinkedKNode[K comparable, T any] struct {
	Prev, Next *LinkedKNode[K, T]
	Key        K
	Value      T
}

type OrdKNode[K constraints.Ordered, T any] struct {
	Next  *OrdKNode[K, T]
	Key   K
	Value T
}

type LinkedOrdKNode[K constraints.Ordered, T any] struct {
	Prev, Next *LinkedOrdKNode[K, T]
	Key        K
	Value      T
}
