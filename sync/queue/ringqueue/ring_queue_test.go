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

// TestRingQueue_CounterWraparound 白盒验证 putPos/getPos 在 uint32 回绕处仍正确计数。
func TestRingQueue_CounterWraparound(t *testing.T) {
	q := New[int](8)
	// 手动把位置计数推到回绕边界（保持 putPos == getPos 即空队列）
	start := ^uint32(0) - 2 // MaxUint32 - 2
	q.putPos, q.getPos = start, start
	for i := range q.cache {
		// 槽位序号 = 下一个会命中该槽的 pos（可能已回绕），与 New 中槽 0 的特殊处理同理
		no := start&^q.capMod + uint32(i)
		if no <= start {
			no += q.capacity
		}
		q.cache[i].putNo = no
		q.cache[i].getNo = no
	}
	for i := 0; i < 6; i++ {
		if ok, _ := q.Put(i); !ok {
			t.Fatalf("Put %d failed near wraparound", i)
		}
		if got := q.Quantity(); got != uint32(i+1) {
			t.Fatalf("Quantity after put %d = %d, want %d", i, got, i+1)
		}
	}
	for i := 0; i < 6; i++ {
		v, ok, _ := q.Get()
		if !ok || v != i {
			t.Fatalf("Get %d = (%v,%v)", i, v, ok)
		}
	}
	if q.Quantity() != 0 {
		t.Fatalf("Quantity = %d, want 0", q.Quantity())
	}
}
