package queue

import (
	"testing"
)

func TestNew(t *testing.T) {
	q := New()
	if q == nil {
		t.Error("New() returned nil")
	}
	if q.Len() != 0 {
		t.Errorf("Len() = %d, want 0", q.Len())
	}
}

func TestQueue_EnqueueDequeue(t *testing.T) {
	q := New()
	q.Enqueue("a")
	q.Enqueue("b")
	q.Enqueue("c")

	if q.Len() != 3 {
		t.Errorf("Len() = %d, want 3", q.Len())
	}

	v := q.Dequeue()
	if v != "a" {
		t.Errorf("Dequeue() = %v, want a", v)
	}
	v = q.Dequeue()
	if v != "b" {
		t.Errorf("Dequeue() = %v, want b", v)
	}
	v = q.Dequeue()
	if v != "c" {
		t.Errorf("Dequeue() = %v, want c", v)
	}
	if q.Len() != 0 {
		t.Errorf("Len() = %d, want 0", q.Len())
	}
	// Dequeue from empty queue
	v = q.Dequeue()
	if v != nil {
		t.Errorf("Dequeue() on empty queue = %v, want nil", v)
	}
}

func TestQueue_Peek(t *testing.T) {
	q := New()
	q.Enqueue(10)
	q.Enqueue(20)

	v := q.Peek()
	if v != 10 {
		t.Errorf("Peek() = %v, want 10", v)
	}
	// Peek should not remove the element
	if q.Len() != 2 {
		t.Errorf("Len() after Peek = %d, want 2", q.Len())
	}
}

func TestQueue_Int(t *testing.T) {
	q := New()
	for i := 0; i < 100; i++ {
		q.Enqueue(i)
	}
	for i := 0; i < 100; i++ {
		v := q.Dequeue()
		if v != i {
			t.Errorf("Dequeue() = %v, want %d", v, i)
		}
	}
}
