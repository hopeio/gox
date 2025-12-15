/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package list

import (
	"github.com/hopeio/gox/container"
)

type List[T any] struct {
	head, tail *container.Node[T]
	size       uint
	zero       T
}

func New[T any]() *List[T] {
	return &List[T]{}
}

func (l *List[T]) Len() uint {
	return l.size
}

func (l *List[T]) Head() *container.Node[T] {
	if l.size == 0 {
		return nil
	}
	return l.head
}

func (l *List[T]) Tail() *container.Node[T] {
	if l.size == 0 {
		return nil
	}
	return l.tail
}

func (l *List[T]) First() (T, bool) {
	if l.size == 0 {
		return l.zero, false
	}
	return l.head.Value, true
}

func (l *List[T]) Last() (T, bool) {
	if l.size == 0 {
		return l.zero, false
	}
	return l.tail.Value, true
}

func (l *List[T]) Pop() (T, bool) {
	if l.size == 0 {
		return l.zero, false
	}

	p := l.head
	l.head = p.Next
	if l.size == 1 {
		l.tail = nil
	}
	l.size--
	return p.Value, true
}

func (l *List[T]) PushFront(v T) {
	node := &container.Node[T]{l.head, v}
	if l.size == 0 {
		l.head = node
		l.tail = node
		l.size++
		return
	}
	l.head = node
	l.size++
}

func (l *List[T]) Push(v T) {
	node := &container.Node[T]{nil, v}
	if l.size == 0 {
		l.head = node
		l.tail = node
		l.size++
		return
	}
	l.tail.Next = node
	l.tail = node
	l.size++
}

func (l *List[T]) PushAt(idx int, v T) {
	if idx < 0 || idx > int(l.size) {
		panic("index out of range")
	}
	node := &container.Node[T]{nil, v}
	if idx == 0 {
		l.PushFront(v)
		return
	}
	if idx == int(l.size) {
		l.Push(v)
		return
	}
	tmpNode := l.head
	for range idx {
		tmpNode = tmpNode.Next
	}
	node.Next = tmpNode.Next
	tmpNode.Next = node
	l.size++
}

type ListIface[T any] interface {
	First() (T, bool)
	Last() (T, bool)
	Pop() (T, bool)
	Push(v T)
	Len() uint
}
