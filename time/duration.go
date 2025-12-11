/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package time

import (
	"context"
	"time"

	"github.com/hopeio/gox/strings"
)

// Duration be used toml unmarshal string time, like 1s, 500ms.
type Duration int64

// UnmarshalText unmarshal text to duration.
func (d *Duration) UnmarshalText(text []byte) error {
	tmp, err := time.ParseDuration(string(text))
	if err == nil {
		*d = Duration(tmp)
	}
	return err
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// Shrink will decrease the duration by comparing with context's timeout duration
// and return new timeout\context\CancelFunc.
func (d Duration) Shrink(c context.Context) (Duration, context.Context, context.CancelFunc) {
	if deadline, ok := c.Deadline(); ok {
		if ctimeout := time.Until(deadline); ctimeout < time.Duration(d) {
			// deliver small timeout
			return Duration(ctimeout), c, func() {}
		}
	}
	ctx, cancel := context.WithTimeout(c, time.Duration(d))
	return d, ctx, cancel
}

// 标准化Duration
func NormalizeDuration(td time.Duration, stdTd time.Duration) time.Duration {
	if td == 0 {
		return td
	}
	if td < stdTd {
		return td * stdTd
	}
	return td
}

func (t Duration) MarshalJSON() ([]byte, error) {
	return strings.ToBytes(time.Duration(t).String()), nil
}

// UnmarshalJSON implements the [encoding/json.Unmarshaler] interface.
// The time must be a quoted string in the RFC 3339 format.
func (t *Duration) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	duration, err := time.ParseDuration(strings.BytesToString(data))
	if err != nil {
		return err
	}
	*t = Duration(duration)
	return nil
}
