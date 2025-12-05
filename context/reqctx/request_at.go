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

type RequestAt struct {
	Time       time.Time
	TimeStamp  int64
	timeString string
}

func (r *RequestAt) String() string {
	if r.timeString == "" {
		r.timeString = r.Time.Format(timex.LayoutTimeMacro)
	}
	return r.timeString
}

func NewRequestAt() RequestAt {
	now := time.Now()
	return RequestAt{
		Time:      now,
		TimeStamp: now.Unix(),
	}
}

func NewRequestAtFromTime(t time.Time) RequestAt {
	return RequestAt{
		Time:      t,
		TimeStamp: t.Unix(),
	}
}
