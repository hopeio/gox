/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package time

import (
	"math/rand"
	"time"
)

type Ticker interface {
	Reset(time.Duration) bool
	Stop() bool
	Wait()
}

type FixedTicker time.Ticker

// Stop closes and releases resources.
func (t *FixedTicker) Stop() bool {
	(*time.Ticker)(t).Stop()
	return true
}

// Reset ...
func (t *FixedTicker) Reset(d time.Duration) bool {
	(*time.Ticker)(t).Reset(d)
	return true
}

// Wait ...
func (t *FixedTicker) Wait() {
	<-t.C
}

// Channel ...
func (t *FixedTicker) Channel() <-chan time.Time {
	return t.C
}

// NewTicker creates and returns a new instance.
func NewTicker(interval time.Duration) Ticker {
	return (*FixedTicker)(time.NewTicker(interval))
}

var _ Ticker = &RandTicker{}

type RandTicker struct {
	timer                 *time.Timer
	limitBase, limitRange time.Duration
}

// Reset ...
func (t *RandTicker) Reset(d time.Duration) bool {
	t.limitBase = d
	return t.reset()
}

// reset ...
func (t *RandTicker) reset() bool {
	if t.limitRange == 0 {
		return t.timer.Reset(t.limitBase)
	}
	return t.timer.Reset(t.limitBase + time.Duration(rand.Intn(int(t.limitRange))))
}

// Wait ...
func (t *RandTicker) Wait() {
	<-t.timer.C
	t.reset()
}

// Stop closes and releases resources.
func (t *RandTicker) Stop() bool {
	return t.timer.Stop()
}

// NewRandTicker creates and returns a new instance.
func NewRandTicker(minInterval, maxInterval time.Duration) Ticker {
	limitRange := maxInterval - minInterval
	if limitRange == 0 {
		return NewTicker(maxInterval)
	}
	if limitRange < 0 {
		minInterval, maxInterval = maxInterval, minInterval
		limitRange = -limitRange
	}
	return &RandTicker{
		timer:      time.NewTimer(minInterval + time.Duration(rand.Intn(int(limitRange)))),
		limitBase:  minInterval,
		limitRange: limitRange,
	}
}
