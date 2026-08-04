/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 */

package time

import (
	"testing"
	stdtime "time"
)

func TestFixedTickerStop(t *testing.T) {
	ticker := NewTicker(stdtime.Millisecond)
	if !ticker.Stop() {
		t.Fatal("Stop should return true for active ticker")
	}
}

func TestFixedTickerWaitOnce(t *testing.T) {
	ticker := NewTicker(stdtime.Millisecond)
	defer ticker.Stop()
	done := make(chan struct{})
	go func() {
		ticker.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-stdtime.After(100 * stdtime.Millisecond):
		t.Fatal("FixedTicker.Wait timed out")
	}
}

func TestFixedTickerReset(t *testing.T) {
	ft := (*FixedTicker)(stdtime.NewTicker(stdtime.Millisecond))
	defer ft.Stop()
	if !ft.Reset(2 * stdtime.Millisecond) {
		t.Fatal("Reset should return true")
	}
}

func TestRandTickerEqualBoundsUsesFixed(t *testing.T) {
	ticker := NewRandTicker(stdtime.Millisecond, stdtime.Millisecond)
	if _, ok := ticker.(*FixedTicker); !ok {
		t.Fatalf("equal bounds should use FixedTicker, got %T", ticker)
	}
	ticker.Stop()
}

func TestRandTickerSwappedBounds(t *testing.T) {
	ticker := NewRandTicker(5*stdtime.Millisecond, stdtime.Millisecond)
	defer ticker.Stop()
	if _, ok := ticker.(*RandTicker); !ok {
		t.Fatalf("expected RandTicker, got %T", ticker)
	}
}

func TestRandTickerWaitOnce(t *testing.T) {
	ticker := NewRandTicker(stdtime.Millisecond, 2*stdtime.Millisecond)
	defer ticker.Stop()
	done := make(chan struct{})
	go func() {
		ticker.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-stdtime.After(100 * stdtime.Millisecond):
		t.Fatal("RandTicker.Wait timed out")
	}
}

func TestRandTickerReset(t *testing.T) {
	rt := &RandTicker{
		timer:      stdtime.NewTimer(stdtime.Millisecond),
		limitBase:  stdtime.Millisecond,
		limitRange: stdtime.Millisecond,
	}
	defer rt.Stop()
	if !rt.Reset(2 * stdtime.Millisecond) {
		t.Fatal("Reset should return true")
	}
}
