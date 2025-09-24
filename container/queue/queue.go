package queue

import (
	"container/list"
)

// Queue 基于 container/list 实现的队列
type Queue struct {
	list *list.List
}

// New 创建一个新的基于 list 的队列
func New() *Queue {
	return &Queue{
		list: list.New(),
	}
}

// Enqueue 向队列尾部添加元素
func (q *Queue) Enqueue(item any) {
	q.list.PushBack(item)
}

// Dequeue 从队列头部移除并返回元素
func (q *Queue) Peek() any {
	front := q.list.Front()
	return front.Value
}

// Dequeue 从队列头部移除并返回元素
func (q *Queue) Dequeue() any {
	front := q.list.Front()
	if front == nil {
		return nil
	}
	q.list.Remove(front)
	return front.Value
}

// Size 返回队列的大小
func (q *Queue) Len() int {
	return q.list.Len()
}
