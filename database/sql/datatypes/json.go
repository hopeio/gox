/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package datatypes

import (
	"database/sql/driver"

	"errors"
	"fmt"

	jsonx "github.com/hopeio/gox/encoding/json"
	"github.com/hopeio/gox/strings"
)

type RawJson []byte

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 RawJson
func (j *RawJson) Scan(value interface{}) error {
	switch bytes := value.(type) {
	case []byte:
		*j = bytes
		return nil
	case string:
		*j = strings.ToBytes(bytes)
		return nil
	default:
		return errors.New(fmt.Sprint("failed to scan RawJson value:", value))
	}

}

// 实现 driver.Valuer 接口，Value 返回 json value
func (j RawJson) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return j, nil
}

func (*RawJson) GormDataType() string {
	return "jsonb"
}

type NullJson[T any] struct {
	V     T
	Valid bool
}

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 Json
func (j *NullJson[T]) Scan(value interface{}) error {
	j.Valid = true
	switch bytes := value.(type) {
	case []byte:
		return jsonx.Unmarshal(bytes, &j.V)
	case string:
		return jsonx.Unmarshal(strings.ToBytes(bytes), &j.V)
	default:
		return errors.New(fmt.Sprint("failed to scan NullJson value:", value))
	}
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (j *NullJson[T]) Value() (driver.Value, error) {
	if !j.Valid {
		return nil, nil
	}
	return jsonx.Marshal(&j.V)
}

func (*NullJson[T]) GormDataType() string {
	return "jsonb"
}

type Json[T any] struct {
	V T
}

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 Json
func (j *Json[T]) Scan(value interface{}) error {
	switch bytes := value.(type) {
	case []byte:
		return jsonx.Unmarshal(bytes, &j.V)
	case string:
		return jsonx.Unmarshal(strings.ToBytes(bytes), &j.V)
	default:
		return errors.New(fmt.Sprint("failed to scan Json value:", value))
	}
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (j *Json[T]) Value() (driver.Value, error) {
	return jsonx.Marshal(&j.V)
}

type MapJson[T any] map[string]T

// 实现 sql.Scanner 接口，Scan 将 value 扫描至 Json
func (j *MapJson[T]) Scan(value interface{}) error {
	switch bytes := value.(type) {
	case []byte:
		return jsonx.Unmarshal(bytes, j)
	case string:
		return jsonx.Unmarshal(strings.ToBytes(bytes), &j)
	default:
		return errors.New(fmt.Sprint("failed to scan MapJson value:", value))
	}
}

// 实现 driver.Valuer 接口，Value 返回 json value
func (j MapJson[T]) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return jsonx.Marshal(j)
}
