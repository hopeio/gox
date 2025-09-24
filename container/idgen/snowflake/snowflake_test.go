/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package snowflake

import (
	"sync"
	"testing"
)

func TestSnowFlake(t *testing.T) {
	node := NewNode(1, 1, 10)
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
