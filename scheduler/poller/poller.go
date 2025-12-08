/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package poller

import (
	"context"
	"time"

	timex "github.com/hopeio/gox/time"
)

type TaskFunc = func(context.Context)

type Poller struct {
	times uint
}

func NewPoller() *Poller {
	return &Poller{}
}

func (task *Poller) Times() uint {
	return task.times
}

func (task *Poller) Run(ctx context.Context, interval time.Duration, do TaskFunc) {
	timer := time.NewTicker(interval)
	task.times++
	do(ctx)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			task.times++
			do(ctx)
		}
	}
}

func (task *Poller) RandRun(ctx context.Context, minInterval, maxInterval time.Duration, do TaskFunc) {
	timer := timex.NewRandTicker(minInterval, maxInterval)
	task.times++
	do(ctx)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.Channel():
			task.times++
			do(ctx)
		}
	}
}

func Run(ctx context.Context, interval time.Duration, do TaskFunc) {
	timer := time.NewTicker(interval)
	times := 1
	do(ctx)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			times++
			do(ctx)
		}
	}
}

func RandRun(ctx context.Context, minInterval, maxInterval time.Duration, do TaskFunc) {
	timer := timex.NewRandTicker(minInterval, maxInterval)
	times := 1
	do(ctx)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.Channel():
			times++
			do(ctx)
		}
	}
}
