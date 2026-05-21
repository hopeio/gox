package ringqueue

import "testing"

func TestNew(t *testing.T) {
	q := New[int](8)
	if q == nil {
		t.Fatal("New() returned nil")
	}
	if q.Capaciity() < 8 {
		t.Errorf("Capacity = %d, want >= 8", q.Capaciity())
	}
}

func TestRingQueue_PutGet(t *testing.T) {
	q := New[int](8)

	ok, _ := q.Put(42)
	if !ok {
		t.Error("Put() should succeed")
	}

	val, ok, _ := q.Get()
	if !ok {
		t.Error("Get() should succeed")
	}
	if val != 42 {
		t.Errorf("Get() = %d, want 42", val)
	}
}

func TestRingQueue_FIFO(t *testing.T) {
	q := New[int](8)

	q.Put(1)
	q.Put(2)
	q.Put(3)

	val, ok, _ := q.Get()
	if !ok || val != 1 {
		t.Errorf("Get() = %d, %v, want 1, true", val, ok)
	}

	val, ok, _ = q.Get()
	if !ok || val != 2 {
		t.Errorf("Get() = %d, %v, want 2, true", val, ok)
	}

	val, ok, _ = q.Get()
	if !ok || val != 3 {
		t.Errorf("Get() = %d, %v, want 3, true", val, ok)
	}
}

func TestRingQueue_Quantity(t *testing.T) {
	q := New[int](8)

	if q.Quantity() != 0 {
		t.Errorf("Quantity() = %d, want 0", q.Quantity())
	}

	q.Put(1)
	q.Put(2)

	if q.Quantity() != 2 {
		t.Errorf("Quantity() = %d, want 2", q.Quantity())
	}

	q.Get()

	if q.Quantity() != 1 {
		t.Errorf("Quantity() = %d, want 1", q.Quantity())
	}
}

func TestRingQueue_GetEmpty(t *testing.T) {
	q := New[int](8)
	_, ok, _ := q.Get()
	if ok {
		t.Error("Get() on empty queue should return false")
	}
}

func TestRingQueue_PutsGets(t *testing.T) {
	q := New[int](16)

	puts, _ := q.Puts([]int{1, 2, 3, 4})
	if puts != 4 {
		t.Errorf("Puts() = %d, want 4", puts)
	}

	values := make([]int, 3)
	gets, _ := q.Gets(values)
	if gets != 3 {
		t.Errorf("Gets() = %d, want 3", gets)
	}
	if values[0] != 1 || values[1] != 2 || values[2] != 3 {
		t.Errorf("Gets values = %v, want [1 2 3]", values)
	}
}

func TestRingQueue_String(t *testing.T) {
	q := New[int](8)
	s := q.String()
	if len(s) == 0 {
		t.Error("String() should not be empty")
	}
}
