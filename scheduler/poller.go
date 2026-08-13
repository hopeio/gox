/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scheduler

import (
	"sync"
	"time"

	timex "github.com/hopeio/gox/time"
)

type Poller struct {
	times    uint
	limit    time.Duration
	ch       chan struct{}
	stopOnce sync.Once
}

// NewPoller creates and returns a new instance.
func NewPoller() *Poller {
	return &Poller{ch: make(chan struct{})}
}

// Times returns the result.
func (p *Poller) Times() uint {
	return p.times
}

// LimitDuration performs the operation.
func (p *Poller) LimitDuration(d time.Duration) {
	p.limit = d
}

// Run executes the operation.
func (p *Poller) Run(interval time.Duration, do func()) {
	timer := time.NewTicker(interval)
	p.times++
	do()
	for {
		select {
		case <-p.ch:
			timer.Stop()
			return
		case <-timer.C:
			p.times++
			do()
		}
	}
}

// RandRun performs the operation.
func (p *Poller) RandRun(minInterval, maxInterval time.Duration, do func()) {
	timer := timex.NewRandTicker(minInterval, maxInterval)
	p.times++
	do()
	for {
		select {
		case <-p.ch:
			timer.Stop()
			return
		default:
			timer.Wait()
			p.times++
			do()
		}
	}
}

// Stop closes and releases resources. Safe to call multiple times.
func (p *Poller) Stop() {
	p.stopOnce.Do(func() {
		close(p.ch)
	})
}
