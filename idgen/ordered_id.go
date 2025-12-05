/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package idgen

import (
	"sync/atomic"
	"time"
)

var currentID uint64 = uint64(time.Now().Unix()) << 32

// 单机顺序id
func NewOrderedID() uint64 {
	return atomic.AddUint64(&currentID, 1)
}
