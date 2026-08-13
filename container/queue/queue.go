package queue

import (
	"container/list"
)

// Queue is a queue backed by container/list
type Queue struct {
	list *list.List
}

// New creates a new instance.
func New() *Queue {
	return &Queue{
		list: list.New(),
	}
}

// Enqueue updates or inserts a value.
func (q *Queue) Enqueue(item any) {
	q.list.PushBack(item)
}

// Peek returns the result, or nil when the queue is empty.
func (q *Queue) Peek() any {
	front := q.list.Front()
	if front == nil {
		return nil
	}
	return front.Value
}

// Dequeue removes or resets state.
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
