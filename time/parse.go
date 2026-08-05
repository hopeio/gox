/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package time

import (
	"time"
)

// ParseTime parses a time-only string.
func ParseTime(t string) (time.Time, error) {
	return time.Parse(time.TimeOnly, t)
}

// ParseDateTime parses a full date-time string.
func ParseDateTime(t string) (time.Time, error) {
	return time.Parse(time.DateTime, t)
}

// ParseDate parses a date-only string.
func ParseDate(t string) (time.Time, error) {
	return time.Parse(time.DateOnly, t)
}
