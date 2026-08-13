/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package linkedlist

import (
	"errors"
)

type Node[T any] struct {
	data T
	next *Node[T]
}

// linked list
type LinkedList[T comparable] struct {
	head, tail *Node[T]
	size       int
}

// New creates a new instance.
func New[T comparable]() *LinkedList[T] {
	l := LinkedList[T]{}
	return &l
}

// IsEmpty reports whether the condition holds.
func (l *LinkedList[T]) IsEmpty() bool {
	return l.size == 0
}

// Len returns the number of elements.
func (l *LinkedList[T]) Len() int {
	return l.size
}

// Exist reports whether the condition holds.
func (l *LinkedList[T]) Exist(node *Node[T]) bool {
	var p = l.head
	for p != nil {
		if p == node {
			return true
		} else {
			p = p.next
		}
	}
	return false
}

// GetNode returns the value.
func (l *LinkedList[T]) GetNode(e T) *Node[T] {
	var p = l.head
	for p != nil {
		//Find the node holding the data
		if e == p.data {
			return p
		} else {
			p = p.next
		}
	}
	return nil
}

// Append updates or inserts a value.
func (l *LinkedList[T]) Append(e T) {
	//Create a new node for the data
	newNode := Node[T]{}
	newNode.data = e
	newNode.next = nil

	if l.size == 0 {
		l.head = &newNode
		l.tail = &newNode
	} else {
		l.tail.next = &newNode
		l.tail = &newNode
	}
	l.size++
}

// InsertHead updates or inserts a value.
func (l *LinkedList[T]) InsertHead(e T) {
	newNode := Node[T]{}
	newNode.data = e
	newNode.next = l.head
	l.head = &newNode
	if l.size == 0 {
		l.tail = &newNode
	}
	l.size++
}

// InsertAfterNode updates or inserts a value.
func (l *LinkedList[T]) InsertAfterNode(pre *Node[T], e T) error {
	//Insert only if the node exists in the list
	if l.Exist(pre) {
		if pre.next == nil {
			// Append 内部已递增 size，不能再计一次
			l.Append(e)
			return nil
		}
		newNode := Node[T]{}
		newNode.data = e
		newNode.next = pre.next
		pre.next = &newNode
		l.size++
		return nil
	}
	return errors.New("node does not exist in the list")
}

// InsertAfterData updates or inserts a value.
func (l *LinkedList[T]) InsertAfterData(preData T, e T) error {
	var p = l.head
	for p != nil {
		//Find the node holding the data
		if p.data == preData {
			l.InsertAfterNode(p, e)
			return nil
		} else {
			p = p.next
		}
	}
	//Data not found
	return errors.New("data not found in the list; insert failed")
}

// Insert updates or inserts a value.
func (l *LinkedList[T]) Insert(position int, e T) error {
	if position < 0 {
		return errors.New("index must not be negative")
	} else if position == 0 {
		//Insert at the head
		l.InsertHead(e)
		return nil
	} else if position == l.size {
		//Insert at the tail
		l.Append(e)
		return nil
	} else if position > l.size {
		return errors.New("index is out of list length")
	} else {
		//Insert in the middle
		var index int
		var p = l.head
		//Advance pointers one by one
		//position is the new node's index after insert; at position-1 locate the previous node
		for index = 0; index < position-1; index++ {
			p = p.next
		}
		//found
		l.InsertAfterNode(p, e)
		return nil
	}

}

// DeleteNode removes or resets state.
func (l *LinkedList[T]) DeleteNode(node *Node[T]) {
	//node exists
	if l.Exist(node) {
		//If it is the head node
		if node == l.head {
			if node == l.tail {
				l.head = nil
				l.tail = nil
			} else {
				l.head = l.head.next
			}
			l.size--
			return
			//If it is the tail node
		} else if node == l.tail {
			//Find the pointer to the previous node
			var p = l.head
			for p.next != l.tail {
				p = p.next
			}
			p.next = nil
			l.tail = p
			//middle node
		} else {
			var p = l.head
			for p.next != node {
				p = p.next
			}
			p.next = node.next
		}
		l.size--
	}
}

// Delete removes or resets state.
func (l *LinkedList[T]) Delete(e T) {
	p := l.GetNode(e)
	if p == nil {
		return
	}
	l.DeleteNode(p)
}

// traverse performs the operation.
func (l *LinkedList[T]) traverse(f func(T)) {
	var p = l.head
	if l.IsEmpty() {
		return
	}
	for p != nil {
		if f != nil {
			f(p.data)
		}
		p = p.next
	}
}
