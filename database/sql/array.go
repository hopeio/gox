/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package sql

import (
	"bytes"
	"database/sql/driver"
	"encoding"
	"errors"
	"fmt"
	"time"

	jsonx "github.com/hopeio/gox/encoding/json"
	stringsx "github.com/hopeio/gox/strings"

	"strconv"
	"strings"

	"golang.org/x/exp/constraints"
)

// adpter postgres
type IntArray[T constraints.Integer] []T

// Scan performs the operation.
func (d *IntArray[T]) Scan(value any) error {
	str, ok := value.(string)
	if !ok {
		data, ok := value.([]byte)
		if !ok {
			return errors.New(fmt.Sprint("failed to scan int array value:", value))
		}
		str = stringsx.FromBytes(data)
	}
	strs := strings.Split(str[1:len(str)-1], ",")
	var arr []T
	for _, numstr := range strs {
		num, err := strconv.Atoi(numstr)
		if err != nil {
			return err
		}
		arr = append(arr, T(num))
	}
	*d = arr
	return nil
}

// Value returns the value.
func (d IntArray[T]) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, num := range d {
		buf.WriteString(strconv.Itoa(int(num)))
		if i != len(d)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte('}')
	return buf.String(), nil
}

type FloatArray[T constraints.Float] []T

// Scan performs the operation.
func (d *FloatArray[T]) Scan(value any) error {
	str, ok := value.(string)
	if !ok {
		data, ok := value.([]byte)
		if !ok {
			return errors.New(fmt.Sprint("failed to scan float array value:", value))
		}
		str = string(data)
	}
	strs := strings.Split(str[1:len(str)-1], ",")
	var arr []T
	for _, numstr := range strs {
		num, err := strconv.ParseFloat(numstr, 64)
		if err != nil {
			return err
		}
		arr = append(arr, T(num))
	}
	*d = arr
	return nil
}

// Value returns the value.
func (d FloatArray[T]) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, num := range d {
		buf.WriteString(strconv.FormatFloat(float64(num), 'g', -1, 64))
		if i != len(d)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte('}')
	return buf.String(), nil
}

type StringArray []string

// Scan accepts a PG array literal, []byte, or a native []string from the driver.
func (d *StringArray) Scan(value any) error {
	if value == nil {
		*d = nil
		return nil
	}
	switch v := value.(type) {
	case []string:
		out := make([]string, len(v))
		copy(out, v)
		*d = out
		return nil
	case string:
		return d.parseLiteral(v)
	case []byte:
		return d.parseLiteral(stringsx.FromBytes(v))
	default:
		return errors.New(fmt.Sprint("failed to scan string array value:", value))
	}
}

func (d *StringArray) parseLiteral(str string) error {
	if str == "" || str == "{}" {
		*d = StringArray{}
		return nil
	}
	if len(str) < 2 || str[0] != '{' || str[len(str)-1] != '}' {
		return fmt.Errorf("invalid string array literal: %q", str)
	}
	inner := str[1 : len(str)-1]
	if inner == "" {
		*d = StringArray{}
		return nil
	}
	var (
		arr     []string
		cur     strings.Builder
		inQuote bool
		escape  bool
	)
	flush := func() {
		arr = append(arr, cur.String())
		cur.Reset()
	}
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		if escape {
			cur.WriteByte(c)
			escape = false
			continue
		}
		if c == '\\' && inQuote {
			escape = true
			continue
		}
		if c == '"' {
			inQuote = !inQuote
			continue
		}
		if c == ',' && !inQuote {
			flush()
			continue
		}
		cur.WriteByte(c)
	}
	flush()
	*d = arr
	return nil
}

// Value encodes as a PostgreSQL array literal: {}, {"a"}, {"a","b"}.
func (d StringArray) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	if len(d) == 0 {
		return "{}", nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, str := range d {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteByte('"')
		for j := 0; j < len(str); j++ {
			c := str[j]
			if c == '\\' || c == '"' {
				buf.WriteByte('\\')
			}
			buf.WriteByte(c)
		}
		buf.WriteByte('"')
	}
	buf.WriteByte('}')
	return buf.String(), nil
}

// Array represents a PostgreSQL array for T. It implements the ArrayGetter and ArraySetter interfaces. It preserves
// PostgreSQL dimensions and custom lower bounds. Use FlatArray if these are not needed.
// only support number
type Array[T any] []T

// Scan performs the operation.
func (d *Array[T]) Scan(value any) error {
	str, ok := value.(string)
	if !ok {
		data, ok := value.([]byte)
		if !ok {
			return errors.New(fmt.Sprint("failed to scan array value:", value))
		}
		str = string(data)
	}
	var arr []T
	str = str[1 : len(str)-1]
	if len(str) > 0 && str[0] == '{' {
		i := 0
		for i < len(str) {
			subArray, ok := stringsx.BracketsIntervals(str[i:], '{', '}')
			if ok {
				i += len(subArray)
				t, err := StringConvertFor[T](subArray)
				if err != nil {
					return err
				}
				arr = append(arr, t)
			} else {
				break
			}
		}
		*d = arr
		return nil
	}
	strs := strings.Split(str, ",")

	for _, elem := range strs {
		t, err := StringConvertFor[T](elem)
		if err != nil {
			return err
		}
		arr = append(arr, t)
	}
	*d = arr
	return nil
}

// Value returns the value.
func (d Array[T]) Value() (driver.Value, error) {
	if len(d) == 0 {
		return nil, nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, v := range d {
		if i > 0 {
			buf.WriteByte(',')
		}
		a, ap := any(v), any(&v)
		ivv, ok := a.(driver.Valuer)
		if !ok {
			ivv, ok = ap.(driver.Valuer)
		}
		if ok {
			v, err := ivv.Value()
			if err != nil {
				return nil, err
			}
			buf.WriteString(stringsx.Format(v))
			continue
		}
		itv, ok := a.(encoding.TextMarshaler)
		if !ok {
			itv, ok = ap.(encoding.TextMarshaler)
		}
		if ok {
			v, err := itv.MarshalText()
			if err != nil {
				return nil, err
			}
			buf.WriteString(strconv.Quote(stringsx.FromBytes(v)))
			continue
		}
		buf.WriteString(stringsx.Format(v))
	}
	buf.WriteByte('}')
	return buf.String(), nil
}

type TimeArray []time.Time

// Scan performs the operation.
func (d *TimeArray) Scan(value any) error {
	str, ok := value.(string)
	if !ok {
		data, ok := value.([]byte)
		if !ok {
			return errors.New(fmt.Sprint("failed to scan string array value:", value))
		}
		str = stringsx.FromBytes(data)
	}
	strs := strings.Split(str[1:len(str)-1], ",")
	var arr []time.Time
	for _, elem := range strs {
		t, err := time.Parse(time.RFC3339Nano, elem[1:len(elem)-1])
		if err != nil {
			return err
		}
		arr = append(arr, t)
	}
	*d = arr
	return nil
}

// Value returns the value.
func (d TimeArray) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	buf.WriteByte('"')
	for i, t := range d {
		buf.WriteString(t.Format(time.RFC3339Nano))
		if i != len(d)-1 {
			buf.WriteByte('"')
			buf.WriteByte(',')
		}
	}
	buf.WriteByte('"')
	buf.WriteByte('}')
	return buf.String(), nil
}

// Do not use this type; the forms below are valid — use jsonb directly
// {[],[]} {"{}","{}"} {"{}",[]}
type jsonArray []map[string]any

// Scan performs the operation.
func (j *jsonArray) Scan(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		data, ok := value.([]byte)
		if !ok {
			return errors.New(fmt.Sprint("failed to scan array value:", value))
		}
		str = string(data)
	}
	var arr []map[string]any
	str = str[1 : len(str)-1]

	for {
		jsonStr := str
		idx := strings.Index(str, `","`)
		if idx != -1 {
			jsonStr = str[:idx+1]
			str = str[idx+2:]
		}
		var err error
		jsonStr, err = strconv.Unquote(jsonStr)
		if err != nil {
			return err
		}
		var m map[string]any
		err = jsonx.Unmarshal(stringsx.ToBytes(jsonStr), &m)
		if err != nil {
			return err
		}
		arr = append(arr, m)
		if idx == -1 {
			break
		}
	}
	*j = arr
	return nil
}

// Value returns the value.
func (j jsonArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, v := range j {
		if i > 0 {
			buf.WriteByte(',')
		}
		data, err := jsonx.Marshal(&v)
		if err != nil {
			return nil, err
		}
		_, err = buf.WriteString(strconv.Quote(stringsx.FromBytes(data)))
		if err != nil {
			return nil, err
		}
	}

	buf.WriteByte('}')
	return buf.String(), nil
}

// GormDataType returns the result.
func (*jsonArray) GormDataType() string {
	return "jsonb[]"
}
