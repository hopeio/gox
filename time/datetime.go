/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package time

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strconv"
	"time"

	"github.com/hopeio/gox/encoding/binary"
	stringsx "github.com/hopeio/gox/strings"
)

// 1-01-01~9999-12-31
type Date int32

func DateFromTime(t time.Time) Date {
	return Date(t.Unix() / SecondsOfDay)
}

func (d Date) Time() time.Time {
	return time.Unix(int64(d)*SecondsOfDay, 0)
}

// Scan scan time.
func (d *Date) Scan(value interface{}) (err error) {
	nullTime := &sql.NullTime{}
	err = nullTime.Scan(value)
	*d = Date(nullTime.Time.Unix() / SecondsOfDay)
	return
}

// Value get time value.
func (d Date) Value() (driver.Value, error) {
	return []byte(time.Unix(int64(d)*SecondsOfDay, 0).Format(time.DateOnly)), nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, 12)
	b = append(b, '"')
	b = append(b, stringsx.ToBytes(d.Time().Format(time.DateOnly))...)
	b = append(b, '"')
	return b, nil
}

func (d *Date) UnmarshalJSON(data []byte) error {
	str := stringsx.BytesToString(data)
	if len(data) == 0 || str == "null" {
		return nil
	}

	if len(str) > 1 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
		t, err := time.Parse(time.DateOnly, str)
		if err != nil {
			return err
		}
		*d = Date(t.Unix() / SecondsOfDay)
		return nil
	} else {
		v, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		if len(str) == 13 {
			*d = Date(v / 1000 / SecondsOfDay)
		} else {
			*d = Date(v / SecondsOfDay)
		}
	}
	return nil
}

func (d Date) MarshalText() ([]byte, error) {
	return stringsx.ToBytes(d.Time().Format(time.DateOnly)), nil
}

func (d *Date) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	str := stringsx.BytesToString(data)
	t, err := time.Parse(time.DateOnly, str)
	if err != nil {
		return err
	}
	*d = Date(t.Unix() / SecondsOfDay)
	return nil
}

func (d Date) GormDataType() string {
	return "time"
}

func (d Date) MarshalBinary() ([]byte, error) {
	return binary.ToBinary(d), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (d *Date) UnmarshalBinary(data []byte) error {
	*d = binary.BinaryTo[Date](data)
	return nil
}

func (d Date) GobEncode() ([]byte, error) {
	return d.MarshalBinary()
}

func (d *Date) GobDecode(data []byte) error {
	return d.UnmarshalBinary(data)
}

func (d Date) MarshalGQL(w io.Writer) {
	w.Write([]byte(d.Time().Format(time.DateOnly)))
}

func (d *Date) UnmarshalGQL(v interface{}) error {
	if i, ok := v.(string); ok {
		t, err := time.Parse(time.DateOnly, i)
		if err != nil {
			return err
		}
		*d = Date(t.Unix() / SecondsOfDay)
		return nil
	}
	return errors.New("enum need integer type")
}

func (d Date) String() string {
	return d.Time().Format(time.DateOnly)
}

// 01-01-01 00:00:00~9999-12-31 23:59:59
type DateTime int64

func DateTimeFromTime(t time.Time) Date {
	return Date(t.Unix())
}

func (d DateTime) Time() time.Time {
	return time.Unix(int64(d), 0)
}

// Scan scan time.
func (d *DateTime) Scan(value interface{}) (err error) {
	nullTime := &sql.NullTime{}
	err = nullTime.Scan(value)
	*d = DateTime(nullTime.Time.Unix())
	return
}

// Value get time value.
func (d DateTime) Value() (driver.Value, error) {
	return time.Unix(int64(d), 0), nil
}

func (d DateTime) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(time.DateTime)+2)
	b = append(b, '"')
	b = append(b, stringsx.ToBytes(time.Unix(int64(d), 0).Format(time.DateTime))...)
	b = append(b, '"')
	return b, nil
}

func (d *DateTime) UnmarshalJSON(data []byte) error {
	str := stringsx.BytesToString(data)
	if len(data) == 0 || str == "null" {
		return nil
	}

	if len(str) > 1 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
		t, err := time.Parse(time.DateTime, str)
		if err != nil {
			return err
		}
		*d = DateTime(t.Unix())
		return nil
	} else {
		v, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		if len(str) == 13 {
			*d = DateTime(v / 1000)
		} else {
			*d = DateTime(v)
		}
	}
	return nil
}

func (d DateTime) MarshalText() ([]byte, error) {
	return stringsx.ToBytes(d.Time().Format(time.DateTime)), nil
}

func (d *DateTime) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	str := stringsx.BytesToString(data)
	t, err := time.Parse(time.DateTime, str)
	if err != nil {
		return err
	}
	*d = DateTime(t.Unix())
	return nil
}

func (ts DateTime) GormDataType() string {
	return "time"
}

func (ts DateTime) Date() Date {
	return Date(int64(ts) / SecondsOfDay)
}
