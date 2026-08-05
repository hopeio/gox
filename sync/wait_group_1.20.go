//go:build go1.20

/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package sync

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// running count, waiter count, and signal count.
type WaitGroup struct {
	noCopy noCopy

	state atomic.Uint64 // high 32 bits are counter, low 32 bits are waiter count.
	sema  uint32
}

// WaitGroupState reads internal counter/state values from the standard sync.WaitGroup.
func WaitGroupState(wg *sync.WaitGroup) (counter int32, wcounter uint32) {
	wgc := (*WaitGroup)(unsafe.Pointer(wg))
	return wgc.State()
}

// State returns the wait-group's internal counter and waiter counter.
func (wg *WaitGroup) State() (counter int32, wcounter uint32) {
	state := wg.state.Load()
	return int32(state >> 32), uint32(state)
}

// WaitGroupStopWait reduces the wait-group counter by its current running count.
func WaitGroupStopWait(wg *sync.WaitGroup) {
	state, _ := WaitGroupState(wg)
	wg.Add(int(-state))
}
