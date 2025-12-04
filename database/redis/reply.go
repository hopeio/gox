// Copyright 2012 Gary Burd
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package redis

import (
	"errors"
	"fmt"

	"github.com/spf13/cast"
)

type Error string

func (err Error) Error() string { return string(err) }

// ErrNil indicates that a reply value is nil.
var ErrNil = errors.New("redis: nil")

// Bytes is a helper that converts a command reply to a slice of bytes. If err
// is not equal to nil, then Bytes returns nil, err. Otherwise Bytes converts
// the reply to a slice of bytes as follows:
//
//	Reply type      Result
//	bulk string     reply, nil
//	simple string   []byte(reply), nil
//	nil             nil, ErrNil
//	other           nil, error
func Bytes(reply interface{}, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	switch reply := reply.(type) {
	case []byte:
		return reply, nil
	case string:
		return []byte(reply), nil
	case nil:
		return nil, ErrNil
	case Error:
		return nil, reply
	}
	return nil, fmt.Errorf("redigo: unexpected type for Bytes, got type %T", reply)
}

// MultiBulk is a helper that converts an array command reply to a []interface{}.
//
// Deprecated: Use Values instead.
func MultiBulk(reply interface{}, err error) ([]interface{}, error) { return Values(reply, err) }

// Values is a helper that converts an array command reply to a []interface{}.
// If err is not equal to nil, then Values returns nil, err. Otherwise, Values
// converts the reply as follows:
//
//	Reply type      Result
//	array           reply, nil
//	nil             nil, ErrNil
//	other           nil, error
func Values(reply interface{}, err error) ([]interface{}, error) {
	if err != nil {
		return nil, err
	}
	switch reply := reply.(type) {
	case []interface{}:
		return reply, nil
	case nil:
		return nil, ErrNil
	case Error:
		return nil, reply
	}
	return nil, fmt.Errorf("redigo: unexpected type for Values, got type %T", reply)
}

func sliceHelper(reply interface{}, err error, name string, makeSlice func(int), assign func(int, interface{}) error) error {
	if err != nil {
		return err
	}
	switch reply := reply.(type) {
	case []interface{}:
		makeSlice(len(reply))
		for i := range reply {
			if reply[i] == nil {
				continue
			}
			if err := assign(i, reply[i]); err != nil {
				return err
			}
		}
		return nil
	case nil:
		return ErrNil
	case Error:
		return reply
	}
	return fmt.Errorf("redigo: unexpected type for %s, got type %T", name, reply)
}

// ByteSlices is a helper that converts an array command reply to a [][]byte.
// If err is not equal to nil, then ByteSlices returns nil, err. Nil array
// items are stay nil. ByteSlices returns an error if an array item is not a
// bulk string or nil.
func ByteSlices(reply interface{}, err error) ([][]byte, error) {
	var result [][]byte
	err = sliceHelper(reply, err, "ByteSlices", func(n int) { result = make([][]byte, n) }, func(i int, v interface{}) error {
		p, ok := v.([]byte)
		if !ok {
			return fmt.Errorf("redigo: unexpected element type for ByteSlices, got type %T", v)
		}
		result[i] = p
		return nil
	})
	return result, err
}

// StringMap is a helper that converts an array of strings (alternating key, value)
// into a map[string]string. The HGETALL and CONFIG GET commands return replies in this format.
// Requires an even number of values in result.
func StringMap(result interface{}, err error) (map[string]string, error) {
	values, err := Values(result, err)
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, errors.New("redigo: StringMap expects even number of values result")
	}
	m := make(map[string]string, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, okKey := values[i].([]byte)
		value, okValue := values[i+1].([]byte)
		if !okKey || !okValue {
			return nil, errors.New("redigo: StringMap key not a bulk string value")
		}
		m[string(key)] = string(value)
	}
	return m, nil
}

// IntMap is a helper that converts an array of strings (alternating key, value)
// into a map[string]int. The HGETALL commands return replies in this format.
// Requires an even number of values in result.
func IntMap(result interface{}, err error) (map[string]int, error) {
	values, err := Values(result, err)
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, errors.New("redigo: IntMap expects even number of values result")
	}
	m := make(map[string]int, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].([]byte)
		if !ok {
			return nil, errors.New("redigo: IntMap key not a bulk string value")
		}
		value, err := cast.ToIntE(values[i+1])
		if err != nil {
			return nil, err
		}
		m[string(key)] = value
	}
	return m, nil
}

// Int64Map is a helper that converts an array of strings (alternating key, value)
// into a map[string]int64. The HGETALL commands return replies in this format.
// Requires an even number of values in result.
func Int64Map(result interface{}, err error) (map[string]int64, error) {
	values, err := Values(result, err)
	if err != nil {
		return nil, err
	}
	if len(values)%2 != 0 {
		return nil, errors.New("redigo: Int64Map expects even number of values result")
	}
	m := make(map[string]int64, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].([]byte)
		if !ok {
			return nil, errors.New("redigo: Int64Map key not a bulk string value")
		}
		value, err := cast.ToInt64E(values[i+1])
		if err != nil {
			return nil, err
		}
		m[string(key)] = value
	}
	return m, nil
}

// Positions is a helper that converts an array of positions (lat, long)
// into a [][2]float64. The GEOPOS command returns replies in this format.
func Positions(result interface{}, err error) ([]*[2]float64, error) {
	values, err := Values(result, err)
	if err != nil {
		return nil, err
	}
	positions := make([]*[2]float64, len(values))
	for i := range values {
		if values[i] == nil {
			continue
		}
		p, ok := values[i].([]interface{})
		if !ok {
			return nil, fmt.Errorf("redigo: unexpected element type for interface slice, got type %T", values[i])
		}
		if len(p) != 2 {
			return nil, fmt.Errorf("redigo: unexpected number of values for a member position, got %d", len(p))
		}
		lat, err := cast.ToFloat64E(p[0])
		if err != nil {
			return nil, err
		}
		long, err := cast.ToFloat64E(p[1])
		if err != nil {
			return nil, err
		}
		positions[i] = &[2]float64{lat, long}
	}
	return positions, nil
}
