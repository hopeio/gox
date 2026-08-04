/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scheduler

import (
	"testing"
	"time"
)

func TestPollerRand(t *testing.T) {
	poller := NewPoller()
	done := make(chan struct{})
	go func() {
		poller.RandRun(50*time.Millisecond, 100*time.Millisecond, func() {
			t.Log("hello")
		})
		close(done)
	}()
	time.Sleep(120 * time.Millisecond)
	poller.Stop()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("RandRun did not return after Stop")
	}
}

func TestPoller(t *testing.T) {
	poller := NewPoller()
	done := make(chan struct{})
	go func() {
		poller.Run(50*time.Millisecond, func() {
			t.Log("hello")
		})
		close(done)
	}()
	time.Sleep(120 * time.Millisecond)
	poller.Stop()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after Stop")
	}
}
