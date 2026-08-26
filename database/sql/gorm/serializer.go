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

	sqlx "github.com/hopeio/gox/database/sql"
	jsonx "github.com/hopeio/gox/encoding/json"
	"gorm.io/gorm/schema"
)

// init initializes package state.
func init() {
	schema.RegisterSerializer("json", JSONSerializer{})
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

// StringArraySerializer is an array serializer
type StringArraySerializer struct {
}

// Scan returns the result.
func (StringArraySerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value,
	dbValue any) (err error) {
	if dbValue == nil {
		return nil
	}
	var arr sqlx.StringArray
	if err = arr.Scan(dbValue); err != nil {
		return err
	}
	field.ReflectValueOf(ctx, dst).Set(reflect.ValueOf([]string(arr)))
	return nil
}

// Value returns the value.
func (StringArraySerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue any) (any, error) {
	if fieldValue == nil {
		return "{}", nil
	}
	switch v := fieldValue.(type) {
	case []string:
		return sqlx.StringArray(v).Value()
	case *[]string:
		if v == nil {
			return nil, nil
		}
		return sqlx.StringArray(*v).Value()
	case sqlx.StringArray:
		return v.Value()
	case *sqlx.StringArray:
		if v == nil {
			return nil, nil
		}
		return v.Value()
	default:
		rv := reflect.ValueOf(fieldValue)
		if rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() == reflect.String {
			out := make(sqlx.StringArray, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				out[i] = rv.Index(i).String()
			}
			return out.Value()
		}
		return nil, fmt.Errorf("string_array: unsupported type %T", fieldValue)
	}
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
