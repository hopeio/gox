/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gorm

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"reflect"
	"time"
	"unsafe"

	sqlx "github.com/hopeio/gox/database/sql"
	jsonx "github.com/hopeio/gox/encoding/json"
	reflectx "github.com/hopeio/gox/reflect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"gorm.io/gorm/schema"
)

// init initializes package state.
func init() {
	schema.RegisterSerializer("json", JSONSerializer{})
	schema.RegisterSerializer("protojson", ProtoJSONSerializer{})
	schema.RegisterSerializer("string_array", StringArraySerializer{})
	schema.RegisterSerializer("unix_milli_time", UnixMilliTimeSerializer{})
	schema.RegisterSerializer("date", DateSerializer{})
}

// JSONSerializer is a JSON serializer
type JSONSerializer struct {
}

// Scan performs the operation.
func (JSONSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	fieldValue := reflect.New(field.FieldType)

	if dbValue != nil {
		var bytes []byte
		switch v := dbValue.(type) {
		case []byte:
			bytes = v
		case string:
			bytes = []byte(v)
		default:
			return fmt.Errorf("failed to unmarshal JSONB value: %#v", dbValue)
		}

		err = jsonx.Unmarshal(bytes, fieldValue.Interface())
	}

	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return
}

// Value returns the value.
func (JSONSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	return jsonx.Marshal(fieldValue)
}

// ProtoJSONSerializer stores proto.Message fields (e.g. google.protobuf.Struct) as JSONB
// using protojson so the wire shape matches JSON object / array semantics.
type ProtoJSONSerializer struct{}

// Scan performs the operation.
func (ProtoJSONSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	if dbValue == nil {
		return nil
	}
	var bytes []byte
	switch v := dbValue.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal protojson JSONB value: %#v", dbValue)
	}
	if len(bytes) == 0 || string(bytes) == "null" {
		return nil
	}

	rv := field.ReflectValueOf(ctx, dst)
	var msg proto.Message
	switch {
	case field.FieldType.Kind() == reflect.Ptr:
		if rv.IsNil() {
			rv.Set(reflect.New(field.FieldType.Elem()))
		}
		msg, _ = rv.Interface().(proto.Message)
	case rv.CanAddr():
		msg, _ = rv.Addr().Interface().(proto.Message)
	}
	if msg == nil {
		return fmt.Errorf("protojson serializer requires proto.Message field, got %s", field.FieldType)
	}
	return protojson.Unmarshal(bytes, msg)
}

// Value returns the value.
func (ProtoJSONSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	if fieldValue == nil {
		return nil, nil
	}
	rv := reflect.ValueOf(fieldValue)
	if !rv.IsValid() || (rv.Kind() == reflect.Ptr && rv.IsNil()) {
		return nil, nil
	}
	msg, ok := fieldValue.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("protojson serializer requires proto.Message value, got %T", fieldValue)
	}
	if proto.Size(msg) == 0 {
		return nil, nil
	}
	return protojson.Marshal(msg)
}

// StringArraySerializer is an array serializer
type StringArraySerializer struct {
}

// Scan returns the result.
func (StringArraySerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value,
	dbValue any) (err error) {
	if dbValue != nil {
		var arr sqlx.StringArray
		err = arr.Scan(dbValue)
		if err != nil {
			return err
		}
		fieldValue := reflect.ValueOf(arr)
		field.ReflectValueOf(ctx, dst).Set(fieldValue)
	}
	return
}

// Value returns the value.
func (StringArraySerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue any) (any, error) {
	arr := (*sqlx.StringArray)(unsafe.Pointer((*reflectx.Eface)(unsafe.Pointer(&fieldValue)).Value))
	return (*arr).Value()
}

type UnixMilliTimeSerializer struct {
}

// Scan implements serializer interface
func (UnixMilliTimeSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	t := sql.NullTime{}
	if err = t.Scan(dbValue); err == nil && t.Valid {
		err = field.Set(ctx, dst, t.Time.UnixMilli())
	}

	return
}

// Value implements serializer interface
func (UnixMilliTimeSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (result interface{}, err error) {
	rv := reflect.ValueOf(fieldValue)
	switch fieldValue.(type) {
	case int, int32, int64:
		result = time.UnixMilli(rv.Int()).UTC()
	case uint, uint32, uint64:
		if uv := rv.Uint(); uv > math.MaxInt64 {
			err = fmt.Errorf("integer overflow conversion uint64(%d) -> int64", uv)
		} else {
			result = time.UnixMilli(int64(uv)).UTC() //nolint:gosec
		}
	case *int, *int32, *int64:
		if rv.IsZero() {
			return nil, nil
		}
		result = time.UnixMilli(rv.Elem().Int()).UTC()
	case *uint, *uint32, *uint64:
		if rv.IsZero() {
			return nil, nil
		}
		if uv := rv.Elem().Uint(); uv > math.MaxInt64 {
			err = fmt.Errorf("integer overflow conversion uint64(%d) -> int64", uv)
		} else {
			result = time.UnixMilli(int64(uv)).UTC() //nolint:gosec
		}
	default:
		err = fmt.Errorf("invalid field type %#v for UnixSecondSerializer, only int, uint supported", fieldValue)
	}
	return
}

type DateSerializer struct {
}

// Scan implements serializer interface
func (DateSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	t := sql.NullTime{}
	if err = t.Scan(dbValue); err == nil && t.Valid {
		err = field.Set(ctx, dst, t.Time.Format("2006-01-02"))
	}
	return
}

// Value implements serializer interface
func (DateSerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (result interface{}, err error) {
	rv := reflect.ValueOf(fieldValue)
	switch fieldValue.(type) {
	case int, int32, int64:
		result = time.Unix(rv.Int()*86400, 0).Format("2006-01-02")
	case uint, uint32, uint64:
		result = time.Unix(int64(rv.Uint())*86400, 0).Format("2006-01-02")
	case *int, *int32, *int64:
		if rv.IsZero() {
			return nil, nil
		}
		result = time.Unix(rv.Elem().Int()*86400, 0).Format("2006-01-02")
	case *uint, *uint32, *uint64:
		if rv.IsZero() {
			return nil, nil
		}
		result = time.Unix(int64(rv.Elem().Uint())*86400, 0).Format("2006-01-02")
	default:
		err = fmt.Errorf("invalid field type %#v for DateSerializer, only int, uint supported", fieldValue)
	}
	return
}
