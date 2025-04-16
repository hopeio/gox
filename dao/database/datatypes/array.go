/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package datatypes

import (
	"bytes"
	"database/sql/driver"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	dbi "github.com/hopeio/utils/dao/database"
	reflecti "github.com/hopeio/utils/reflect/converter"
	stringsi "github.com/hopeio/utils/strings"
	"time"

	"golang.org/x/exp/constraints"
	"strconv"
	"strings"
)

// adpter postgres
type IntArray[T constraints.Integer] []T

func (d *IntArray[T]) Scan(value any) error {
	str, ok := value.(string)
	if !ok {
		data, ok := value.([]byte)
		if !ok {
			return errors.New(fmt.Sprint("failed to scan int array value:", value))
		}
		str = stringsi.FromBytes(data)
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

func (d *StringArray) Scan(value any) error {
	str, ok := value.(string)
	if !ok {
		data, ok := value.([]byte)
		if !ok {
			return errors.New(fmt.Sprint("failed to scan string array value:", value))
		}
		str = stringsi.FromBytes(data)
	}
	strs := strings.Split(str[1:len(str)-1], ",")
	var arr []string
	for _, elem := range strs {
		arr = append(arr, elem)
	}
	*d = arr
	return nil
}

func (d StringArray) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	buf.WriteByte('"')
	for i, str := range d {
		buf.WriteString(str)
		if i != len(d)-1 {
			buf.WriteByte('"')
			buf.WriteByte(',')
		}
	}
	buf.WriteByte('"')
	buf.WriteByte('}')
	return buf.String(), nil
}

// Array represents a PostgreSQL array for T. It implements the ArrayGetter and ArraySetter interfaces. It preserves
// PostgreSQL dimensions and custom lower bounds. Use FlatArray if these are not needed.
// only support number
type Array[T any] []T

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
			subArray, ok := stringsi.BracketsIntervals(str[i:], '{', '}')
			if ok {
				i += len(subArray)
				t, err := dbi.StringConvertFor[T](subArray)
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
		t, err := dbi.StringConvertFor[T](elem)
		if err != nil {
			return err
		}
		arr = append(arr, t)
	}
	*d = arr
	return nil
}

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
			buf.WriteString(reflecti.StringFor(v))
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
			buf.WriteString(strconv.Quote(stringsi.FromBytes(v)))
			continue
		}
		buf.WriteString(reflecti.StringFor(v))
	}
	buf.WriteByte('}')
	return buf.String(), nil
}

type TimeArray []time.Time

func (d *TimeArray) Scan(value any) error {
	str, ok := value.(string)
	if !ok {
		data, ok := value.([]byte)
		if !ok {
			return errors.New(fmt.Sprint("failed to scan string array value:", value))
		}
		str = stringsi.FromBytes(data)
	}
	strs := strings.Split(str[1:len(str)-1], ",")
	var arr []time.Time
	for _, elem := range strs {
		t, err := time.Parse(time.RFC3339Nano, stringsi.Unquote(elem))
		if err != nil {
			return err
		}
		arr = append(arr, t)
	}
	*d = arr
	return nil
}

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

type JsonArray []map[string]any

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 Json
func (j *JsonArray) Scan(value interface{}) error {
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
		err = json.Unmarshal(stringsi.ToBytes(jsonStr), &m)
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

// 实现 driver.Valuer 接口，Value 返回 json value
func (j JsonArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, v := range j {
		if i > 0 {
			buf.WriteByte(',')
		}
		data, err := json.Marshal(&v)
		if err != nil {
			return nil, err
		}
		_, err = buf.WriteString(strconv.Quote(stringsi.FromBytes(data)))
		if err != nil {
			return nil, err
		}
	}

	buf.WriteByte('}')
	return buf.String(), nil
}

func (*JsonArray) GormDataType() string {
	return "jsonb[]"
}
