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

	poller.RandRun(time.Second, time.Second*2, func() {
		t.Log("hello")
	})
}

func TestPoller(t *testing.T) {
	poller := NewPoller()
	poller.Run(time.Second, func() {
		t.Log("hello")
	})
}
