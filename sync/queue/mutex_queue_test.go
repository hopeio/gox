package queue

import "testing"

func TestNewMutexQueue(t *testing.T) {
	q := NewMutexQueue[int]()
	if q == nil {
		t.Fatal("NewMutexQueue() returned nil")
	}
}

func TestMutexQueue_EnqueueDequeue(t *testing.T) {
	q := NewMutexQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	val, ok := q.Dequeue()
	if !ok || val != 1 {
		t.Errorf("Dequeue() = %v, %v, want 1, true", val, ok)
	}

	val, ok = q.Dequeue()
	if !ok || val != 2 {
		t.Errorf("Dequeue() = %v, %v, want 2, true", val, ok)
	}

	val, ok = q.Dequeue()
	if !ok || val != 3 {
		t.Errorf("Dequeue() = %v, %v, want 3, true", val, ok)
	}

	_, ok = q.Dequeue()
	if ok {
		t.Error("Dequeue() on empty queue should return false")
	}
}
