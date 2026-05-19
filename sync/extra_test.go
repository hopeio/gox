package sync

import (
	"sync/atomic"
	"testing"
)

func TestAddFloat64_Basic(t *testing.T) {
	var val float64 = 1.5
	AddFloat64(&val, 2.5)
	if val != 4.0 {
		t.Errorf("AddFloat64: val = %f, want 4.0", val)
	}
	AddFloat64(&val, -1.0)
	if val != 3.0 {
		t.Errorf("AddFloat64: val = %f, want 3.0", val)
	}
}

func TestAddFloat32(t *testing.T) {
	var val float32 = 1.5
	AddFloat32(&val, 2.5)
	if val != 4.0 {
		t.Errorf("AddFloat32: val = %f, want 4.0", val)
	}
}

func TestSubUint32(t *testing.T) {
	var val uint32 = 10
	SubUint32(&val, 3)
	if val != 7 {
		t.Errorf("SubUint32: val = %d, want 7", val)
	}
}

func TestSubUint64(t *testing.T) {
	var val uint64 = 100
	SubUint64(&val, 30)
	if val != 70 {
		t.Errorf("SubUint64: val = %d, want 70", val)
	}
}

func TestAddFloat64_Concurrent(t *testing.T) {
	var val float64 = 0
	var wg = make(chan struct{}, 100)
	for i := 0; i < 100; i++ {
		go func() {
			AddFloat64(&val, 1.0)
			wg <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-wg
	}
	if val != 100.0 {
		t.Errorf("Concurrent AddFloat64: val = %f, want 100.0", val)
	}
}

func TestPool(t *testing.T) {
	p := NewPool[int](func() int {
		return 42
	})
	v := p.Get()
	if v != 42 {
		t.Errorf("Pool.Get() = %d, want 42", v)
	}
	p.Put(100)
	v = p.Get()
	// After Put, we might get 100 or 42 depending on GC
	_ = v
}

func TestSubUint32_Underflow(t *testing.T) {
	var val uint32 = 5
	result := SubUint32(&val, 10)
	// uint32 underflow wraps around: 5 - 10 = -5 as uint32
	if result == 0 {
		t.Errorf("SubUint32 underflow should not be 0")
	}
}

func TestConcurrentAtomicOps(t *testing.T) {
	var val uint64 = 1000
	done := make(chan struct{}, 2)
	go func() {
		for i := 0; i < 100; i++ {
			SubUint64(&val, 1)
		}
		done <- struct{}{}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			atomic.AddUint64(&val, 1)
		}
		done <- struct{}{}
	}()
	<-done
	<-done
	if val != 1000 {
		t.Errorf("Concurrent ops: val = %d, want 1000", val)
	}
}
