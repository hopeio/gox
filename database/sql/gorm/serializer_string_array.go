/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package gorm

import (
	"context"
	"reflect"
	"unsafe"

	sqlx "github.com/hopeio/gox/database/sql"
	reflectx "github.com/hopeio/gox/reflect"
	"gorm.io/gorm/schema"
)

// StringArraySerializer array序列化器
type StringArraySerializer struct {
}

// 实现 Scan 方法
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

// 实现 Value 方法
func (StringArraySerializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue any) (any, error) {
	arr := (*sqlx.StringArray)(unsafe.Pointer((*reflectx.Eface)(unsafe.Pointer(&fieldValue)).Value))
	return (*arr).Value()
}
