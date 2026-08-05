package queue

import (
	"container/list"
)

// Queue 基于 container/list 实现的队列
type Queue struct {
	list *list.List
}

// New ...
func New() *Queue {
	return &Queue{
		list: list.New(),
	}
}

// Enqueue ...
func (q *Queue) Enqueue(item any) {
	q.list.PushBack(item)
}

// Peek ...
func (q *Queue) Peek() any {
	front := q.list.Front()
	return front.Value
}

// Dequeue ...
func (q *Queue) Dequeue() any {
	front := q.list.Front()
	if front == nil {
		return nil
	}
	q.list.Remove(front)
	return front.Value
}

// Len returns the number of elements.
func (q *Queue) Len() int {
	return q.list.Len()
}
