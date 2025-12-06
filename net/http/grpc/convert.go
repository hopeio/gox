/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package grpc

import (
	"fmt"
	"strconv"
	"strings"

	strconv2 "github.com/hopeio/gox/strconv"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Timestamp converts the given RFC3339 formatted string into a timestamp.Timestamp.
func Timestamp(val string) (*timestamppb.Timestamp, error) {
	var r timestamppb.Timestamp
	val = strconv.Quote(strings.Trim(val, `"`))
	unmarshaler := &protojson.UnmarshalOptions{}
	if err := unmarshaler.Unmarshal([]byte(val), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Duration converts the given string into a timestamp.Duration.
func Duration(val string) (*durationpb.Duration, error) {
	var r durationpb.Duration
	val = strconv.Quote(strings.Trim(val, `"`))
	unmarshaler := &protojson.UnmarshalOptions{}
	if err := unmarshaler.Unmarshal([]byte(val), &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// Enum converts the given string into an int32 that should be type casted into the
// correct enum proto type.
func Enum(val string, enumValMap map[string]int32) (int32, error) {
	e, ok := enumValMap[val]
	if ok {
		return e, nil
	}

	i, err := strconv2.Signed[int32](val)
	if err != nil {
		return 0, fmt.Errorf("%s is not valid", val)
	}
	for _, v := range enumValMap {
		if v == i {
			return i, nil
		}
	}
	return 0, fmt.Errorf("%s is not valid", val)
}

// EnumSlice converts 'val' where individual enums are separated by 'sep'
// into a int32 slice. Each individual int32 should be type casted into the
// correct enum proto type.
func EnumSlice(val, sep string, enumValMap map[string]int32) ([]int32, error) {
	s := strings.Split(val, sep)
	values := make([]int32, len(s))
	for i, v := range s {
		value, err := Enum(v, enumValMap)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}
	return values, nil
}

// Support for google.protobuf.wrappers on top of primitive types

// StringValue well-known type support as wrapper around string type
func StringValue(val string) (*wrapperspb.StringValue, error) {
	return wrapperspb.String(val), nil
}

// FloatValue well-known type support as wrapper around float32 type
func FloatValue(val string) (*wrapperspb.FloatValue, error) {
	parsedVal, err := strconv2.Float32(val)
	return wrapperspb.Float(parsedVal), err
}

// DoubleValue well-known type support as wrapper around float64 type
func DoubleValue(val string) (*wrapperspb.DoubleValue, error) {
	parsedVal, err := strconv2.Float64(val)
	return wrapperspb.Double(parsedVal), err
}

// BoolValue well-known type support as wrapper around bool type
func BoolValue(val string) (*wrapperspb.BoolValue, error) {
	parsedVal, err := strconv2.Bool(val)
	return wrapperspb.Bool(parsedVal), err
}

// Int32Value well-known type support as wrapper around int32 type
func Int32Value(val string) (*wrapperspb.Int32Value, error) {
	parsedVal, err := strconv2.Int32(val)
	return wrapperspb.Int32(parsedVal), err
}

// UInt32Value well-known type support as wrapper around uint32 type
func UInt32Value(val string) (*wrapperspb.UInt32Value, error) {
	parsedVal, err := strconv2.Uint32(val)
	return wrapperspb.UInt32(parsedVal), err
}

// Int64Value well-known type support as wrapper around int64 type
func Int64Value(val string) (*wrapperspb.Int64Value, error) {
	parsedVal, err := strconv2.Int64(val)
	return wrapperspb.Int64(parsedVal), err
}

// UInt64Value well-known type support as wrapper around uint64 type
func UInt64Value(val string) (*wrapperspb.UInt64Value, error) {
	parsedVal, err := strconv2.Uint64(val)
	return wrapperspb.UInt64(parsedVal), err
}

// BytesValue well-known type support as wrapper around bytes[] type
func BytesValue(val string) (*wrapperspb.BytesValue, error) {
	parsedVal, err := strconv2.Bytes(val)
	return wrapperspb.Bytes(parsedVal), err
}
