/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package time

import (
	"strconv"
	"time"
)

type encodeType int8

const (
	encodeTypeLayout encodeType = iota
	encodeTypeUnixSeconds
	encodeTypeUnixMilliseconds
	encodeTypeUnixMicroseconds
	encodeTypeUnixNanoseconds
)

var (
	DefaultEncoding = &Encoding{
		Layout: time.RFC3339Nano,
	}
	EncodingUnixSeconds = &Encoding{
		encodeType: encodeTypeUnixSeconds,
	}
	EncodingUnixMilliseconds = &Encoding{
		encodeType: encodeTypeUnixMilliseconds,
	}
	EncodingUnixMicroseconds = &Encoding{
		encodeType: encodeTypeUnixMicroseconds,
	}
	EncodingUnixNanoseconds = &Encoding{
		encodeType: encodeTypeUnixNanoseconds,
	}
)

func MarshalJSON(t time.Time) ([]byte, error) {
	return DefaultEncoding.marshalJSON(t)
}

func UnmarshalJSON(t *time.Time, data []byte) error {
	return DefaultEncoding.unmarshalJSON(t, data)
}

func MarshalText(t time.Time) ([]byte, error) {
	return DefaultEncoding.marshalText(t)
}

func UnmarshalText(t *time.Time, data []byte) error {
	return DefaultEncoding.unmarshalText(t, data)
}

type Encoding struct {
	encodeType
	Layout string
}

func NewLayOutEncoding(layout string) *Encoding {
	return &Encoding{
		Layout: layout,
	}
}

func (u *Encoding) marshalText(t time.Time) ([]byte, error) {
	switch u.encodeType {
	case encodeTypeLayout:
		if u.Layout == "" || u.Layout == time.RFC3339Nano {
			return t.MarshalText()
		}
		return []byte(t.Format(u.Layout)), nil
	case encodeTypeUnixSeconds:
		return strconv.AppendInt(nil, t.Unix(), 10), nil
	case encodeTypeUnixMilliseconds:
		return strconv.AppendInt(nil, t.UnixMilli(), 10), nil
	case encodeTypeUnixMicroseconds:
		return strconv.AppendInt(nil, t.UnixMicro(), 10), nil
	case encodeTypeUnixNanoseconds:
		return strconv.AppendInt(nil, t.UnixNano(), 10), nil
	}
	return t.MarshalText()
}

func (u *Encoding) unmarshalText(t *time.Time, data []byte) error {
	tstr := string(data)
	if tstr == "" {
		return nil
	}

	if u.encodeType == encodeTypeLayout {
		if u.Layout == "" || u.Layout == time.RFC3339Nano {
			return t.UnmarshalText(data)
		} else {
			var err error
			*t, err = time.Parse(u.Layout, string(data))
			return err
		}
	} else {
		parseInt, err := strconv.ParseInt(tstr, 10, 64)
		if err != nil {
			return err
		}
		switch u.encodeType {
		case encodeTypeUnixSeconds:
			*t = time.Unix(parseInt, 0)
			return nil
		case encodeTypeUnixMilliseconds:
			*t = time.UnixMilli(parseInt)
			return nil
		case encodeTypeUnixMicroseconds:
			*t = time.UnixMicro(parseInt)
			return nil
		case encodeTypeUnixNanoseconds:
			*t = time.Unix(0, parseInt)
			return nil
		}
	}

	return t.UnmarshalText(data)
}

func (u *Encoding) marshalJSON(t time.Time) ([]byte, error) {
	if u.encodeType == encodeTypeLayout {
		if u.Layout == "" || u.Layout == time.RFC3339Nano {
			return t.MarshalJSON()
		}
		return []byte(`"` + t.Format(u.Layout) + `"`), nil
	} else {
		switch u.encodeType {
		case encodeTypeUnixSeconds:
			return strconv.AppendInt(nil, t.Unix(), 10), nil
		case encodeTypeUnixMilliseconds:
			return strconv.AppendInt(nil, t.UnixMilli(), 10), nil
		case encodeTypeUnixMicroseconds:
			return strconv.AppendInt(nil, t.UnixMicro(), 10), nil
		case encodeTypeUnixNanoseconds:
			return strconv.AppendInt(nil, t.UnixNano(), 10), nil
		}
	}

	return t.MarshalJSON()
}

func (u *Encoding) unmarshalJSON(t *time.Time, data []byte) error {
	tstr := string(data)
	if tstr == "null" {
		return nil
	}
	if u.encodeType == encodeTypeLayout {
		if u.Layout == "" || u.Layout == time.RFC3339Nano {
			return t.UnmarshalJSON(data)
		} else {
			var err error
			*t, err = time.Parse(`"`+u.Layout+`"`, string(data))
			return err
		}
	} else {
		parseInt, err := strconv.ParseInt(tstr, 10, 64)
		if err != nil {
			return err
		}
		switch u.encodeType {
		case encodeTypeUnixSeconds:
			*t = time.Unix(parseInt, 0)
			return nil
		case encodeTypeUnixMilliseconds:
			*t = time.UnixMilli(parseInt)
			return nil
		case encodeTypeUnixMicroseconds:
			*t = time.UnixMicro(parseInt)
			return nil
		case encodeTypeUnixNanoseconds:
			*t = time.Unix(0, parseInt)
			return nil
		}
	}

	return t.UnmarshalJSON(data)
}
