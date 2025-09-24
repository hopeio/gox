/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package linkedlist

import (
	"errors"
)

// 链表
type LinkedList[T comparable] struct {
	head, tail *Node[T]
	size       int
}

// 新建空链表，即创建Node指针head，用来指向链表第一个结点，初始为空
func New[T comparable]() *LinkedList[T] {
	l := LinkedList[T]{}
	return &l
}

// 是否为空链表
func (l *LinkedList[T]) IsEmpty() bool {
	return l.size == 0
}

// 获取链表长度
func (l *LinkedList[T]) Len() int {
	return l.size
}

// 是否含有指定结点
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

// 获取含有指定数据的第一个结点
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

// 在链表尾部添加数据
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

// 在链表头部插入数据
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

// 在指定结点后面插入数据
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

// 在第一次出现指定数据的结点后插入数据,若链表中无该数据，返回false
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

// 在指定下标处插入数据
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

// 删除指定结点
func (l *LinkedList[T]) DeleteNode(node *Node[T]) {
	//存在该结点
	if l.Exist(node) {
		//如果是头部结点
		if node == l.head {
			l.head = l.head.next
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

// 删除第一个含指定数据的结点
func (l *LinkedList[T]) Delete(e T) {
	p := l.GetNode(e)
	if p == nil {
		return
	}
	l.DeleteNode(p)
}

// 遍历链表
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
