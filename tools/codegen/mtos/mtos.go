/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package mtos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hopeio/gox/log"
	stringsx "github.com/hopeio/gox/strings"
	"io"
	"reflect"
)

func ParseJson(data []byte) (string, error) {
	var m map[string]any
	err := json.Unmarshal(data, &m)
	if err != nil {
		return "", err
	}
	buff := bytes.Buffer{}
	Gen(&buff, "T", m, "json")
	return buff.String(), nil
}

func Gen(writer io.Writer, name string, m map[string]any, tag string) {
	writer.Write([]byte(fmt.Sprintf("type %s struct{\n", name)))
	var nexts []*Next
	for key, value := range m {
		dataValue := reflect.ValueOf(value)
		fieldCode, next := generateFieldCode(key, "", dataValue, tag)
		if fieldCode != "" {
			writer.Write([]byte("\t" + fieldCode + "\n"))
		}
		if next != nil {
			nexts = append(nexts, next...)
		}

	}
	writer.Write([]byte("}\n"))
	for _, next := range nexts {
		Gen(writer, next.Key, next.Value, tag)
	}
}

type Next struct {
	Key   string
	Value map[string]any
}

func generateFieldCode(k string, typePrefix string, fieldValue reflect.Value, tag string) (string, []*Next) {
	fieldType := fieldValue.Type()
	fieldName := stringsx.SnakeToCamel(k)
	switch fieldType.Kind() {
	case reflect.String, reflect.Int, reflect.Float64, reflect.Bool:
		return fmt.Sprintf("%s %s `%s:\"%s\"`", fieldName, typePrefix+fieldType.Kind().String(), tag, k), nil
	case reflect.Interface:
		return fmt.Sprintf("%s %s `%s:\"%s\"`", fieldName, typePrefix+"any", tag, k), nil
	case reflect.Map:
		return fmt.Sprintf("%s %s `%s:\"%s\"`", fieldName, typePrefix+fieldName, tag, k), nil
	case reflect.Slice:
		if fieldValue.Len() == 0 {
			return fmt.Sprintf("%s %s `%s:\"%s\"`", fieldName, typePrefix+"[]any", tag, k), nil
		}
		// TODO 切片元素类型相同与不同,相同（map直接比较是否同一个类型）,不相同(全部基本类型和含map,slice类型)
		elementValue := fieldValue.Index(0)
		return generateFieldCode(k, typePrefix+"[]", elementValue, tag)
	default:
		log.Printf("unsupported type for field %s: %v", fieldName, fieldType)
		return "", nil
	}
}
