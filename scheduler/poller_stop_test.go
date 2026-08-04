package scheduler

import (
	"testing"
	"time"
)

func TestPoller_StopBeforeTick(t *testing.T) {
	p := NewPoller()
	done := make(chan struct{})
	go func() {
		p.Run(24*time.Hour, func() {})
		close(done)
	}()
	// first do() runs sync; Stop should exit select promptly
	time.Sleep(10 * time.Millisecond)
	p.Stop()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after Stop")
	}
	if p.Times() < 1 {
		t.Fatal("Times should be >= 1 after first do()")
	}
}
