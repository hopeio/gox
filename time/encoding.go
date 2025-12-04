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

type EncodeType int8

const (
	EncodeTypeLayout EncodeType = iota
	EncodeTypeUnixSeconds
	EncodeTypeUnixMilliseconds
	EncodeTypeUnixMicroseconds
	EncodeTypeUnixNanoseconds
)

var DefaultEncoding = &Encoding{
	Layout: time.RFC3339Nano,
}

func SetDefaultEncoding(e *Encoding) {
	DefaultEncoding = e
}

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
	EncodeType
	Layout string
}

func (u *Encoding) SetType(typ EncodeType) {
	u.EncodeType = typ
}

func (u *Encoding) SetLayout(l string) {
	u.Layout = l
}

func (u *Encoding) marshalText(t time.Time) ([]byte, error) {
	switch u.EncodeType {
	case EncodeTypeLayout:
		if u.Layout == "" || u.Layout == time.RFC3339Nano {
			return t.MarshalText()
		}
		return []byte(t.Format(u.Layout)), nil
	case EncodeTypeUnixSeconds:
		return strconv.AppendInt(nil, t.Unix(), 10), nil
	case EncodeTypeUnixMilliseconds:
		return strconv.AppendInt(nil, t.UnixMilli(), 10), nil
	case EncodeTypeUnixMicroseconds:
		return strconv.AppendInt(nil, t.UnixMicro(), 10), nil
	case EncodeTypeUnixNanoseconds:
		return strconv.AppendInt(nil, t.UnixNano(), 10), nil
	}
	return t.MarshalText()
}

func (u *Encoding) unmarshalText(t *time.Time, data []byte) error {
	tstr := string(data)
	if tstr == "" {
		return nil
	}

	if u.EncodeType == EncodeTypeLayout {
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
		switch u.EncodeType {
		case EncodeTypeUnixSeconds:
			*t = time.Unix(parseInt, 0)
			return nil
		case EncodeTypeUnixMilliseconds:
			*t = time.UnixMilli(parseInt)
			return nil
		case EncodeTypeUnixMicroseconds:
			*t = time.UnixMicro(parseInt)
			return nil
		case EncodeTypeUnixNanoseconds:
			*t = time.Unix(0, parseInt)
			return nil
		}
	}

	return t.UnmarshalText(data)
}

func (u *Encoding) marshalJSON(t time.Time) ([]byte, error) {
	if u.EncodeType == EncodeTypeLayout {
		if u.Layout == "" || u.Layout == time.RFC3339Nano {
			return t.MarshalJSON()
		}
		return []byte(`"` + t.Format(u.Layout) + `"`), nil
	} else {
		switch u.EncodeType {
		case EncodeTypeUnixSeconds:
			return strconv.AppendInt(nil, t.Unix(), 10), nil
		case EncodeTypeUnixMilliseconds:
			return strconv.AppendInt(nil, t.UnixMilli(), 10), nil
		case EncodeTypeUnixMicroseconds:
			return strconv.AppendInt(nil, t.UnixMicro(), 10), nil
		case EncodeTypeUnixNanoseconds:
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
	if u.EncodeType == EncodeTypeLayout {
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
		switch u.EncodeType {
		case EncodeTypeUnixSeconds:
			*t = time.Unix(parseInt, 0)
			return nil
		case EncodeTypeUnixMilliseconds:
			*t = time.UnixMilli(parseInt)
			return nil
		case EncodeTypeUnixMicroseconds:
			*t = time.UnixMicro(parseInt)
			return nil
		case EncodeTypeUnixNanoseconds:
			*t = time.Unix(0, parseInt)
			return nil
		}
	}

	return t.UnmarshalJSON(data)
}
