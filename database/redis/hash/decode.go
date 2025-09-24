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

func Unmarshal(v interface{}, strings []string) {
	uValue := reflect.ValueOf(v).Elem()
	for i := 0; i < len(strings); i += 2 {
		fieldValue := uValue.FieldByName(strings[i])
		switch fieldValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v, _ := strconv.ParseInt(strings[i+1], 10, 64)
			fieldValue.SetInt(v)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v, _ := strconv.ParseUint(strings[i+1], 10, 64)
			fieldValue.SetUint(v)
		case reflect.String:
			fieldValue.SetString(strings[i+1])
		case reflect.Float32, reflect.Float64:
			v, _ := strconv.ParseFloat(strings[i+1], 64)
			fieldValue.SetFloat(v)
		case reflect.Bool:
			v, _ := strconv.ParseBool(strings[i+1])
			fieldValue.SetBool(v)
		}
	}
}
