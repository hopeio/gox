/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package reqctx

import (
	"time"

	timex "github.com/hopeio/gox/time"
)

type RequestTime struct {
	Time       time.Time
	TimeStamp  int64
	TimeString string
}

func (r *RequestTime) String() string {
	if r.TimeString == "" {
		r.TimeString = r.Time.Format(timex.LayoutTimeMacro)
	}
	return r.TimeString
}

func NewRequestAt() RequestTime {
	now := time.Now()
	return RequestTime{
		Time:      now,
		TimeStamp: now.Unix(),
	}
}

func NewRequestAtFromTime(t time.Time) RequestTime {
	return RequestTime{
		Time:      t,
		TimeStamp: t.Unix(),
	}
}
