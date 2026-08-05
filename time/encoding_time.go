/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package time

import (
	"time"
)

type EncodingTime struct {
	time.Time
	*Encoding
}

// MarshalJSON encodes the value.
func (u EncodingTime) MarshalJSON() ([]byte, error) {
	return u.marshalJSON(u.Time)
}

// UnmarshalJSON decodes into the value.
func (u *EncodingTime) UnmarshalJSON(data []byte) error {
	return u.unmarshalJSON(&u.Time, data)
}

type GlobEncodingTime time.Time

// MarshalJSON encodes the value.
func (u GlobEncodingTime) MarshalJSON() ([]byte, error) {
	return DefaultEncoding.marshalJSON(time.Time(u))
}

// UnmarshalJSON decodes into the value.
func (u *GlobEncodingTime) UnmarshalJSON(data []byte) error {
	return DefaultEncoding.unmarshalJSON((*time.Time)(u), data)
}
