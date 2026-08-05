/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package idgen

import (
	"sync/atomic"
)

var defaultOrderedIDGenerator = NewOrderedIDGenerator(0)

// NewOrderedIDGenerator creates and returns a new instance.
func NewOrderedIDGenerator(initial uint64) func() uint64 {
	return func() uint64 {
		return atomic.AddUint64(&initial, 1)
	}
}

// NewOrderedID creates and returns a new instance.
func NewOrderedID() uint64 {
	return defaultOrderedIDGenerator()
}
