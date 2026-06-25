/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package scheduler

import (
	"time"

	timex "github.com/hopeio/gox/time"
)


type Poller struct {
	times uint
	limit time.Duration
	ch    chan struct{}
}

func NewPoller() *Poller {
	return &Poller{}
}

func (p *Poller) Times() uint {
	return p.times
}

func (p *Poller) LimitDuration(d time.Duration) {
	p.limit = d
}

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

func (p *Poller) Stop() {
	close(p.ch)
}
