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

// 链表
type LinkedList[T comparable] struct {
	head, tail *Node[T]
	size       int
}

// New ...
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

// Exist ...
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

// GetNode ...
func (l *LinkedList[T]) GetNode(e T) *Node[T] {
	var p = l.head
	for p != nil {
		//找到该数据所在结点
		if e == p.data {
			return p
		} else {
			p = p.next
		}
	}
	return nil
}

// Append ...
func (l *LinkedList[T]) Append(e T) {
	//为数据创建新结点
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

// InsertHead ...
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

// InsertAfterNode ...
func (l *LinkedList[T]) InsertAfterNode(pre *Node[T], e T) error {
	//如果链表中存在该结点，才进行插入
	if l.Exist(pre) {
		newNode := Node[T]{}
		newNode.data = e
		if pre.next == nil {
			l.Append(e)
		} else {
			newNode.next = pre.next
			pre.next = &newNode
		}
		l.size++
		return nil
	}
	return errors.New("链表中不存在该结点")
}

// InsertAfterData ...
func (l *LinkedList[T]) InsertAfterData(preData T, e T) error {
	var p = l.head
	for p != nil {
		//找到该数据所在结点
		if p.data == preData {
			l.InsertAfterNode(p, e)
			return nil
		} else {
			p = p.next
		}
	}
	//没有找到该数据
	return errors.New("链表中没有该数据，插入失败")
}

// Insert ...
func (l *LinkedList[T]) Insert(position int, e T) error {
	if position < 0 {
		return errors.New("下标不能为负数")
	} else if position == 0 {
		//在头部插入
		l.InsertHead(e)
		return nil
	} else if position == l.size {
		//在尾部插入
		l.Append(e)
		return nil
	} else if position > l.size {
		return errors.New("指定下标超出链表长度")
	} else {
		//在中间插入
		var index int
		var p = l.head
		//逐个移动指针
		//position是插入后新结点的下标，position-1时需要定位到的其前一个结点的下标
		for index = 0; index < position-1; index++ {
			p = p.next
		}
		//找到
		l.InsertAfterNode(p, e)
		return nil
	}

}

// DeleteNode ...
func (l *LinkedList[T]) DeleteNode(node *Node[T]) {
	//存在该结点
	if l.Exist(node) {
		//如果是头部结点
		if node == l.head {
			if node == l.tail {
				l.head = nil
				l.tail = nil
			} else {
				l.head = l.head.next
			}
			return
			//如果是尾部结点
		} else if node == l.tail {
			//寻找指向其前一个结点的指针
			var p = l.head
			for p.next != l.tail {
				p = p.next
			}
			p.next = nil
			l.tail = p
			//中间结点
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

// Delete ...
func (l *LinkedList[T]) Delete(e T) {
	p := l.GetNode(e)
	if p == nil {
		return
	}
	l.DeleteNode(p)
}

// traverse ...
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
