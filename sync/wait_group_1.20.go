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

// 运行计数，等待计数，信号计数
type WaitGroup struct {
	noCopy noCopy

	state atomic.Uint64 // high 32 bits are counter, low 32 bits are waiter count.
	sema  uint32
}

// WaitGroupState返回 sync.WaitGroup 的状态,
func WaitGroupState(wg *sync.WaitGroup) (counter int32, wcounter uint32) {
	wgc := (*WaitGroup)(unsafe.Pointer(wg))
	return wgc.State()
}

func (wg *WaitGroup) State() (counter int32, wcounter uint32) {
	state := wg.state.Load()
	return int32(state >> 32), uint32(state)
}

func WaitGroupStopWait(wg *sync.WaitGroup) {
	state, _ := WaitGroupState(wg)
	wg.Add(int(-state))
}
