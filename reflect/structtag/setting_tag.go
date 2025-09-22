/*
 * Copyright 2024 hopeio. All rights reserved.
 * Licensed under the MIT License that can be found in the LICENSE file.
 * @Created by jyb
 */

package structtag

import (
	"github.com/hopeio/gox/reflect/mtos"
	"reflect"
	"strings"
)

type SettingTag string

// ParseSettingTag 适用于子tag为多配置项
/*
type example struct {
	db  string `specifyTagName:"config:db;default:postgres`
}
*/
// default sep ;
func ParseSettingTagToMap(tag string, sep byte) map[string]string {
	if tag == "" || tag == "-" {
		return nil
	}
	if sep == 0 {
		sep = ';'
	}
	sepStr := string(sep)
	settings := map[string]string{}
	names := strings.Split(tag, sepStr)

	for i := 0; i < len(names); i++ {
		j := i
		if len(names[j]) > 0 {
			for {
				if names[j][len(names[j])-1] == '\\' {
					i++
					names[j] = names[j][0:len(names[j])-1] + sepStr + names[i]
					names[i] = ""
				} else {
					break
				}
			}
		}

		values := strings.Split(names[j], ":")
		k := strings.TrimSpace(strings.ToUpper(values[0]))

		if len(values) >= 2 {
			settings[k] = strings.Join(values[1:], ":")
		} else if k != "" {
			settings[k] = "true"
		}
	}

	return settings
}

func ParseSettingTagToStruct[T any](tag string, sep byte) (*T, error) {
	if tag == "-" {
		return nil, nil
	}
	settings := new(T)
	err := ParseSettingTagIntoStruct(tag, sep, settings)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

const metaTag = "meta"

// ParseSettingTagIntoStruct 解析tag中的子tag, meta标识tag中都有哪些字段
// ParseTagSettingInto default sep ;
/*
type tag struct {
	ConfigName   string `meta:"config"`
	DefaultValue string `meta:"default"`
}
type example struct {
	db  string `specifyTagName:"config:db;default:postgres`
}
var tag tag
ParseSettingTagIntoStruct("tagName",';',&tag)
*/

func ParseSettingTagIntoStruct(tag string, sep byte, settings any) error {
	if tag == "-" {
		return ErrTagIgnore
	}
	tagSettings := ParseSettingTagToMap(tag, sep)
	if tagSettings == nil {
		return ErrTagNotExist
	}
	settingsValue := reflect.ValueOf(settings).Elem()
	settingsType := reflect.TypeOf(settings).Elem()
	for i := 0; i < settingsValue.NumField(); i++ {
		structField := settingsType.Field(i)
		var name string
		if metatag, ok := structField.Tag.Lookup(metaTag); ok {
			name = metatag
		} else {
			name = structField.Name
		}
		if flagtag, ok := tagSettings[strings.ToUpper(name)]; ok {
			err := mtos.SetValueByString(settingsValue.Field(i), flagtag)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
