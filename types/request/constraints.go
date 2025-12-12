/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package request

import (
	"time"

	"golang.org/x/exp/constraints"
)

type Ordered interface {
	constraints.Ordered | time.Time | ~*time.Time
}
