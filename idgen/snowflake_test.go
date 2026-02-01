/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package idgen

import (
	"sync"
	"testing"
)

func TestSnowFlake(t *testing.T) {
	node := NewSnowflake(1, 1)
	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			id := node.Generate()
			t.Log(id)
			wg.Done()
		}()
	}
	wg.Wait()
}
