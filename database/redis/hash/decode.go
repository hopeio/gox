/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package hash

import (
	"reflect"
	"strconv"
)

func Unmarshal(v interface{}, args map[string]string) {
	uValue := reflect.ValueOf(v).Elem()
	uType := uValue.Type()
	for i := 0; i < uValue.NumField(); i++ {
		fieldValue := uValue.Field(i)
		field := uType.Field(i)
		switch fieldValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v, _ := strconv.ParseInt(args[field.Name], 10, 64)
			fieldValue.SetInt(v)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v, _ := strconv.ParseUint(args[field.Name], 10, 64)
			fieldValue.SetUint(v)
		case reflect.String:
			fieldValue.SetString(args[field.Name])
		case reflect.Float32, reflect.Float64:
			v, _ := strconv.ParseFloat(args[field.Name], 64)
			fieldValue.SetFloat(v)
		case reflect.Bool:
			v, _ := strconv.ParseBool(args[field.Name])
			fieldValue.SetBool(v)
		}
	}
}
