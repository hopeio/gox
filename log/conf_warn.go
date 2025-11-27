/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package log

import (
	"cmp"

	constraintsx "github.com/hopeio/gox/types/constraints"
)

func ValueRangeNotify[T cmp.Ordered](msg string, v, rangeMin, rangeMax T) {
	if v > rangeMax || v < rangeMin {
		CallerSkipLogger(1).Warnf("%s except: %v - %v, but got %v", msg, rangeMin, rangeMax, v)
	}
}

func ValueLevelNotify[T constraintsx.Number](msg string, v, std T) {
	if v > 0 && v < std {
		CallerSkipLogger(1).Warnf("%s except: %v level, but got %v", msg, std, v)
	}
}
